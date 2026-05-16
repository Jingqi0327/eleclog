package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger
var SugarLog *zap.SugaredLogger

var IsDev bool

// InitLogger 接受一个参数，决定是否是开发环境
func InitLogger(isDev bool) {
	IsDev = isDev
	// 1. 基础配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 时间格式：2026-05-16T15:04:05.000+0800

	var encoder zapcore.Encoder

	// 2. 根据环境动态选择编码器
	if isDev {
		// 开发环境：使用带颜色的 Console 编码器
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 级别带颜色（如红色 ERROR，蓝色 INFO）
		// 可以自定义控制台输出的格式，这里使用标准的 Console 编码
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// 生产环境：使用 JSON 编码器
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 级别大写，不带颜色
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 3. 输出位置
	consoleSyncer := zapcore.AddSync(os.Stdout)

	// 4. 创建 Core
	// 开发环境通常把级别设为 Debug，生产环境设为 Info
	level := zap.InfoLevel
	if isDev {
		level = zap.DebugLevel
	}
	core := zapcore.NewCore(encoder, consoleSyncer, level)

	// 5. 生成 Logger
	// zap.Development() 会改变一些行为，比如触发 DPanic
	if isDev {
		Log = zap.New(core, zap.Development())
	} else {
		Log = zap.New(core)
	}

	SugarLog = Log.Sugar()
}
