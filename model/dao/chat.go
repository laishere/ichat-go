package dao

import (
	"fmt"
	_ "gorm.io/gorm"
	"ichat-go/model/entity"
	"ichat-go/utils"
)

type ChatDao interface {
	CreateChatRoom(e *entity.ChatRoom)
	UpdateRoomName(e *entity.ChatRoom)
	FindRoomById(roomId uint64) *entity.ChatRoom
	FindOrCreateChatRoomForContact(c *entity.Contact)
	CreateMessage(e *entity.ChatMessage)
	UpdateMessage(e *entity.ChatMessage)
	FindMessageById(messageId uint64) *entity.ChatMessage
	UpdateCallId(e *entity.ChatMessage)
	CreateDelivery(e *entity.MessageDelivery)
	GetMessages(roomId uint64, lastMessageId uint64, limit int) []*entity.ChatMessage
	FindLastMessage(roomId uint64) *entity.ChatMessage
	FindLastMessageId(roomId uint64) uint64
	GetDeliveries(receiverId, lastId, syncedId uint64, limit int) []*DeliveryMessage
	FindLastDeliveryId(receiverId uint64) uint64
}

type chatDao struct {
	tx Tx
}

type DeliveryMessage struct {
	entity.ChatMessage
	DeliveryId uint64 `json:"deliveryId"`
}

func (d chatDao) CreateChatRoom(e *entity.ChatRoom) {
	assertNoError(d.tx.Create(e))
}

func (d chatDao) UpdateRoomName(e *entity.ChatRoom) {
	assertNoError(d.tx.Model(e).Update("name", e.Name))
}

func (d chatDao) FindRoomById(roomId uint64) *entity.ChatRoom {
	var room entity.ChatRoom
	tx := d.tx.First(&room, roomId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &room
}

func (d chatDao) FindOrCreateChatRoomForContact(c *entity.Contact) {
	if c.RoomId != 0 {
		return
	}
	var name string
	if c.UserId != 0 {
		var uid1, uid2 uint64
		if c.OwnerId < c.UserId {
			uid1 = c.OwnerId
			uid2 = c.UserId
		} else {
			uid1 = c.UserId
			uid2 = c.OwnerId
		}
		name = fmt.Sprintf("u-%d-%d", uid1, uid2)
	} else {
		panic("不能为群组")
	}
	var room entity.ChatRoom
	tx := d.tx.Select("room_id").First(&room, "name = ?", name)
	if checkIsEmpty(tx) {
		room = entity.ChatRoom{
			Name: name,
		}
		d.CreateChatRoom(&room)
	}
	c.RoomId = room.RoomId
}

func (d chatDao) CreateMessage(e *entity.ChatMessage) {
	utils.Assert(e.Type != 0)
	assertNoError(d.tx.Create(e))
}

func (d chatDao) UpdateMessage(e *entity.ChatMessage) {
	assertNoError(d.tx.Model(e).Updates(map[string]interface{}{
		"text":      e.Text,
		"image":     e.Image,
		"thumbnail": e.Thumbnail,
		"revoked":   e.Revoked,
	}))
}

func (d chatDao) FindMessageById(messageId uint64) *entity.ChatMessage {
	var message entity.ChatMessage
	tx := d.tx.First(&message, messageId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &message
}

func (d chatDao) UpdateCallId(e *entity.ChatMessage) {
	assertNoError(d.tx.Model(e).Update("call_id", e.CallId))
}

func (d chatDao) CreateDelivery(e *entity.MessageDelivery) {
	assertNoError(d.tx.Create(e))
}

func (d chatDao) GetMessages(roomId uint64, lastMessageId uint64, limit int) []*entity.ChatMessage {
	var messages []*entity.ChatMessage
	tx := d.tx.Where("room_id = ?", roomId).Order("message_id desc").Limit(limit)
	if lastMessageId != 0 {
		tx = tx.Where("message_id < ?", lastMessageId)
	}
	assertNoError(tx.Find(&messages))
	return messages
}

func (d chatDao) FindLastMessage(roomId uint64) *entity.ChatMessage {
	var message entity.ChatMessage
	tx := d.tx.Where("room_id = ?", roomId).Order("message_id desc").First(&message)
	if checkIsEmpty(tx) {
		return nil
	}
	return &message
}

func (d chatDao) FindLastMessageId(roomId uint64) uint64 {
	var message entity.ChatMessage
	tx := d.tx.Where("room_id = ?", roomId).Order("message_id desc").First(&message)
	if checkIsEmpty(tx) {
		return 0
	}
	return message.MessageId
}

func (d chatDao) GetDeliveries(receiverId, lastId, syncedId uint64, limit int) []*DeliveryMessage {
	tx := d.tx.Model(&entity.MessageDelivery{}).
		Select("chat_messages.*, message_deliveries.id as delivery_id").
		Joins("LEFT JOIN chat_messages ON message_deliveries.message_id = chat_messages.message_id").
		Where("message_deliveries.id > ?", syncedId).
		Where("receiver_id = ?", receiverId).
		Order("message_deliveries.id DESC").
		Limit(limit)
	if lastId != 0 {
		tx = tx.Where("message_deliveries.id < ?", lastId)
	}
	var messages []*DeliveryMessage
	assertNoError(tx.Find(&messages))
	return messages
}

func (d chatDao) FindLastDeliveryId(receiverId uint64) uint64 {
	var delivery entity.MessageDelivery
	tx := d.tx.Where("receiver_id = ?", receiverId).Order("id DESC").First(&delivery)
	if checkIsEmpty(tx) {
		return 0
	}
	return delivery.Id
}

func NewChatDao(tx Tx) ChatDao {
	return chatDao{tx: tx}
}
