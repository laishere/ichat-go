package api

import "github.com/gin-gonic/gin"

func Init(r *gin.RouterGroup) {
	apiMap := map[string]func(*gin.RouterGroup){
		"login":    loginApis,
		"register": registerApis,
		"user":     userApis,
		"contact":  contactApis,
		"chat":     chatApis,
		"ws":       wsApis,
		"call":     callApis,
		"file":     fileApis,
	}
	for path, apis := range apiMap {
		g := r.Group(path)
		apis(g)
	}
}
