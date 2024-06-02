package ctx

import (
	"github.com/gin-gonic/gin"
	"ichat-go/errs"
	"ichat-go/model/entity"
)

func SetLoginUser(c *gin.Context, u *entity.LoginUser) {
	c.Set("login_user", u)
}

func GetLoginUser(c *gin.Context) *entity.LoginUser {
	u, exists := c.Get("login_user")
	if !exists {
		panic(errs.Unauthorized)
	}
	return u.(*entity.LoginUser)
}
