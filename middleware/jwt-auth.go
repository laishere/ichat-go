package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	jwt2 "github.com/golang-jwt/jwt"
	"ichat-go/ctx"
	"ichat-go/di"
	"ichat-go/errs"
	"ichat-go/jwt"
	"ichat-go/security"
)

func JwtAuth(c *gin.Context) {
	if security.IsWhiteListed(c.Request.URL.Path) {
		return
	}
	token := c.GetHeader("Authorization")
	if token == "" {
		token = c.Query("token")
	}
	if token == "" {
		panic(errs.Unauthorized)
	}
	loginId, e := jwt.ValidateToken(token)
	if e != nil {
		var e2 *jwt2.ValidationError
		if errors.As(e, &e2); e2.Errors == jwt2.ValidationErrorExpired {
			panic(errs.CredentialsExpired)
		}
		panic(errs.Unauthorized)
	}
	loginUser := di.ENV().LoginUserDao().FindLoginUserByLoginId(loginId)
	if loginUser == nil {
		panic(errs.CredentialsExpired)
	}
	ctx.SetLoginUser(c, loginUser)
	c.Next()
}
