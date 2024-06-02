package dto

type CreateGroupDto struct {
	Name       string   `json:"name"`
	Avatar     string   `json:"avatar" validate:"omitempty,url"`
	ContactIds []uint64 `json:"contactIds" validate:"min=1,max=100"`
}
