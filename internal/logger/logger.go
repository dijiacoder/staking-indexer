package logger

import (
	"go.uber.org/zap"
)

// Logger 全局日志实例
var Logger *zap.Logger

func init() {
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
}

// Sync 刷新日志缓冲区
func Sync() {
	_ = Logger.Sync()
}