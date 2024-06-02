package entity

import "time"

type LoginUser struct {
	LoginId  string    `json:"loginId"`
	UserId   uint64    `json:"userId"`
	Enabled  bool      `json:"enabled"`
	LoginAt  time.Time `json:"loginAt"`
	ExpireAt time.Time `json:"expireAt"`
}
