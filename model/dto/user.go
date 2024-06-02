package dto

import "ichat-go/model/entity"

type UpdateUserInfoDto struct {
	Nickname string `json:"nickname" validate:"required,min=1,max=20"`
	Avatar   string `json:"avatar" validate:"omitempty,url"`
}

type SearchUserItem struct {
	entity.User
	IsFriend       bool                   `json:"isFriend"`
	PendingRequest *entity.ContactRequest `json:"pendingRequest"`
}

type UserSettingsDto struct {
	Wallpaper string `json:"wallpaper" validate:"omitempty,url"`
}
