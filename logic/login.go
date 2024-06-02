package logic

import (
	"github.com/google/uuid"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/jwt"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
	"ichat-go/security"
	"time"
)

func createLoginUser(u *entity.User) *entity.LoginUser {
	return &entity.LoginUser{
		LoginId:  uuid.New().String(),
		UserId:   u.UserId,
		Enabled:  true,
		LoginAt:  time.Now(),
		ExpireAt: time.Now().Add(time.Hour * 24),
	}
}

func Login(d *dto.LoginDto) dto.LoginResultDto {
	user := di.ENV().UserDao().FindUserByUsername(d.Username)
	if user == nil {
		panic(errs.UserNotFound)
	}
	if !security.ComparePassword(user.Password, d.Password) {
		panic(errs.BadCredentials)
	}
	return login(user)
}

func login(user *entity.User) dto.LoginResultDto {
	loginUser := createLoginUser(user)
	di.ENV().LoginUserDao().SaveLoginUser(loginUser)
	token := jwt.GenerateToken(loginUser.LoginId, loginUser.ExpireAt)
	return dto.LoginResultDto{
		UserId: user.UserId,
		Token:  token,
	}
}
