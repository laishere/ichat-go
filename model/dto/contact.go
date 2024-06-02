package dto

import (
	"ichat-go/model/entity"
)

type AddUserContactDto struct {
	UserId uint64 `form:"userId" validate:"required"`
}

type ContactDto = entity.Contact
