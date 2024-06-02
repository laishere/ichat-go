package entity

import "time"

const (
	ChatMessageTypeText  = 1
	ChatMessageTypeImage = 2
	ChatMessageTypeCall  = 3
)

const (
	MessageDeliveryStatusSending  = 1
	MessageDeliveryStatusReceived = 2
	MessageDeliveryStatusRead     = 3
)

type ChatRoom struct {
	RoomId    uint64    `json:"roomId" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ChatMessage struct {
	MessageId uint64    `json:"messageId" gorm:"primaryKey"`
	RoomId    uint64    `json:"roomId"`
	SenderId  uint64    `json:"senderId"`
	Type      int       `json:"type"`
	Text      string    `json:"text"`
	Image     string    `json:"image"`
	Thumbnail string    `json:"thumbnail"`
	CallId    uint64    `json:"callId"`
	Revoked   bool      `json:"revoked"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type MessageDelivery struct {
	Id         uint64    `json:"id" gorm:"primaryKey"`
	MessageId  uint64    `json:"messageId"`
	ReceiverId uint64    `json:"receiverId"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (*MessageDelivery) TableName() string {
	return "message_deliveries"
}
