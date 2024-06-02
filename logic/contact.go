package logic

import (
	"gorm.io/gorm"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/logic/notification"
	"ichat-go/model/dto"
	"ichat-go/model/entity"
	"ichat-go/utils"
	"time"
)

func ContactRequestAddUser(myId uint64, d *dto.AddUserContactDto) *entity.ContactRequest {
	userId := d.UserId
	utils.Assert(myId != userId)
	contact := di.ENV().ContactDao().FindUserContact(myId, userId)
	if contact != nil {
		panic(errs.ContactExists)
	}
	user := di.ENV().UserDao().FindUserByUserId(userId)
	if user == nil {
		panic(errs.UserNotFound)
	}
	if findPendingRequest(myId, userId) != nil {
		panic(errs.ContactRequestExists)
	}
	request := &entity.ContactRequest{
		RequestUid: myId,
		UserId:     userId,
		Status:     entity.ContactRequestStatusPending,
		ExpiredAt:  time.Now().Add(time.Hour * 24),
	}
	di.ENV().ContactDao().CreateContactRequest(request)
	notification.SendNewContactRequest(userId, request)
	return request
}

func findPendingRequest(uid1, uid2 uint64) *entity.ContactRequest {
	if req := di.ENV().ContactDao().FindPendingRequest(uid1, uid2); req != nil {
		updateContactRequestStatusIfNeeded(req)
		if req.Status == entity.ContactRequestStatusPending {
			return req
		}
	}
	return nil
}

func ContactRequestGetAllPending(myId uint64) []*entity.ContactRequest {
	requests := make([]*entity.ContactRequest, 0)
	for _, request := range di.ENV().ContactDao().GetAllPendingRequests(myId) {
		updateContactRequestStatusIfNeeded(request)
		if request.Status == entity.ContactRequestStatusPending {
			requests = append(requests, request)
		}
	}
	return requests
}

func createUserContact(tx *gorm.DB, uid1, uid2 uint64) *entity.Contact {
	contact := &entity.Contact{
		OwnerId: uid1,
		UserId:  uid2,
		Status:  entity.ContactStatusNormal,
	}
	di.ENV().ChatDao(tx).FindOrCreateChatRoomForContact(contact)
	di.ENV().ContactDao(tx).CreateContact(contact)
	utils.Assert(contact.RoomId != 0)
	return contact
}

func updateContactRequestStatusIfNeeded(e *entity.ContactRequest) {
	if time.Now().After(e.ExpiredAt) && e.Status == entity.ContactRequestStatusPending {
		e.Status = entity.ContactRequestStatusExpired
		di.ENV().ContactDao().UpdateContactRequestStatus(e.Id, entity.ContactRequestStatusExpired)
	}
}

func checkContactRequest(myId uint64, request *entity.ContactRequest) {
	if request == nil {
		panic(errs.ContactRequestNotFound)
	}
	if request.UserId != myId {
		panic(errs.ContactRequestNotFound)
	}
	updateContactRequestStatusIfNeeded(request)
	if request.Status == entity.ContactRequestStatusExpired {
		panic(errs.ContactRequestExpired)
	}
	if request.Status != entity.ContactRequestStatusPending {
		panic(errs.ContactRequestStatusInvalid)
	}
}

func ContactRequestAccept(myId uint64, requestId uint64) {
	tx := di.ENV().DB().Begin()
	defer commitOrRollback(tx)
	contactDao := di.ENV().ContactDao(tx)
	request := contactDao.FindContactRequestById(requestId)
	checkContactRequest(myId, request)
	c1 := createUserContact(tx, request.RequestUid, request.UserId)
	c2 := createUserContact(tx, request.UserId, request.RequestUid)
	contactDao.UpdateContactRequestStatus(requestId, entity.ContactRequestStatusAccepted)
	notification.SendNewContact(c1.OwnerId, c1)
	notification.SendNewContact(c2.OwnerId, c2)
}

func ContactRequestReject(myId uint64, requestId uint64) {
	contactDao := di.ENV().ContactDao()
	request := contactDao.FindContactRequestById(requestId)
	checkContactRequest(myId, request)
	contactDao.UpdateContactRequestStatus(requestId, entity.ContactRequestStatusRejected)
}

func ContactGetAll(myId uint64) []*entity.Contact {
	return di.ENV().ContactDao().GetAll(myId)
}

func ContactGetMembers(myId uint64, contactId uint64) []*entity.User {
	c := di.ENV().ContactDao().FindContactById(contactId)
	verifyContact(c, myId)
	if c.UserId != 0 {
		return getUserContactMembers(c)
	}
	return getGroupMembers(c.GroupId)
}

func getUserContactMembers(c *entity.Contact) []*entity.User {
	user1 := di.ENV().UserDao().FindUserByUserId(c.OwnerId)
	user2 := di.ENV().UserDao().FindUserByUserId(c.UserId)
	return []*entity.User{user1, user2}
}

func verifyContact(c *entity.Contact, myId uint64) {
	if c == nil {
		panic(errs.ContactNotFound)
	}
	if c.OwnerId != myId {
		panic(errs.Forbidden)
	}
	utils.Assert(c.RoomId != 0)
}

func verifyContactMembers(c *entity.Contact, userIds []uint64) {
	if c.UserId != 0 {
		if len(userIds) > 1 || (len(userIds) == 1 && userIds[0] != c.UserId) {
			panic(errs.Forbidden)
		}
	} else if c.GroupId != 0 {
		verifyGroupMembers(c.GroupId, userIds)
	}
}
