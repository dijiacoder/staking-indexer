package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 全局日志实例
var Logger *zap.Logger

func init() {
	// 使用带颜色的开发配置，保留不同级别的颜色区分
	config := zap.NewDevelopmentConfig()

	// 自定义编码器配置，使用带颜色的级别编码器
	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// Sync 刷新日志缓冲区
func Sync() {
	_ = Logger.Sync()
}
