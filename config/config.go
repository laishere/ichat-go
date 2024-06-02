package config

import (
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"ichat-go/logging"
	"os"
)

type Config struct {
	Jwt       JwtConfig   `yaml:"jwt"`
	Mysql     MysqlConfig `yaml:"mysql"`
	Redis     RedisConfig `yaml:"redis"`
	ApiPrefix string      `yaml:"api-prefix"`
	LogLevel  string      `yaml:"log-level"`
	UploadDir string      `yaml:"upload-dir"`
	Dev       bool        `yaml:"dev"`
}

var App Config

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func findConfigFile() string {
	list := []string{"config"}
	for _, f := range list {
		if exists(f + ".yml") {
			return f + ".yml"
		}
		if exists(f + ".yaml") {
			return f + ".yaml"
		}
	}
	return ""
}

func parseLogLevel() {
	lv := zapcore.DebugLevel
	switch App.LogLevel {
	case "debug":
		lv = zapcore.DebugLevel
	case "info":
		lv = zapcore.InfoLevel
	case "warn":
		lv = zapcore.WarnLevel
	case "error":
		lv = zapcore.ErrorLevel
	}
	logging.ConfigLogLevel = lv
}

func fillDefault() {
	if App.ApiPrefix == "" {
		App.ApiPrefix = "/api"
	}
	if App.UploadDir == "" {
		App.UploadDir = "upload"
	}
	if App.Redis.Host == "" {
		App.Redis.Host = "localhost"
	}
	if App.Redis.Port == 0 {
		App.Redis.Port = 6379
	}
}

func Init() {
	// parse yaml file
	file := findConfigFile()
	if file == "" {
		panic("config file not found")
	}
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &App)
	if err != nil {
		panic(err)
	}
	fillDefault()
	parseLogLevel()
}
