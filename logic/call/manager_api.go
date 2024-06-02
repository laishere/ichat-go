package call

import (
	"ichat-go/sched"
)

type ManagerApi interface {
	UserJoined(userId uint64)
	UserOnline(userId uint64)
	UserOffline(userId uint64)
	Hangup(userId uint64)
	Signaling(fromUserId, toUserId uint64, message string)
	HeartBeat(userId uint64)
}

type managerApi struct {
	mq sched.MQ
}

func FindManager(callId uint64) ManagerApi {
	mq := sched.NewMQ(managerKey(callId))
	if mq.State() != 1 {
		return nil
	}
	return &managerApi{mq: mq}
}

func (m *managerApi) UserJoined(userId uint64) {
	_ = m.mq.Push(newActionMessage(actionTypeUserAccepted, userId))
}

func (m *managerApi) UserOnline(userId uint64) {
	_ = m.mq.Push(newActionMessage(actionTypeUserOnline, userId))
}

func (m *managerApi) UserOffline(userId uint64) {
	_ = m.mq.Push(newActionMessage(actionTypeUserOffline, userId))
}

func (m *managerApi) Hangup(userId uint64) {
	_ = m.mq.Push(newActionMessage(actionTypeHangup, userId))
}

func (m *managerApi) Signaling(fromUserId, toUserId uint64, message string) {
	_ = m.mq.Push(newActionMessage(actionTypeSignaling, actionSignaling{FromUserId: fromUserId, ToUserId: toUserId, Message: message}))
}

func (m *managerApi) HeartBeat(userId uint64) {
	_ = m.mq.Push(newActionMessage(actionTypeHeartBeat, userId))
}
