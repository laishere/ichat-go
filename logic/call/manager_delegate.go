package call

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/logging"
	"ichat-go/model/dao"
	"ichat-go/model/entity"
	"ichat-go/sched"
	"strconv"
	"time"
)

const userTTL = time.Second * 30
const lockTTL = userTTL * 2

type ManagerDelegate interface {
	CallId() uint64
	CallerId() uint64
	UpdateUserCallLock(userId uint64, lock bool) error
	IsUserLockValid(userId uint64) bool
	UserIds() []uint64
	UserStates() []UserState
	UserState(userId uint64) UserState
	SaveUserState(state UserState)
	ManagerLock() error
	ManagerUnlock()
	CallStatus() int
	UserSession(userId uint64) Session
	UpdateUserTTL(userId uint64)
	AliveUserCount(online bool) int
	CloseUserSession(userId uint64)
	CallReady() error
	CallStart()
	CallEnd(reason int)
	DeadUsers() <-chan uint64
	Close()
}

type delegate struct {
	callId         uint64
	call           *entity.Call
	c              *redis.Client
	ctx            context.Context
	cancel         context.CancelFunc
	logger         logging.Logger
	callDao        dao.CallDao
	dq             sched.DQ
	deadUsers      chan uint64
	updateCallback NotifyCallUpdateCallback
}

type NotifyCallUpdateCallback = func(messageId uint64)

var callback NotifyCallUpdateCallback

func SetNotifyCallUpdateCallback(cb NotifyCallUpdateCallback) {
	callback = cb
}

func NewManagerDelegate(callId uint64) ManagerDelegate {
	c := di.ENV().RDB()
	ctx, cancel := context.WithCancel(context.Background())
	logger := logging.NewLogger("call:" + strconv.FormatUint(callId, 10))
	dq := sched.NewDQ("call:userTTL:" + strconv.FormatUint(callId, 10))
	callDao := di.ENV().CallDao()
	call := callDao.FindCallById(callId)
	return &delegate{callId: callId, call: call, c: c,
		ctx: ctx, cancel: cancel, logger: logger, dq: dq, callDao: callDao, updateCallback: callback}
}

func (d *delegate) CallId() uint64 {
	return d.callId
}

func (d *delegate) CallerId() uint64 {
	return d.call.CallerId
}

func (d *delegate) AfterSetup() {
	// 切换为非事务模式
	d.callDao = di.ENV().CallDao()
}

func (d *delegate) userCallLockKey(userId uint64) string {
	return "call:userLock:" + strconv.FormatUint(userId, 10)
}

func (d *delegate) userIdsKey() string {
	return "call:userIds:" + strconv.FormatUint(d.callId, 10)
}

func (d *delegate) userStatesMapKey() string {
	return "call:userStates:" + strconv.FormatUint(d.callId, 10)
}

func (d *delegate) managerLockKey() string {
	return "call:managerLock:" + strconv.FormatUint(d.callId, 10)
}

func (d *delegate) callStatusKey() string {
	return "call:status:" + strconv.FormatUint(d.callId, 10)
}

func (d *delegate) UpdateUserCallLock(userId uint64, lock bool) error {
	key := d.userCallLockKey(userId)
	if lock {
		script := `
			if redis.call("exists", KEYS[1]) == 0 then
				redis.call("set", KEYS[1], ARGV[1], "PX", ARGV[2])
				return 0
			elseif redis.call("get", KEYS[1]) == ARGV[1] then
				redis.call("pexpire", KEYS[1], ARGV[2])
				return 0
			else
				return 1
			end
			`
		r, err := d.c.Eval(d.ctx, script, []string{key}, []string{strconv.FormatUint(d.callId, 10),
			strconv.FormatInt(lockTTL.Milliseconds(), 10)}).Result()
		if err != nil {
			d.logger.Error("Failed to lock user", err)
			return err
		}
		if r.(int64) == 1 {
			return errs.CallUserLockInvalid
		}
		return nil
	}
	script := `
		if redis.call("exists", KEYS[1]) == 0 then
			return 0
		elseif redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 1
		end
		`
	r, err := d.c.Eval(d.ctx, script, []string{key}, []string{strconv.FormatUint(d.callId, 10)}).Result()
	if err != nil {
		d.logger.Error("Failed to unlock user", err)
		return err
	}
	if r.(int64) == 1 {
		return errs.CallUserLockInvalid
	}
	return nil
}

func (d *delegate) IsUserLockValid(userId uint64) bool {
	key := d.userCallLockKey(userId)
	r, err := d.c.Get(d.ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			d.logger.Error("Failed to check user lock", err)
		}
		return false
	}
	return r == strconv.FormatUint(d.callId, 10)
}

func (d *delegate) UserIds() []uint64 {
	r, err := d.c.Get(d.ctx, d.userIdsKey()).Result()
	var userIds []uint64
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			d.logger.Error("Failed to get user ids", err)
		}
		userIds = d.callDao.GetUserIds(d.callId)
		userIdsJson, _ := json.Marshal(userIds)
		d.c.Set(d.c.Context(), d.userIdsKey(), userIdsJson, 0)
	} else {
		_ = json.Unmarshal([]byte(r), &userIds)
	}
	return userIds
}

func (d *delegate) UserStates() []UserState {
	r, err := d.c.HGetAll(d.ctx, d.userStatesMapKey()).Result()
	var userStates []UserState
	if err != nil {
		d.logger.Error("Failed to get user states", err)
		return nil
	} else {
		for _, v := range r {
			var state UserState
			_ = json.Unmarshal([]byte(v), &state)
			userStates = append(userStates, state)
		}
	}
	return userStates
}

func (d *delegate) UserState(userId uint64) UserState {
	r, err := d.c.HGet(d.ctx, d.userStatesMapKey(), strconv.FormatUint(userId, 10)).Result()
	if err != nil {
		d.logger.Error("Failed to get user state", err)
		return UserState{UserId: userIdInvalid}
	}
	var state UserState
	_ = json.Unmarshal([]byte(r), &state)
	return state
}

func (d *delegate) SaveUserState(state UserState) {
	if state.UserId == userIdInvalid {
		d.logger.Error("Invalid user id")
		return
	}
	key := d.userStatesMapKey()
	field := strconv.FormatUint(state.UserId, 10)
	value, _ := json.Marshal(state)
	_, err := d.c.HSet(d.ctx, key, field, value).Result()
	if err != nil {
		d.logger.Error("Failed to save user state", err)
	}
}

func (d *delegate) ManagerLock() error {
	r, err := d.c.SetNX(d.ctx, d.managerLockKey(), "1", 0).Result()
	if err != nil {
		d.logger.Error("Failed to lock manager", err)
		return err
	}
	if !r {
		return errs.CallManagerLocked
	}
	return nil
}

func (d *delegate) ManagerUnlock() {
	_, err := d.c.Del(d.ctx, d.managerLockKey()).Result()
	if err != nil {
		d.logger.Error("Failed to unlock manager", err)
	}
}

func (d *delegate) CallStatus() int {
	r, err := d.c.Get(d.ctx, d.callStatusKey()).Result()
	s := 0
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			d.logger.Error("Failed to get call status", err)
		}
		s = d.callDao.GetCallStatus(d.callId)
		_ = d.updateCallStatusCache(s)
	} else {
		s, _ = strconv.Atoi(r)
	}
	return s
}

func (d *delegate) updateCallStatus(s int) error {
	err := d.callDao.UpdateCallStatus(d.callId, s)
	if err == nil {
		err = d.updateCallStatusCache(s)
	}
	if err != nil {
		d.logger.Error("Failed to update call status", err)
	}
	return err
}

func (d *delegate) updateCallStatusCache(s int) error {
	_, err := d.c.Set(d.ctx, d.callStatusKey(), strconv.Itoa(s), 0).Result()
	d.notifyCallStatusChanged()
	return err
}

func (d *delegate) CallReady() error {
	return d.updateCallStatus(entity.CallStatusReady)
}

func (d *delegate) UserSession(userId uint64) Session {
	return findSession(d.callId, userId)
}

func (d *delegate) UpdateUserTTL(userId uint64) {
	//d.logger.Debug("Update user ttl ", userId)
	if err := d.UpdateUserCallLock(userId, true); err != nil {
		d.logger.Error("Failed to update user lock", err)
	}
	mid := strconv.FormatUint(userId, 10)
	d.dq.Delete(mid)
	if err := d.dq.Delay(userTTL, sched.Message{Id: mid}); err != nil {
		d.logger.Error("Failed to update user ttl", err)
	}
}

func (d *delegate) DeadUsers() <-chan uint64 {
	if d.deadUsers != nil {
		return d.deadUsers
	}
	ch := make(chan uint64)
	d.deadUsers = ch
	go func() {
		defer close(ch)
		for {
			select {
			case <-d.ctx.Done():
				return
			case m := <-d.dq.Channel():
				id, _ := strconv.ParseUint(m.Id, 10, 64)
				ch <- id
			}
		}
	}()
	return ch
}

func (d *delegate) AliveUserCount(online bool) int {
	// todo 动态维护数量而不是每次都计算
	userStates := d.UserStates()
	count := 0
	for _, state := range userStates {
		if online {
			if state.State == userStateOnline {
				count++
			}
		} else if state.State != userStateDead && state.State != userStateRejected {
			count++
		}
	}
	return count
}

func (d *delegate) CloseUserSession(userId uint64) {
	s := findSession(d.callId, userId)
	if s != nil {
		s.Close()
	}
}

func (d *delegate) CallStart() {
	d.callDao.UpdateStartTime(d.callId)
	_ = d.updateCallStatusCache(entity.CallStatusActive)
}

func (d *delegate) CallEnd(reason int) {
	d.callDao.UpdateEndReasonAndTime(d.callId, reason)
	_ = d.updateCallStatusCache(entity.CallStatusEnd)
}

func (d *delegate) notifyCallStatusChanged() {
	callback(d.call.MessageId)
}

func (d *delegate) Close() {
	d.cleanUp()
}

func (d *delegate) cleanUp() {
	d.logger.Debug("clean up")
	d.cancel()
	d.dq.Close(true)
	d.c.Del(context.Background(), d.managerLockKey(), d.userIdsKey(), d.userStatesMapKey(), d.callStatusKey())
}
