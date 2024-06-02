package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"ichat-go/errs"
	"ichat-go/model"
)

func AppError(c *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			var appErr errs.AppError
			if e, ok := e.(error); ok && errors.As(e, &appErr) {
				var httpCode int
				if appErr.Code() < 1000 {
					httpCode = appErr.Code()
				} else {
					httpCode = 500
				}
				c.JSON(httpCode, model.ErrorResponse(appErr.Code(), e.Error()))
				c.Abort()
				return
			}
			c.JSON(500, model.ErrorResponse(500, "服务器内部错误"))
			c.Abort()
			panic(e) // print trace
		}
	}()
	c.Next()
}
