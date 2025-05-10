package logger

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
)

// LogWithContext 通过上下文记录日志，包含追踪ID
func LogWithContext(ctx context.Context, level string, format string, args ...interface{}) {
	// 提取追踪信息
	spanCtx := trace.SpanContextFromContext(ctx)

	// 构建日志前缀，如果有追踪ID则添加
	prefix := ""
	if spanCtx.IsValid() {
		prefix = fmt.Sprintf("[trace_id=%s] ", spanCtx.TraceID().String())
	}

	// 完整日志消息
	message := prefix + fmt.Sprintf(format, args...)

	// 根据级别记录日志
	switch level {
	case "info":
		Info(message)
	case "debug":
		Debug(message)
	case "warn":
		Warn(message)
	case "error":
		Error(message)
	case "fatal":
		Fatal(message)
	default:
		Info(message)
	}
}

// InfoWithContext 用上下文记录信息日志
func InfoWithContext(ctx context.Context, format string, args ...interface{}) {
	LogWithContext(ctx, "info", format, args...)
}

// DebugWithContext 用上下文记录调试日志
func DebugWithContext(ctx context.Context, format string, args ...interface{}) {
	LogWithContext(ctx, "debug", format, args...)
}

// WarnWithContext 用上下文记录警告日志
func WarnWithContext(ctx context.Context, format string, args ...interface{}) {
	LogWithContext(ctx, "warn", format, args...)
}

// ErrorWithContext 用上下文记录错误日志
func ErrorWithContext(ctx context.Context, format string, args ...interface{}) {
	LogWithContext(ctx, "error", format, args...)
}

// FatalWithContext 用上下文记录致命错误日志
func FatalWithContext(ctx context.Context, format string, args ...interface{}) {
	LogWithContext(ctx, "fatal", format, args...)
}
