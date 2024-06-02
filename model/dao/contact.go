package dao

import (
	"ichat-go/model/entity"
	"ichat-go/utils"
)

type ContactDao interface {
	FindUserContact(ownerId uint64, userId uint64) *entity.Contact
	FindGroupContact(ownerId uint64, groupId uint64) *entity.Contact
	FindContactById(id uint64) *entity.Contact
	CreateContact(c *entity.Contact)
	FindPendingRequest(uid1 uint64, uid2 uint64) *entity.ContactRequest
	CreateContactRequest(c *entity.ContactRequest)
	FindContactRequestById(id uint64) *entity.ContactRequest
	UpdateContactRequestStatus(id uint64, status int)
	CheckContactExists(ownerId uint64, userId uint64) bool
	GetAll(ownerId uint64) []*entity.Contact
	GetAllPendingRequests(receiverId uint64) []*entity.ContactRequest
	UpdateLastMessageByRoomId(c *entity.Contact)
}

type contactDao struct {
	tx Tx
}

func (d contactDao) FindUserContact(ownerId uint64, userId uint64) *entity.Contact {
	var contact entity.Contact
	tx := d.tx.First(&contact, "owner_id = ? and user_id = ? and group_id is NULL", ownerId, userId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &contact
}

func (d contactDao) FindGroupContact(ownerId uint64, groupId uint64) *entity.Contact {
	var contact entity.Contact
	tx := d.tx.First(&contact, "owner_id = ? and user_id is NULL and group_id = ?", ownerId, groupId)
	if checkIsEmpty(tx) {
		return nil
	}
	return &contact
}

func (d contactDao) FindContactById(id uint64) *entity.Contact {
	var contact entity.Contact
	tx := d.tx.First(&contact, id)
	if checkIsEmpty(tx) {
		return nil
	}
	return &contact
}

func (d contactDao) CreateContact(c *entity.Contact) {
	utils.Assert(c.RoomId != 0)
	tx := d.tx
	// 外键约束
	if c.UserId == 0 {
		tx = tx.Omit("user_id")
	} else {
		tx = tx.Omit("group_id")
	}
	tx = tx.Create(c)
	assertNoError(tx)
}

func (d contactDao) CreateContactRequest(c *entity.ContactRequest) {
	assertNoError(d.tx.Create(c))
}

func (d contactDao) FindContactRequestById(id uint64) *entity.ContactRequest {
	var request entity.ContactRequest
	tx := d.tx.First(&request, id)
	if checkIsEmpty(tx) {
		return nil
	}
	return &request
}

func (d contactDao) FindPendingRequest(uid1 uint64, uid2 uint64) *entity.ContactRequest {
	var request entity.ContactRequest
	tx := d.tx.Where("((request_uid = ? and user_id = ?) or (request_uid = ? and user_id = ?)) and status = ?",
		uid1, uid2, uid2, uid1, entity.ContactRequestStatusPending).Order("id DESC").First(&request)
	if checkIsEmpty(tx) {
		return nil
	}
	return &request
}

func (d contactDao) UpdateContactRequestStatus(id uint64, status int) {
	tx := d.tx.Model(&entity.ContactRequest{}).Where("id = ?", id).Update("status", status)
	assertNoError(tx)
}

func (d contactDao) CheckContactExists(ownerId uint64, userId uint64) bool {
	var count int64
	tx := d.tx.Model(&entity.Contact{}).
		Where("owner_id = ? and user_id = ?", ownerId, userId).
		Limit(1).Count(&count)
	assertNoError(tx)
	return count > 0
}

func (d contactDao) GetAll(ownerId uint64) []*entity.Contact {
	var contacts []*entity.Contact
	tx := d.tx.Where("owner_id = ?", ownerId).Order("updated_at DESC").Find(&contacts)
	assertNoError(tx)
	return contacts
}

func (d contactDao) GetAllPendingRequests(receiverId uint64) []*entity.ContactRequest {
	var requests []*entity.ContactRequest
	tx := d.tx.Where("user_id = ? and status = ?", receiverId, entity.ContactRequestStatusPending).
		Order("created_at DESC").Find(&requests)
	assertNoError(tx)
	return requests
}

func (d contactDao) UpdateLastMessageByRoomId(c *entity.Contact) {
	update := entity.Contact{
		LastMessageId:      c.LastMessageId,
		LastMessageTime:    c.LastMessageTime,
		LastMessageContent: c.LastMessageContent,
	}
	tx := d.tx.Model(&entity.Contact{}).
		Where("room_id = ?", c.RoomId).
		Where("(select message_id from chat_messages where room_id = ? order by message_id desc limit 1) = ?", c.RoomId, c.LastMessageId).
		UpdateColumns(update)
	assertNoError(tx)
}

func NewContactDao(tx Tx) ContactDao {
	return contactDao{tx: tx}
}
