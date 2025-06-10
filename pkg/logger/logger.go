package logger

import (
	"go-job/pkg/config"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Init 初始化日志
func Init(cfg *config.Config) error {
	log = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return err
	}
	log.SetLevel(level)

	// 设置输出
	var output io.Writer
	switch cfg.Logger.Output {
	case "stdout":
		output = os.Stdout
	case "stderr":
		output = os.Stderr
	default:
		// 文件输出
		file, err := os.OpenFile(cfg.Logger.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		output = file
	}
	log.SetOutput(output)

	// 设置格式
	switch cfg.Logger.Format {
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return nil
}

// GetLogger 获取日志实例
func GetLogger() *logrus.Logger {
	return log
}

// Debug 调试日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 格式化信息日志
func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 格式化错误日志
func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// WithField 添加字段
func WithField(key string, value interface{}) *logrus.Entry {
	return log.WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

// WithError 添加错误字段
func WithError(err error) *logrus.Entry {
	return log.WithError(err)
}
