package call

import (
	"fmt"
	"ichat-go/sched"
)

type Session interface {
	UpdateUserStates(states []UserState)
	UpdateUserState(state UserState)
	Signaling(fromUserId uint64, message string)
	CallStart()
	CallEnd(reason int)
	Close()
}

func sessionKey(callId, userId uint64) string {
	return fmt.Sprintf("call:ws:%d:%d", callId, userId)
}

func findSession(callId, userId uint64) Session {
	mq := sched.NewMQ(sessionKey(callId, userId))
	if mq.State() != 1 {
		return nil
	}
	return &wsSessionApi{mq: mq}
}

type wsSessionApi struct {
	mq sched.MQ
}

func (s *wsSessionApi) UpdateUserStates(states []UserState) {
	_ = s.mq.Push(newActionMessage(wsActionTypeUpdateUserStates, states))
}

func (s *wsSessionApi) UpdateUserState(state UserState) {
	_ = s.mq.Push(newActionMessage(wsActionTypeUpdateUserState, state))
}

func (s *wsSessionApi) Signaling(fromUserId uint64, message string) {
	_ = s.mq.Push(newActionMessage(wsActionTypeSignaling, wsActionSignaling{FromUserId: fromUserId, Message: message}))
}

func (s *wsSessionApi) CallStart() {
	_ = s.mq.Push(newActionMessage(wsActionTypeCallStart, nil))
}

func (s *wsSessionApi) CallEnd(reason int) {
	_ = s.mq.Push(newActionMessage(wsActionTypeCallEnd, reason))
}

func (s *wsSessionApi) Close() {
	_ = s.mq.Push(newActionMessage(wsActionTypeClose, nil))
}
