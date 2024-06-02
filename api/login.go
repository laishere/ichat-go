package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

func loginApis(r *gin.RouterGroup) {
	r.POST("", func(c *gin.Context) {
		var d dto.LoginDto
		mustBindBody(c, &d)
		ok(c, logic.Login(&d))
	})
}
