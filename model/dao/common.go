package dao

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"ichat-go/db"
	"time"
)

func rdb() *redis.Client {
	return db.RedisClient()
}

func checkIsEmpty(tx *gorm.DB) bool {
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return true
	}
	if tx.Error != nil {
		panic(tx.Error)
	}
	return false
}

func rowsAffected(tx *gorm.DB) int64 {
	if tx.Error != nil {
		panic(tx.Error)
	}
	return tx.RowsAffected
}

func assertNoError(tx *gorm.DB) {
	if tx.Error != nil {
		panic(tx.Error)
	}
}

func rdbGet(key string, dest any) bool {
	c := rdb()
	result, err := c.Get(c.Context(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false
		}
		panic(err)
	}
	err = json.Unmarshal([]byte(result), dest)
	if err != nil {
		panic(err)
	}
	return true
}

func rbdSet(key string, value any, expiration time.Duration) {
	bytes, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	c := rdb()
	err = c.Set(c.Context(), key, bytes, expiration).Err()
	if err != nil {
		panic(err)
	}
}
