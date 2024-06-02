package call

import (
	"context"
	"errors"
	"ichat-go/errs"
	"ichat-go/logging"
	"ichat-go/model/entity"
	"ichat-go/sched"
	"strconv"
	"time"
)

func managerKey(callId uint64) string {
	return "call:manager:" + strconv.FormatUint(callId, 10)
}

type Manager interface {
	Loop()
	CleanAfterDied()
	ManagerApi
}

type manager struct {
	mq               sched.MQ
	ctx              context.Context
	cancel           context.CancelFunc
	delegate         ManagerDelegate
	logger           logging.Logger
	callFailedReason int
	callFailedTimer  *time.Timer
}

func (m *manager) checkCall() bool {
	if err := m.delegate.ManagerLock(); err != nil {
		m.logger.Error("Can't get call lock", err)
		return false
	}
	if m.delegate.CallStatus() != entity.CallStatusNew {
		m.logger.Warn("Call status is not new")
		return false
	}
	return true
}

func (m *manager) handleSetupError(err error) {
	if err != nil {
		if errors.Is(err, errs.CallerBusy) || errors.Is(err, errs.CalleeBusy) {
			m.callEnd(entity.CallEndReasonBusy)
		} else {
			m.callEnd(entity.CallEndReasonError)
		}
	}
}

func (m *manager) setup() bool {
	m.logger.Debug("Setup begins")
	if err := m.tryLockUsers(); err != nil {
		m.logger.Warn("Can't lock users:", err)
		m.handleSetupError(err)
		return false
	}
	m.initUserStates()
	m.mq = sched.NewMQ(managerKey(m.delegate.CallId()))
	m.mq.SaveState(1)
	if err := m.delegate.CallReady(); err != nil {
		m.handleSetupError(err)
		return false
	}
	m.logger.Debug("Setup success!")
	return true
}

func (m *manager) tryLockUsers() error {
	userIds := m.delegate.UserIds()
	lockedUserIds := make([]uint64, 0)
	isCallerLocked := false
	for _, userId := range userIds {
		if err := m.delegate.UpdateUserCallLock(userId, true); err == nil {
			lockedUserIds = append(lockedUserIds, userId)
		} else if userId == m.delegate.CallerId() {
			isCallerLocked = true
			break
		}
	}
	m.logger.Debugf("users count: %d, locked: %d", len(userIds), len(lockedUserIds))
	if isCallerLocked {
		return errs.CallerBusy
	}
	if len(lockedUserIds) < 2 {
		return errs.CalleeBusy
	}
	return nil
}

func (m *manager) initUserStates() {
	m.logger.Debug("Init user states")
	userIds := m.delegate.UserIds()
	for _, userId := range userIds {
		state := UserState{UserId: userId, State: userStateInvited, Ping: userPingNone}
		m.delegate.SaveUserState(state)
	}
}

func _canTransferUserState(from, to int) bool {
	if to == userStateDead {
		return from == userStateOffline || from == userStateOnline
	}
	if to == userStateOnline {
		return from == userStateAccepted || from == userStateOffline || from == userStateDead
	}
	if to == userStateOffline {
		return from == userStateOnline
	}
	if to == userStateAccepted || to == userStateRejected {
		return from == userStateInvited
	}
	return false
}

func (m *manager) canTransferUserState(from, to int) bool {
	valid := _canTransferUserState(from, to)
	if !valid {
		m.logger.Error("Invalid state transfer ", from, to)
	}
	return valid
}

func (m *manager) forEachSession(excludeUserId uint64, callback func(Session)) {
	for _, userId := range m.delegate.UserIds() {
		if userId == excludeUserId {
			continue
		}
		if s := m.delegate.UserSession(userId); s != nil {
			callback(s)
		}
	}
}

func (m *manager) notifyUserStateUpdated(state UserState) {
	if m.delegate.CallStatus() == entity.CallStatusEnd {
		return
	}
	m.forEachSession(state.UserId, func(s Session) {
		s.UpdateUserState(state)
	})
}

func (m *manager) UserJoined(userId uint64) {
	m.logger.Debugf("User %d joined", userId)
	state := m.delegate.UserState(userId)
	if state.State == userStateInvited && m.canTransferUserState(state.State, userStateAccepted) {
		state.State = userStateAccepted
		m.delegate.SaveUserState(state)
		m.delegate.UpdateUserTTL(userId)
		m.notifyUserStateUpdated(state)
	}
}

func (m *manager) UserOnline(userId uint64) {
	m.logger.Debugf("User %d online", userId)
	state := m.delegate.UserState(userId)
	if m.canTransferUserState(state.State, userStateOnline) {
		if state.State == userStateDead {
			// 已经退出的用户重新进入
			err := m.delegate.UpdateUserCallLock(userId, true)
			if err != nil {
				m.onUserExit(userId, userExitReasonReenterBusy)
				return
			}
		}
		state.State = userStateOnline
		m.delegate.SaveUserState(state)
		m.delegate.UpdateUserTTL(userId)
		if m.delegate.CallStatus() == entity.CallStatusReady {
			m.checkIsCallStarted(userId)
		}
		if s := m.delegate.UserSession(userId); s != nil {
			s.UpdateUserStates(m.delegate.UserStates())
		}
		m.notifyUserStateUpdated(state)
	}
}

func (m *manager) notifyCallStarted() {
	m.forEachSession(0, func(s Session) {
		s.CallStart()
	})
}

func (m *manager) checkIsCallStarted(onlineUserId uint64) {
	if m.delegate.CallStatus() != entity.CallStatusReady {
		m.logger.Error("Call status is not ready")
		return
	}
	if m.delegate.AliveUserCount(true) >= 2 {
		m.delegate.CallStart()
		m.stopCallFailedTimer()
		m.notifyCallStarted()
	} else if onlineUserId == m.delegate.CallerId() {
		m.callFailedReason = entity.CallEndReasonNoAnswer
	}
}

func (m *manager) UserOffline(userId uint64) {
	m.logger.Debugf("User %d offline", userId)
	state := m.delegate.UserState(userId)
	if state.State != userStateDead && m.canTransferUserState(state.State, userStateOffline) {
		state.State = userStateOffline
		m.delegate.SaveUserState(state)
		m.notifyUserStateUpdated(state)
	}
}

func (m *manager) userDead(userId uint64, reason int) {
	m.logger.Debugf("User %d dead, reason: %d", userId, reason)
	state := m.delegate.UserState(userId)
	if m.canTransferUserState(state.State, userStateDead) {
		state.State = userStateDead
		m.delegate.SaveUserState(state)
		m.notifyUserStateUpdated(state)
		m.onUserExit(userId, reason)
	}
}

func (m *manager) onUserExit(userId uint64, reason int) {
	endReason := 0
	cancelled := false
	switch reason {
	case userExitReasonNormal:
		if m.delegate.CallStatus() == entity.CallStatusReady && m.delegate.CallerId() == userId {
			cancelled = true
			endReason = entity.CallEndReasonCancelled
		} else {
			endReason = entity.CallEndReasonNormal
		}
	case userExitReasonLost:
		endReason = entity.CallEndReasonLostConnection
	case userExitReasonRejected:
		endReason = entity.CallEndReasonRejected
	case userExitReasonReenterBusy:
		endReason = entity.CallEndReasonError
	default:
		endReason = entity.CallEndReasonNormal
		m.logger.Errorf("Unknown user exit reason: %d", reason)
	}
	m.cleanUpUser(userId, endReason)
	if cancelled || m.delegate.AliveUserCount(false) < 2 {
		m.callEnd(endReason)
	}
}

func (m *manager) cleanUpUser(userId uint64, reason int) {
	m.logger.Debugf("Clean up user: %d", userId)
	if s := m.delegate.UserSession(userId); s != nil {
		s.CallEnd(reason)
	}
	m.delegate.CloseUserSession(userId)
	_ = m.delegate.UpdateUserCallLock(userId, false)
}

func (m *manager) Hangup(userId uint64) {
	m.logger.Debugf("User %d hangup", userId)
	state := m.delegate.UserState(userId)
	if state.UserId == userIdInvalid {
		m.logger.Error("User state not found")
		return
	}
	if state.State == userStateInvited {
		state.State = userStateRejected
		m.delegate.SaveUserState(state)
		m.onUserExit(userId, userExitReasonRejected)
		m.notifyUserStateUpdated(state)
	} else {
		m.userDead(userId, userExitReasonNormal)
	}
}

func (m *manager) Signaling(fromUserId, toUserId uint64, message string) {
	m.logger.Debugf("Signaling from %d to %d", fromUserId, toUserId)
	if s := m.delegate.UserSession(toUserId); s != nil {
		s.Signaling(fromUserId, message)
	}
}

func (m *manager) HeartBeat(userId uint64) {
	//m.logger.Debugf("User %d heartbeat", userId)
	m.delegate.UpdateUserTTL(userId)
}

func (m *manager) callEnd(reason int) {
	callStatus := m.delegate.CallStatus()
	m.logger.Debugf("Call end, reason: %d, status: %d", reason, callStatus)
	if callStatus == entity.CallStatusEnd {
		m.logger.Warn("Call already ended")
		return
	}
	m.delegate.CallEnd(reason)
	m.cleanUp(reason)
	m.cancel()
}

func (m *manager) cleanUp(reason int) {
	m.logger.Debug("Clean up")
	userIds := m.delegate.UserIds()
	for _, userId := range userIds {
		m.cleanUpUser(userId, reason)
	}
}

func (m *manager) exit() {
	if m.delegate.CallStatus() != entity.CallStatusEnd {
		m.logger.Error("exit before call end")
		m.callEnd(entity.CallEndReasonError)
	}
	m.stopCallFailedTimer()
	m.cancel()
	if m.mq != nil {
		m.mq.Close(true)
	}
	m.delegate.ManagerUnlock()
	clearManagerHeartbeat(m.delegate.CallId())
	m.delegate.Close()
	m.logger.Debug("exit")
}

func (m *manager) startCallFailedTimer() {
	m.callFailedReason = entity.CallEndReasonError
	m.callFailedTimer = time.NewTimer(time.Second * 60)
}

func (m *manager) stopCallFailedTimer() {
	if m.callFailedTimer != nil {
		m.callFailedTimer.Stop()
	}
}

func (m *manager) Loop() {
	m.logger.Debug("enter loop")
	defer func() {
		if err := recover(); err != nil {
			m.logger.Error("Loop panic", err)
		}
		m.exit()
	}()
	if !m.checkCall() {
		return
	}
	if !m.setup() {
		return
	}
	managerHeartbeat(m.delegate.CallId())
	hbTick := time.NewTicker(managerTTL / 2)
	defer hbTick.Stop()
	m.startCallFailedTimer()
	for {
		select {
		case <-m.callFailedTimer.C:
			m.logger.Error("Call failed")
			m.callEnd(m.callFailedReason)
			return
		case msg := <-m.mq.Channel():
			m.handleAction(&msg)
			m.mq.Ack(true)
		case userId := <-m.delegate.DeadUsers():
			m.userDead(userId, userExitReasonLost)
		case <-m.ctx.Done():
			m.logger.Debug("loop exit")
			return
		case <-hbTick.C:
			managerHeartbeat(m.delegate.CallId())
		}
	}
}

func (m *manager) handleAction(msg *sched.Message) {
	//m.logger.Debugf("action %d", msg.Type)
	switch msg.Type {
	case actionTypeUserAccepted:
		m.UserJoined(userAcceptedAction(msg))
	case actionTypeUserOnline:
		m.UserOnline(userOnlineAction(msg))
	case actionTypeUserOffline:
		m.UserOffline(userOfflineAction(msg))
	case actionTypeHangup:
		m.Hangup(hangupAction(msg))
	case actionTypeSignaling:
		a := signalingAction(msg)
		m.Signaling(a.FromUserId, a.ToUserId, a.Message)
	case actionTypeHeartBeat:
		m.HeartBeat(heartBeatAction(msg))
	default:
		m.logger.Errorf("Unknown action type: %d", msg.Type)
	}
}

func (m *manager) CleanAfterDied() {
	m.logger.Debug("cleaning call")
	if m.delegate.CallStatus() == entity.CallStatusEnd {
		m.logger.Error("call already ended")
		return
	}
	m.callEnd(entity.CallEndReasonError)
	m.delegate.ManagerUnlock()
	m.delegate.Close()
	m.logger.Debug("clean done")
}

func NewManager(delegate ManagerDelegate) Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &manager{
		ctx:      ctx,
		cancel:   cancel,
		delegate: delegate,
		logger:   logging.NewLogger("call:" + strconv.FormatUint(delegate.CallId(), 10)),
	}
}
