package logic

import (
	"database/sql"
	"fmt"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/logic/notification"
	"ichat-go/model/dao"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
	"ichat-go/utils/strs"
	"slices"
	"time"
)

type deliverItem struct {
	userId     uint64
	deliveryId uint64
}

type deliverCtx struct {
	tx      dao.Tx
	contact *entity.Contact
	room    *entity.ChatRoom
	m       *dto.ChatMessageDto
	new     bool
	items   []deliverItem
}

func (d *deliverCtx) findUserIds() []uint64 {
	if d.contact != nil {
		if d.contact.UserId != 0 {
			return []uint64{d.contact.UserId, d.contact.OwnerId}
		}
		return di.ENV().GroupDao().GetMemberUserIds(d.contact.GroupId)
	}
	if d.room.Name[0] == 'u' {
		var uid1, uid2 uint64
		_, _ = fmt.Sscanf(d.room.Name, "u-%d-%d", &uid1, &uid2)
		return []uint64{uid1, uid2}
	}
	var groupId uint64
	_, _ = fmt.Sscanf(d.room.Name, "g-%d", &groupId)
	return di.ENV().GroupDao().GetMemberUserIds(groupId)
}

func (d *deliverCtx) createDelivery(uid uint64) {
	delivery := &entity.MessageDelivery{
		MessageId:  d.m.MessageId,
		ReceiverId: uid,
		Status:     entity.MessageDeliveryStatusSending,
	}
	di.ENV().ChatDao(d.tx).CreateDelivery(delivery)
	d.items = append(d.items, deliverItem{userId: uid, deliveryId: delivery.Id})
}

func (d *deliverCtx) createDeliveries() {
	for _, uid := range d.findUserIds() {
		d.createDelivery(uid)
	}
}

func (d *deliverCtx) notifyUsers() {
	for _, item := range d.items {
		m := *d.m
		if item.userId != m.SenderId {
			m.LocalId = ""
		}
		m.DeliveryId = item.deliveryId
		if m.Call != nil && m.Call.Status != entity.CallStatusEnd {
			m.Call.Handled = di.ENV().CallDao(d.tx).IsHandled(m.Call.CallId, item.userId)
		}
		notification.SendChatMessage(item.userId, &m, d.new)
	}
}

func (d *deliverCtx) deliver() {
	if d.contact == nil && d.room == nil {
		panic("invalid")
	}
	d.createDeliveries()
	updateContactsLastMessage(d.tx, &d.m.ChatMessage)
	go d.notifyUsers()
}

func messageType(e *entity.ChatMessage) int {
	if e.Text != "" {
		return entity.ChatMessageTypeText
	} else if e.Image != "" || e.Thumbnail != "" {
		return entity.ChatMessageTypeImage
	} else if e.CallId != 0 {
		return entity.ChatMessageTypeCall
	}
	panic(errs.MessageTypeNotSupported)
}

func onMessageUpdated(tx dao.Tx, m *entity.ChatMessage) {
	d := messageToDto(m, true)
	ctx := deliverCtx{
		tx:   tx,
		room: di.ENV().ChatDao().FindRoomById(m.RoomId),
		m:    d,
		new:  false,
	}
	ctx.deliver()
}

func checkMessageForm(d *dto.SendMessageDto) {
	if d.Text == "" && d.Thumbnail == "" && d.Image == "" {
		panic(errs.MessageEmpty)
	}
}

func ChatSendMessage(senderId uint64, d *dto.SendMessageDto) *dto.ChatMessageDto {
	checkMessageForm(d)
	contact := di.ENV().ContactDao().FindContactById(d.ContactId)
	verifyContact(contact, senderId)
	tx := di.ENV().DB().Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	defer commitOrRollback(tx)
	chatDao := di.ENV().ChatDao(tx)
	message := &entity.ChatMessage{
		RoomId:    contact.RoomId,
		SenderId:  senderId,
		Text:      d.Text,
		Image:     d.Image,
		Thumbnail: d.Thumbnail,
	}
	message.Type = messageType(message)
	chatDao.CreateMessage(message)
	m := messageToDto(message, false)
	m.LocalId = d.LocalId // 发送消息时，将本地消息ID返回给客户端
	ctx := deliverCtx{
		tx:      tx,
		contact: contact,
		m:       m,
		new:     true,
	}
	ctx.deliver()
	return m
}

func ChatDelayUpload(myId uint64, d *dto.DelayUploadDto) {
	m := di.ENV().ChatDao().FindMessageById(d.MessageId)
	if m == nil || m.SenderId != myId ||
		m.Type != entity.ChatMessageTypeImage || m.Image != "" ||
		m.Revoked {
		panic(errs.Forbidden)
	}
	m.Image = d.Image
	tx := di.ENV().DB().Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	defer commitOrRollback(tx)
	di.ENV().ChatDao(tx).UpdateMessage(m)
	onMessageUpdated(tx, m)
}

func messageToDto(e *entity.ChatMessage, checkCall bool) *dto.ChatMessageDto {
	d := &dto.ChatMessageDto{
		ChatMessage: *e,
	}
	if checkCall && e.CallId != 0 {
		d.Call = findCall(e.CallId)
	}
	return d
}

func ChatGetHistoryMessages(myId uint64, d *dto.QueryChatMessageDto) []*dto.ChatMessageDto {
	if d.Limit == 0 {
		d.Limit = 10
	}
	contact := di.ENV().ContactDao().FindContactById(d.ContactId)
	verifyContact(contact, myId)
	messages := di.ENV().ChatDao().GetMessages(contact.RoomId, d.LastMessageId, d.Limit)
	list := make([]*dto.ChatMessageDto, 0, len(messages))
	for _, e := range messages {
		list = append(list, messageToDto(e, true))
	}
	slices.Reverse(list)
	return list
}

func ChatRevokeMessage(myId uint64, messageId uint64) {
	m := di.ENV().ChatDao().FindMessageById(messageId)
	if m == nil || m.SenderId != myId || m.Type == entity.ChatMessageTypeCall {
		panic(errs.Forbidden)
	}
	if time.Now().Add(-time.Minute * 2).After(m.CreatedAt) {
		panic(errs.MessageRevokeExpired)
	}
	m.Text = ""
	m.Image = ""
	m.Thumbnail = ""
	m.Revoked = true
	tx := di.ENV().DB().Begin(&sql.TxOptions{Isolation: sql.LevelReadCommitted})
	defer commitOrRollback(tx)
	di.ENV().ChatDao(tx).UpdateMessage(m)
	onMessageUpdated(tx, m)
}

func ChatSyncMessages(myId uint64, d *dto.SyncMessagesDto) []*dto.ChatMessageDto {
	messages := di.ENV().ChatDao().GetDeliveries(myId, d.Last, d.Synced, d.Limit)
	results := make([]*dto.ChatMessageDto, 0)
	for _, e := range messages {
		d := &dto.ChatMessageDto{
			ChatMessage: e.ChatMessage,
			DeliveryId:  e.DeliveryId,
		}
		if e.ChatMessage.CallId != 0 {
			d.Call = findCall(e.ChatMessage.CallId)
		}
		results = append(results, d)
	}
	return results
}

func updateContactsLastMessage(tx dao.Tx, m *entity.ChatMessage) {
	if m.RoomId == 0 {
		return
	}
	c := &entity.Contact{
		RoomId:             m.RoomId,
		LastMessageId:      m.MessageId,
		LastMessageContent: describeChatMessage(m),
		LastMessageTime:    &m.CreatedAt,
	}
	di.ENV().ContactDao(tx).UpdateLastMessageByRoomId(c)
}

func describeChatMessage(e *entity.ChatMessage) string {
	if e.Revoked {
		return "[消息已撤回]"
	}
	switch e.Type {
	case entity.ChatMessageTypeText:
		return strs.TakeFirstN(e.Text, 20, true)
	case entity.ChatMessageTypeImage:
		return "[图片]"
	case entity.ChatMessageTypeCall:
		return "[通话]"
	default:
		panic("invalid message type")
	}
}
