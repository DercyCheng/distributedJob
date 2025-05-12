package middleware

import (
	"distributedJob/pkg/metrics"
	"distributedJob/pkg/tracing"
	"time"

	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// TracingMiddleware 分布式追踪中间件
func TracingMiddleware(tracer *tracing.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tracer == nil {
			c.Next()
			return
		}

		// 从HTTP请求头中提取上下文
		carrier := propagation.HeaderCarrier(c.Request.Header)
		ctx := tracer.Extract(c.Request.Context(), carrier)

		// 创建新的span
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		spanName := c.Request.Method + " " + path
		ctx, span := tracer.StartSpanWithAttributes(
			ctx,
			spanName,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.path", path),
			attribute.String("http.client_ip", c.ClientIP()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)
		defer span.End()

		// 将追踪上下文保存到gin的上下文中
		c.Set("tracing_context", ctx)
		c.Set("span", span)

		// 继续处理请求
		c.Next()

		// 在处理完成后添加响应信息到span
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int("http.response_size", c.Writer.Size()),
		)

		// 如果有错误则记录
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				tracing.RecordError(ctx, err.Err)
			}
		}
	}
}

// MetricsMiddleware 指标监控中间件
func MetricsMiddleware(metrics *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		if metrics == nil {
			c.Next()
			return
		}

		// 开始时间
		startTime := time.Now()

		// 处理请求
		c.Next()

		// 测量请求持续时间
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// 增加请求计数
		metrics.IncrementCounter("requests_total",
			c.Request.Method,
			path,
			fmt.Sprint(c.Writer.Status()),
		)

		// 记录请求持续时间
		metrics.MeasureRequestDuration(c.Request.Method, path, startTime)

		// 记录响应大小
		metrics.AddToCounter("response_size_bytes", float64(c.Writer.Size()),
			c.Request.Method,
			path,
		)

		// 如果有错误，增加错误计数
		if len(c.Errors) > 0 {
			metrics.IncrementCounter("request_errors",
				c.Request.Method,
				path,
			)
		}
	}
}
