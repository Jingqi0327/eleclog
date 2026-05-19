package worker

import (
	"fmt"

	"github.com/Jingqi0327/eleclog/logger"
	"go.uber.org/zap/zapcore"
)

type AsynqLogger struct{}

func NewAsynqLogger() *AsynqLogger {
	return &AsynqLogger{}
}

func (al *AsynqLogger) Printf(level zapcore.Level, args ...interface{}) {
	logger.Log.Log(level, "[Asynq] "+fmt.Sprintf("%v", args...))
}

func (al *AsynqLogger) Debug(args ...interface{}) {
	al.Printf(zapcore.DebugLevel, args...)
}

func (al *AsynqLogger) Info(args ...interface{}) {
	al.Printf(zapcore.InfoLevel, args...)
}

func (al *AsynqLogger) Warn(args ...interface{}) {
	al.Printf(zapcore.WarnLevel, args...)
}

func (al *AsynqLogger) Error(args ...interface{}) {
	al.Printf(zapcore.ErrorLevel, args...)
}

func (al *AsynqLogger) Fatal(args ...interface{}) {
	al.Printf(zapcore.FatalLevel, args...)
}
