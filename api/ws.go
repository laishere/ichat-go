package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/logic/call"
	"ichat-go/logic/notification"
)

func wsApis(r *gin.RouterGroup) {
	r.GET("/notification", func(c *gin.Context) {
		notification.WebSocketHandler(c)
	})
	r.GET("/call", func(c *gin.Context) {
		call.WebSocketHandler(c)
	})
}
