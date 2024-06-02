package db

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"ichat-go/config"
	"ichat-go/utils"
)

var client *redis.Client

func initRedis() {
	c := config.App.Redis
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       0,
	})
	_, err := client.Ping(client.Context()).Result()
	utils.PanicIfError(err)
}

func RedisClient() *redis.Client {
	return client
}
