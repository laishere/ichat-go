package entity

import "time"

type User struct {
	UserId    uint64    `json:"userId" gorm:"primaryKey"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UserSettings struct {
	UserId    uint64    `json:"userId" gorm:"primaryKey"`
	Wallpaper string    `json:"wallpaper"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
