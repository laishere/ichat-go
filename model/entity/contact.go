package entity

import "time"

const (
	ContactStatusNormal = 1
)

const (
	ContactRequestStatusPending  = 1
	ContactRequestStatusAccepted = 2
	ContactRequestStatusRejected = 3
	ContactRequestStatusExpired  = 4
)

type Contact struct {
	ContactId          uint64     `json:"contactId" gorm:"primaryKey"`
	OwnerId            uint64     `json:"ownerId"`
	UserId             uint64     `json:"userId"`
	GroupId            uint64     `json:"groupId"`
	RoomId             uint64     `json:"roomId"`
	Status             int        `json:"status"`
	LastMessageId      uint64     `json:"lastMessageId" gorm:"column:last_msg_id"`
	LastMessageTime    *time.Time `json:"lastMessageTime" gorm:"column:last_msg_time"`
	LastMessageContent string     `json:"lastMessageContent" gorm:"column:last_msg_content"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

type ContactRequest struct {
	Id         uint64    `json:"id" gorm:"primaryKey"`
	RequestUid uint64    `json:"requestUid"`
	UserId     uint64    `json:"userId"`
	Status     int       `json:"status"`
	ExpiredAt  time.Time `json:"expiredAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
