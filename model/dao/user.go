package dao

import (
	"ichat-go/model/entity"
)

type UserDao interface {
	FindUserByUserId(userId uint64) *entity.User
	FindUsersByUserId(userIds []uint64) []*entity.User
	FindUserByUsername(username string) *entity.User
	SearchUsers(myId uint64, username string, offset int, count int) []*entity.User
	CreateUser(user *entity.User)
	UpdateUser(userId uint64, u *entity.User)
	FindSettings(userId uint64) *entity.UserSettings
	UpdateSettings(s *entity.UserSettings)
}

type userDao struct {
	tx Tx
}

func (d userDao) FindUserByUserId(userId uint64) *entity.User {
	var user entity.User
	tx := d.tx.First(&user, userId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &user
}

func (d userDao) FindUserByUsername(username string) *entity.User {
	var user entity.User
	tx := d.tx.First(&user, "username = ?", username)
	if checkIsEmpty(tx) {
		return nil
	}
	return &user
}

func (d userDao) CreateUser(user *entity.User) {
	tx := d.tx.Create(user)
	assertNoError(tx)
}

func (d userDao) FindUsersByUserId(userIds []uint64) []*entity.User {
	var users []*entity.User
	d.tx.Find(&users, userIds)
	return users
}

func (d userDao) UpdateUser(userId uint64, u *entity.User) {
	tx := d.tx.
		Model(&entity.User{UserId: userId}).
		Updates(u)
	assertNoError(tx)
}

func (d userDao) SearchUsers(myId uint64, username string, offset int, count int) []*entity.User {
	var users []*entity.User
	d.tx.
		Where("username like ? and user_id != ?", username+"%", myId).
		Order("user_id ASC").
		Offset(offset).
		Limit(count).
		Find(&users)
	return users
}

func (d userDao) FindSettings(userId uint64) *entity.UserSettings {
	var settings entity.UserSettings
	tx := d.tx.First(&settings, userId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &settings
}

func (d userDao) UpdateSettings(s *entity.UserSettings) {
	tx := d.tx.
		Model(s).
		Updates(map[string]any{
			"wallpaper": s.Wallpaper,
		})
	if tx.RowsAffected == 0 {
		tx = d.tx.Create(s)
	}
	assertNoError(tx)
}

func NewUserDao(tx Tx) UserDao {
	return userDao{tx: tx}
}
