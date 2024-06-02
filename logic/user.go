package logic

import (
	"ichat-go/di"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
)

func UserGetInfos(userIds []uint64) []*entity.User {
	users := di.ENV().UserDao().FindUsersByUserId(userIds)
	m := make(map[uint64]*entity.User)
	for _, user := range users {
		m[user.UserId] = user
	}
	result := make([]*entity.User, 0, len(userIds))
	for _, userId := range userIds {
		user, _ := m[userId]
		result = append(result, user)
	}
	return result
}

func UserUpdateInfo(userId uint64, d *dto.UpdateUserInfoDto) {
	di.ENV().UserDao().UpdateUser(userId, &entity.User{
		Nickname: d.Nickname,
		Avatar:   d.Avatar,
	})
}

func UserSearch(myId uint64, username string, page int, size int) []*dto.SearchUserItem {
	users := di.ENV().UserDao().SearchUsers(myId, username, (page-1)*size, size)
	items := make([]*dto.SearchUserItem, 0, len(users))
	for _, user := range users {
		isFriend := di.ENV().ContactDao().CheckContactExists(myId, user.UserId)
		var pendingRequest *entity.ContactRequest
		if !isFriend {
			pendingRequest = findPendingRequest(myId, user.UserId)
		}
		items = append(items, &dto.SearchUserItem{
			User:           *user,
			IsFriend:       isFriend,
			PendingRequest: pendingRequest,
		})
	}
	return items
}

func UserGetSettings(myId uint64) *dto.UserSettingsDto {
	settings := di.ENV().UserDao().FindSettings(myId)
	if settings == nil {
		return nil
	}
	return &dto.UserSettingsDto{
		Wallpaper: settings.Wallpaper,
	}
}

func UserSaveSettings(myId uint64, d *dto.UserSettingsDto) {
	di.ENV().UserDao().UpdateSettings(&entity.UserSettings{
		UserId:    myId,
		Wallpaper: d.Wallpaper,
	})
}
