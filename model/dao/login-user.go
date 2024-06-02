package dao

import (
	"ichat-go/model/entity"
	"time"
)

type LoginUserDao interface {
	SaveLoginUser(u *entity.LoginUser)
	FindLoginUserByLoginId(id string) *entity.LoginUser
}

type loginUserDao struct {
}

func keyForLoginUser(id string) string {
	return "login_user:" + id
}

func (loginUserDao) SaveLoginUser(u *entity.LoginUser) {
	rbdSet(keyForLoginUser(u.LoginId), u, u.ExpireAt.Sub(time.Now()))
}

func (loginUserDao) FindLoginUserByLoginId(id string) *entity.LoginUser {
	var user entity.LoginUser
	if !rdbGet(keyForLoginUser(id), &user) {
		return nil
	}
	return &user
}

func NewLoginUserDao(_ Tx) LoginUserDao {
	return loginUserDao{}
}
