package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/ctx"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

type contactRequestIdParams struct {
	RequestId uint64 `form:"requestId"`
}

type contactIdParams struct {
	ContactId uint64 `form:"contactId"`
}

type groupInfosParams struct {
	GroupIds []uint64 `json:"groupIds"`
}

func contactApis(g *gin.RouterGroup) {
	g.POST("/user", func(c *gin.Context) {
		var d dto.AddUserContactDto
		mustBindBody(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ContactRequestAddUser(myId, &d))
	})
	g.POST("/accept", func(c *gin.Context) {
		var p contactRequestIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		logic.ContactRequestAccept(myId, p.RequestId)
		ok(c)
	})
	g.POST("/reject", func(c *gin.Context) {
		var p contactRequestIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		logic.ContactRequestReject(myId, p.RequestId)
		ok(c)
	})
	g.GET("", func(c *gin.Context) {
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ContactGetAll(myId))
	})
	g.GET("/members", func(c *gin.Context) {
		var p contactIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ContactGetMembers(myId, p.ContactId))
	})
	g.GET("/pending", func(c *gin.Context) {
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ContactRequestGetAllPending(myId))
	})
	g.POST("/group", func(c *gin.Context) {
		var d dto.CreateGroupDto
		mustBindBody(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		logic.GroupCreate(myId, &d)
		ok(c)
	})
	g.POST("/groups", func(c *gin.Context) {
		var p groupInfosParams
		mustBindBody(c, &p)
		ok(c, logic.GroupGetInfos(p.GroupIds))
	})
}
