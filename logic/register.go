package logic

import (
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
	"ichat-go/security"
)

func Register(dto *dto.RegisterDto) dto.LoginResultDto {
	userDao := di.ENV().UserDao()
	if userDao.FindUserByUsername(dto.Username) != nil {
		panic(errs.UserExists)
	}
	user := &entity.User{
		Username: dto.Username,
		Password: security.EncodePassword(dto.Password),
		Nickname: dto.Nickname,
		Avatar:   dto.Avatar,
		Enabled:  true,
	}
	userDao.CreateUser(user)
	return login(user)
}
