package db

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"ichat-go/config"
	"ichat-go/logging"
	"log"
	"os"
	"time"
)

var db *gorm.DB

func waitForMysql() {
	maxTries := 10
	sleep := time.Second
	dsn := config.App.Mysql.Dsn()
	for i := 0; i < maxTries; i++ {
		var err error
		var d *sql.DB
		d, err = sql.Open("mysql", dsn)
		if err == nil {
			err = d.Ping()
		}
		if err == nil {
			_ = d.Close()
			break
		}
		if i == 0 {
			logging.NewLogger("mysql").Info("Waiting for mysql to be ready")
		}
		if i == maxTries-1 {
			panic(err)
		}
		time.Sleep(sleep)
		sleep = time.Second * 3
	}
}

func initMysql() {
	var err error
	var l logger.Interface
	if config.App.Dev {
		l = logger.Default.LogMode(logger.Info)
	} else {
		waitForMysql()
		l = logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		})
	}
	db, err = gorm.Open(mysql.Open(config.App.Mysql.Dsn()), &gorm.Config{
		Logger: l,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		panic(err)
	}
}

func MysqlDB() *gorm.DB {
	return db
}
