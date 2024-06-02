package api

import (
	"github.com/gin-gonic/gin"
	"ichat-go/logic"
)

func fileApis(g *gin.RouterGroup) {
	g.POST("/upload", func(c *gin.Context) {
		ok(c, logic.FileUpload(c))
	})
	g.GET("/*path", func(c *gin.Context) {
		path := c.Param("path")
		logic.FileDownload(path, c)
	})
}
