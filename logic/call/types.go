package call

import (
	"encoding/json"
	"ichat-go/sched"
)

const userIdInvalid = 0

const (
	userStateInvited  = 1
	userStateRejected = 2
	userStateAccepted = 3
	userStateOnline   = 4
	userStateOffline  = 5
	userStateDead     = 6
)

const (
	userExitReasonNormal      = 1
	userExitReasonLost        = 2
	userExitReasonRejected    = 3
	userExitReasonReenterBusy = 4 /* 已退出用户重新进入，但是无法获取锁 */
)

const (
	userPingNone = -2
	userPingLost = -1
)

type UserState struct {
	UserId uint64 `json:"userId"`
	State  int    `json:"state"`
	Ping   int    `json:"ping"`
}

const (
	actionTypeUserAccepted = 1
	actionTypeUserOnline   = 2
	actionTypeUserOffline  = 3
	actionTypeHangup       = 4
	actionTypeSignaling    = 5
	actionTypeHeartBeat    = 6
)

type actionSignaling struct {
	FromUserId uint64 `json:"fromUserId"`
	ToUserId   uint64 `json:"toUserId"`
	Message    string `json:"message"`
}

func newActionMessage(t int, payload any) sched.Message {
	p, _ := json.Marshal(payload)
	return sched.Message{Type: t, Payload: p}
}

func userOnlineAction(m *sched.Message) uint64 {
	var id uint64
	_ = json.Unmarshal(m.Payload, &id)
	return id
}

func userAcceptedAction(m *sched.Message) uint64 {
	return userOnlineAction(m)
}

func userOfflineAction(m *sched.Message) uint64 {
	return userOnlineAction(m)
}

func hangupAction(m *sched.Message) uint64 {
	return userOnlineAction(m)
}

func signalingAction(m *sched.Message) actionSignaling {
	var a actionSignaling
	_ = json.Unmarshal(m.Payload, &a)
	return a
}

func heartBeatAction(m *sched.Message) uint64 {
	return userOnlineAction(m)
}

type wsMessage struct {
	Type    int    `json:"type"`
	Payload string `json:"payload"`
}

type payloadSignaling struct {
	ToUserId uint64 `json:"toUserId"`
	Message  string `json:"message"`
}

const (
	wsMessageTypeUnauthorized     = 1
	wsMessageTypeUpdateUserStates = 2
	wsMessageTypeUpdateUserState  = 3
	wsMessageTypeSignaling        = 4
	wsMessageTypeHeartBeat        = 5
	wsMessageTypeCallStart        = 6
	wsMessageTypeCallEnd          = 7
	wsMessageTypeError            = 8
)

const (
	wsActionTypeUpdateUserStates = 1
	wsActionTypeUpdateUserState  = 2
	wsActionTypeSignaling        = 3
	wsActionTypeCallStart        = 4
	wsActionTypeCallEnd          = 5
	wsActionTypeClose            = 6
)

type wsActionSignaling struct {
	FromUserId uint64 `json:"fromUserId"`
	Message    string `json:"message"`
}

func signalingPayload(m *wsMessage) payloadSignaling {
	var p payloadSignaling
	_ = json.Unmarshal([]byte(m.Payload), &p)
	return p
}
