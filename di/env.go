package di

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"ichat-go/model/dao"
)

type Env interface {
	DB() *gorm.DB
	RDB() *redis.Client
	UserDao(t ...dao.Tx) dao.UserDao
	LoginUserDao(t ...dao.Tx) dao.LoginUserDao
	ContactDao(t ...dao.Tx) dao.ContactDao
	ChatDao(t ...dao.Tx) dao.ChatDao
	GroupDao(t ...dao.Tx) dao.GroupDao
	CallDao(t ...dao.Tx) dao.CallDao
}

var env Env = &app{}

func SetEnv(e Env) {
	env = e
}

func ENV() Env {
	return env
}
