package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	once   sync.Once
)

// Init 初始化日志
func Init(level, filename string, maxSize, maxBackups, maxAge int, compress bool) {
	once.Do(func() {
		// 设置日志级别
		var logLevel zapcore.Level
		switch level {
		case "debug":
			logLevel = zap.DebugLevel
		case "info":
			logLevel = zap.InfoLevel
		case "warn":
			logLevel = zap.WarnLevel
		case "error":
			logLevel = zap.ErrorLevel
		default:
			logLevel = zap.InfoLevel
		}

		// 设置日志编码配置
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// 设置日志输出
		var writeSyncer zapcore.WriteSyncer
		if filename != "" {
			// 使用lumberjack进行日志轮转
			ljWriter := &lumberjack.Logger{
				Filename:   filename,
				MaxSize:    maxSize,
				MaxBackups: maxBackups,
				MaxAge:     maxAge,
				Compress:   compress,
			}
			writeSyncer = zapcore.AddSync(ljWriter)
		} else {
			// 如果未指定文件名，则输出到控制台
			writeSyncer = zapcore.AddSync(os.Stdout)
		}

		// 创建核心
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			writeSyncer,
			logLevel,
		)

		// 创建Logger
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		sugar = logger.Sugar()

		Debug("Logger initialized")
	})
}

// Debug 输出Debug级别日志
func Debug(format string, args ...interface{}) {
	ensureLogger()
	if len(args) > 0 {
		sugar.Debugf(format, args...)
	} else {
		sugar.Debug(format)
	}
}

// Info 输出Info级别日志
func Info(format string, args ...interface{}) {
	ensureLogger()
	if len(args) > 0 {
		sugar.Infof(format, args...)
	} else {
		sugar.Info(format)
	}
}

// Warn 输出Warn级别日志
func Warn(format string, args ...interface{}) {
	ensureLogger()
	if len(args) > 0 {
		sugar.Warnf(format, args...)
	} else {
		sugar.Warn(format)
	}
}

// Error 输出Error级别日志
func Error(format string, args ...interface{}) {
	ensureLogger()
	if len(args) > 0 {
		sugar.Errorf(format, args...)
	} else {
		sugar.Error(format)
	}
}

// Fatal 输出Fatal级别日志并退出
func Fatal(format string, args ...interface{}) {
	ensureLogger()
	if len(args) > 0 {
		sugar.Fatalf(format, args...)
	} else {
		sugar.Fatal(format)
	}
}

// Debugf 格式化输出Debug级别日志
func Debugf(format string, args ...interface{}) {
	ensureLogger()
	sugar.Debugf(format, args...)
}

// Infof 格式化输出Info级别日志
func Infof(format string, args ...interface{}) {
	ensureLogger()
	sugar.Infof(format, args...)
}

// Warnf 格式化输出Warn级别日志
func Warnf(format string, args ...interface{}) {
	ensureLogger()
	sugar.Warnf(format, args...)
}

// Errorf 格式化输出Error级别日志
func Errorf(format string, args ...interface{}) {
	ensureLogger()
	sugar.Errorf(format, args...)
}

// Fatalf 格式化输出Fatal级别日志并退出
func Fatalf(format string, args ...interface{}) {
	ensureLogger()
	sugar.Fatalf(format, args...)
}

// Close 关闭日志
func Close() {
	if logger != nil {
		logger.Sync()
	}
}

// ensureLogger 确保日志初始化
func ensureLogger() {
	if logger == nil {
		fmt.Println("Logger not initialized, using default configuration")
		Init("info", "", 0, 0, 0, false)
	}
}
