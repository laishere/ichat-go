package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger = *zap.SugaredLogger

var ConfigLogLevel = zapcore.DebugLevel

func NewLogger(tag string) Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(ConfigLogLevel)
	l, _ := config.Build()

	return l.Sugar().Named(tag)
}
