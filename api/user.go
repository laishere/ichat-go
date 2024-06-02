package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/ctx"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

type searchUserParams struct {
	Keyword string `form:"keyword"`
	Page    int    `form:"page"`
	Size    int    `form:"size"`
}
type userInfosParams struct {
	UserIds []uint64 `json:"userIds"`
}

func userApis(r *gin.RouterGroup) {
	r.POST("/infos", func(c *gin.Context) {
		var p userInfosParams
		mustBindBody(c, &p)
		ok(c, logic.UserGetInfos(p.UserIds))
	})
	r.POST("/info", func(c *gin.Context) {
		userId := ctx.GetLoginUser(c).UserId
		var d dto.UpdateUserInfoDto
		mustBindBody(c, &d)
		logic.UserUpdateInfo(userId, &d)
		ok(c)
	})
	r.GET("/search", func(c *gin.Context) {
		var p searchUserParams
		mustBindQuery(c, &p)
		ok(c, logic.UserSearch(ctx.GetLoginUser(c).UserId, p.Keyword, p.Page, p.Size))
	})
	r.GET("/settings", func(c *gin.Context) {
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.UserGetSettings(myId))
	})
	r.POST("/settings", func(c *gin.Context) {
		myId := ctx.GetLoginUser(c).UserId
		var d dto.UserSettingsDto
		mustBindBody(c, &d)
		logic.UserSaveSettings(myId, &d)
		ok(c)
	})
}
