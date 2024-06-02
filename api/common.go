package api

import (
	"github.com/gin-gonic/gin"
	apperror "ichat-go/errs"
	"ichat-go/model"
	"ichat-go/utils"
	"ichat-go/validate"
)

func ok(c *gin.Context, data ...any) {
	if len(data) == 0 {
		c.JSON(200, model.EmptyResponse)
		return
	}
	utils.Assert(len(data) == 1)
	c.JSON(200, model.OkResponse(data[0]))
}

func handleBindErr(err error) {
	if err != nil {
		validate.HandleError(err)
		panic(apperror.NewAppError(apperror.CodeBadRequest, "请求参数错误"))
	}
}

func mustBindQuery(c *gin.Context, obj any) {
	err := c.ShouldBindQuery(obj)
	handleBindErr(err)
}

func mustBindBody(c *gin.Context, obj any) {
	err := c.ShouldBindJSON(obj)
	handleBindErr(err)
}

func mustBindParams(c *gin.Context, obj any) {
	err := c.ShouldBindUri(obj)
	handleBindErr(err)
}
