package dto

type CreateCallDto struct {
	ContactId uint64   `json:"contactId"`
	UserIds   []uint64 `json:"userIds"`
}
