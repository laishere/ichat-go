package logic

import (
	"fmt"
	"github.com/google/uuid"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/logic/notification"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
)

func verifyGroupMembers(groupId uint64, userIds []uint64) {
	ids := di.ENV().GroupDao().GetMemberUserIds(groupId)
	idSet := make(map[uint64]bool)
	for _, id := range ids {
		idSet[id] = true
	}
	for _, userId := range userIds {
		if !idSet[userId] {
			panic(errs.Forbidden)
		}
	}
}

func verifyContacts(myId uint64, contactIds []uint64) []uint64 {
	userIds := []uint64{myId}
	for _, contactId := range contactIds {
		c := di.ENV().ContactDao().FindContactById(contactId)
		verifyContact(c, myId)
		if c.UserId == 0 {
			panic(errs.NewAppError(errs.CodeBadRequest, "只能给用户联系人创建群组"))
		}
		userIds = append(userIds, c.UserId)
	}
	return userIds
}

func GroupCreate(myId uint64, d *dto.CreateGroupDto) {
	tx := di.ENV().DB().Begin()
	defer rollbackWhenPanic(tx)
	chatDao := di.ENV().ChatDao(tx)
	groupDao := di.ENV().GroupDao(tx)
	contactDao := di.ENV().ContactDao(tx)
	userIds := verifyContacts(myId, d.ContactIds)
	/* room 和 group循环依赖，只能先让一个为临时值 */
	room := &entity.ChatRoom{Name: uuid.New().String()}
	chatDao.CreateChatRoom(room)
	g := &entity.Group{
		RoomId:  room.RoomId,
		OwnerId: myId,
		Name:    d.Name,
		Avatar:  d.Avatar,
	}
	groupDao.CreateGroup(g)
	groupDao.CreateMember(g, userIds)
	room.Name = fmt.Sprintf("g-%d", g.GroupId)
	chatDao.UpdateRoomName(room)
	var contacts []*entity.Contact
	for _, userId := range userIds {
		c := &entity.Contact{
			OwnerId: userId,
			GroupId: g.GroupId,
			RoomId:  room.RoomId,
			Status:  entity.ContactStatusNormal,
		}
		contactDao.CreateContact(c)
		contacts = append(contacts, c)
	}
	tx.Commit()
	for _, c := range contacts {
		notification.SendNewContact(c.OwnerId, c)
	}
}

func GroupGetInfos(groupIds []uint64) []*entity.Group {
	return di.ENV().GroupDao().FindGroups(groupIds)
}

func getGroupMembers(groupId uint64) []*entity.User {
	return di.ENV().GroupDao().GetMembers(groupId)
}
