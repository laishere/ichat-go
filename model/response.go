package model

import "github.com/gin-gonic/gin"

func OkResponse(data any) gin.H {
	return gin.H{
		"status": "ok",
		"code":   200,
		"data":   data,
	}
}

var EmptyResponse = OkResponse(nil)

func ErrorResponse(code int, message string) gin.H {
	return gin.H{
		"status":  "error",
		"code":    code,
		"message": message,
	}
}
