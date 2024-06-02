package main

import (
	"github.com/gin-gonic/gin"
	"ichat-go/api"
	"ichat-go/config"
	"ichat-go/daemon"
	"ichat-go/db"
	"ichat-go/middleware"
)

func main() {
	config.Init()
	db.Init()
	if !config.App.Dev {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(middleware.AppError)
	r.Use(middleware.JwtAuth)
	api.Init(r.Group(config.App.ApiPrefix))
	daemon.Run()
	_ = r.Run(":8080")
}
