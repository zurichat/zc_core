package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func init() {
	var err error

	config := zap.NewProductionConfig()
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""
	config.EncoderConfig = encoderConfig

	log, err = config.Build(zap.AddCallerSkip(1))

	if err != nil {
		panic(err)
	}
}

func Info(message string, args ...interface{}) {
	log.Info(fmt.Sprintf(message, args...))
}

func Error(message string, args ...interface{}) {
	log.Error(fmt.Sprintf(message, args...))
}

func Debug(message string, args ...interface{}) {
	log.Debug(fmt.Sprintf(message, args...))
}
