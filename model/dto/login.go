package dto

type LoginDto struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResultDto struct {
	UserId uint64 `json:"userId"`
	Token  string `json:"token"`
}
