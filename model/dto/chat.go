package dto

import "ichat-go/model/entity"

type SendMessageDto struct {
	ContactId uint64 `json:"contactId"`
	LocalId   string `json:"localId"`
	Text      string `json:"text" validate:"omitempty"`
	Image     string `json:"image" validate:"omitempty,url"`
	Thumbnail string `json:"thumbnail" validate:"omitempty"`
}

type DelayUploadDto struct {
	MessageId uint64 `json:"messageId"`
	Image     string `json:"image" validate:"url"`
}

type QueryChatMessageDto struct {
	ContactId     uint64 `form:"contactId"`
	LastMessageId uint64 `form:"lastMessageId" validate:"omitempty"`
	Limit         int    `form:"limit" validate:"omitempty,max=100"`
}

type CallDto struct {
	entity.Call
	Handled bool `json:"handled"`
}

type ChatMessageDto struct {
	entity.ChatMessage
	Call       *CallDto `json:"call"`
	LocalId    string   `json:"localId"`
	DeliveryId uint64   `json:"deliveryId"`
}

type NotificationMessageDto struct {
	ChatMessageDto
	IsNew bool `json:"isNew"`
}

type SyncMessagesDto struct {
	Synced uint64 `form:"synced"`
	Last   uint64 `form:"last" validate:"omitempty"`
	Limit  int    `form:"limit" validate:"min=5,max=50"`
}
