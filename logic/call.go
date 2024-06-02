package logic

import (
	"encoding/json"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/logic/call"
	"ichat-go/logic/notification"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
)

func encodeCallMembers(members []uint64) string {
	s, _ := json.Marshal(members)
	return string(s)
}

func decodeCallMembers(s string) []uint64 {
	var members []uint64
	_ = json.Unmarshal([]byte(s), &members)
	return members
}

func CallCreate(myId uint64, d *dto.CreateCallDto) uint64 {
	if len(d.UserIds) == 0 {
		panic(errs.CallMemberCountNotEnough)
	}
	contact := di.ENV().ContactDao().FindContactById(d.ContactId)
	verifyContact(contact, myId)
	verifyContactMembers(contact, d.UserIds)
	userIds := []uint64{myId}
	for _, userId := range d.UserIds {
		if userId == myId {
			panic(errs.NewAppError(errs.CodeBadRequest, "userIds不能包含自己"))
		}
		userIds = append(userIds, userId)
	}
	tx := di.ENV().DB().Begin()
	defer rollbackWhenPanic(tx)
	chatDao := di.ENV().ChatDao(tx)
	callDao := di.ENV().CallDao(tx)
	message := &entity.ChatMessage{
		RoomId:   contact.RoomId,
		SenderId: myId,
	}
	message.Type = entity.ChatMessageTypeCall
	chatDao.CreateMessage(message)
	c := &entity.Call{
		CallerId:  myId,
		MessageId: message.MessageId,
		Members:   encodeCallMembers(userIds),
		Status:    entity.CallStatusNew,
	}
	callDao.CreateCall(c)
	message.CallId = c.CallId
	chatDao.UpdateCallId(message)
	tx.Commit()
	delegate := call.NewManagerDelegate(c.CallId)
	mgr := call.NewManager(delegate)
	go mgr.Loop()
	return c.CallId
}

func manager(callId uint64) call.ManagerApi {
	mgr := call.FindManager(callId)
	if mgr == nil {
		panic(errs.CallManagerNotFound)
	}
	return mgr
}

func verifyCall(callId uint64) *entity.Call {
	c := di.ENV().CallDao().FindCallById(callId)
	if c == nil {
		panic(errs.CallNotFound)
	}
	if c.Status == entity.CallStatusNew {
		panic(errs.CallStatusNotReady)
	}
	if c.Status == entity.CallStatusEnd {
		panic(errs.CallStatusInvalid)
	}
	return c
}

func onCallHandled(myId uint64, c *entity.Call) {
	dao := di.ENV().CallDao()
	if !dao.IsHandled(c.CallId, myId) {
		dao.SetHandled(c.CallId, myId)
		notification.SendCallHandled(myId, c.CallId)
	}
}

func CallHangup(myId uint64, callId uint64) {
	c := verifyCall(callId)
	manager(callId).Hangup(myId)
	onCallHandled(myId, c)
}

func CallJoin(myId uint64, callId uint64) string {
	c := verifyCall(callId)
	checkCallMembers(myId, c)
	manager(callId).UserJoined(myId)
	token := call.GenerateToken(callId, myId)
	onCallHandled(myId, c)
	return token
}

func checkCallMembers(myId uint64, c *entity.Call) {
	for _, userId := range decodeCallMembers(c.Members) {
		if userId == myId {
			return
		}
	}
	panic(errs.Forbidden)
}

func CallInfo(myId uint64, callId uint64) *entity.Call {
	c := di.ENV().CallDao().FindCallById(callId)
	if c == nil {
		panic(errs.CallNotFound)
	}
	checkCallMembers(myId, c)
	return c
}

func findCall(callId uint64) *dto.CallDto {
	c := di.ENV().CallDao().FindCallById(callId)
	if c == nil {
		return nil
	}
	return &dto.CallDto{
		Call:    *c,
		Handled: true,
	}
}

func init() {
	// 依赖注入，避免循环依赖
	call.SetNotifyCallUpdateCallback(func(messageId uint64) {
		onMessageUpdated(nil, di.ENV().ChatDao().FindMessageById(messageId))
	})
}
