package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/ctx"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

type callIdParams struct {
	CallId uint64 `form:"callId"`
}

func callApis(g *gin.RouterGroup) {
	g.POST("", func(c *gin.Context) {
		var d dto.CreateCallDto
		mustBindBody(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.CallCreate(myId, &d))
	})
	g.POST("/join", func(c *gin.Context) {
		var p callIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.CallJoin(myId, p.CallId))
	})
	g.POST("/hangup", func(c *gin.Context) {
		var p callIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		logic.CallHangup(myId, p.CallId)
		ok(c)
	})
	g.GET("", func(c *gin.Context) {
		var p callIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.CallInfo(myId, p.CallId))
	})
}
