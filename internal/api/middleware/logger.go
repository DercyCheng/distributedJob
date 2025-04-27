package middleware

import (
	"time"

	"github.com/distributedJob/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 设置请求时间到上下文
		c.Set("requestTime", startTime)

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()

		// 执行时间
		latencyTime := endTime.Sub(startTime)

		// 请求方法
		reqMethod := c.Request.Method

		// 请求路由
		reqUri := c.Request.RequestURI

		// 状态码
		statusCode := c.Writer.Status()

		// 客户端IP
		clientIP := c.ClientIP()

		// 用户代理
		userAgent := c.Request.UserAgent()

		// 错误信息
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if errorMessage != "" {
			// 记录包含错误的日志
			logger.Errorf("| %3d | %13v | %15s | %s | %s | %s | %s",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
				userAgent,
				errorMessage,
			)
		} else {
			// 记录正常的访问日志
			logger.Infof("| %3d | %13v | %15s | %s | %s | %s",
				statusCode,
				latencyTime,
				clientIP,
				reqMethod,
				reqUri,
				userAgent,
			)
		}
	}
}
