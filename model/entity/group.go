package entity

import "time"

type Group struct {
	GroupId   uint64    `json:"groupId" gorm:"primaryKey"`
	OwnerId   uint64    `json:"ownerId"`
	RoomId    uint64    `json:"roomId"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (g *Group) TableName() string {
	return "chat_groups"
}

type GroupMember struct {
	Id        uint64    `json:"id" gorm:"primaryKey"`
	GroupId   uint64    `json:"groupId"`
	UserId    uint64    `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
