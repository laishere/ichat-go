package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/ctx"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

type messageIdParams struct {
	MessageId uint64 `form:"messageId"`
}

func chatApis(g *gin.RouterGroup) {
	g.POST("/send", func(c *gin.Context) {
		var d dto.SendMessageDto
		mustBindBody(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ChatSendMessage(myId, &d))
	})
	g.GET("/messages", func(c *gin.Context) {
		var d dto.QueryChatMessageDto
		mustBindQuery(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ChatGetHistoryMessages(myId, &d))
	})
	g.POST("/revoke", func(c *gin.Context) {
		var p messageIdParams
		mustBindQuery(c, &p)
		myId := ctx.GetLoginUser(c).UserId
		logic.ChatRevokeMessage(myId, p.MessageId)
		ok(c)
	})
	g.GET("/sync", func(c *gin.Context) {
		var d dto.SyncMessagesDto
		mustBindQuery(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		ok(c, logic.ChatSyncMessages(myId, &d))
	})
	g.POST("/delayUpload", func(c *gin.Context) {
		var d dto.DelayUploadDto
		mustBindBody(c, &d)
		myId := ctx.GetLoginUser(c).UserId
		logic.ChatDelayUpload(myId, &d)
		ok(c)
	})
}
