package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/logic"
	"ichat-go/model/dto"
)

func registerApis(r *gin.RouterGroup) {
	r.POST("", func(c *gin.Context) {
		var d dto.RegisterDto
		mustBindBody(c, &d)
		ok(c, logic.Register(&d))
	})
}
