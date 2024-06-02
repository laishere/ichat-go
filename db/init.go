package db

func Init() {
	initMysql()
	initRedis()
}

func InitForTest() {
	initRedis()
}
