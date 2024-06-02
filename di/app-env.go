package di

import (
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"ichat-go/db"
	"ichat-go/model/dao"
)

type app struct {
}

func (a *app) DB() *gorm.DB {
	return db.MysqlDB()
}

func (a *app) RDB() *redis.Client {
	return db.RedisClient()
}

func (a *app) txOrDB(t ...dao.Tx) *gorm.DB {
	if len(t) == 0 || t[0] == nil {
		return a.DB()
	}
	return t[0]
}

func (a *app) UserDao(t ...dao.Tx) dao.UserDao {
	return dao.NewUserDao(a.txOrDB(t...))
}

func (a *app) LoginUserDao(t ...dao.Tx) dao.LoginUserDao {
	return dao.NewLoginUserDao(a.txOrDB(t...))
}

func (a *app) ContactDao(t ...dao.Tx) dao.ContactDao {
	return dao.NewContactDao(a.txOrDB(t...))
}

func (a *app) ChatDao(t ...dao.Tx) dao.ChatDao {
	return dao.NewChatDao(a.txOrDB(t...))
}

func (a *app) GroupDao(t ...dao.Tx) dao.GroupDao {
	return dao.NewGroupDao(a.txOrDB(t...))
}

func (a *app) CallDao(t ...dao.Tx) dao.CallDao {
	return dao.NewCallDao(a.txOrDB(t...))
}
