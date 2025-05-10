package middleware

import (
	"distributedJob/pkg/metrics"
	"distributedJob/pkg/tracing"

	"github.com/gin-gonic/gin"
)

// InstrumentationMiddleware 组合了跟踪和指标的中间件
func InstrumentationMiddleware(tracer *tracing.Tracer, metrics *metrics.Metrics) gin.HandlerFunc {
	tracingMiddleware := TracingMiddleware(tracer)
	metricsMiddleware := MetricsMiddleware(metrics)

	return func(c *gin.Context) {
		// 先应用跟踪中间件
		tracingMiddleware(c)

		// 如果链路已经中断，则不继续执行
		if c.IsAborted() {
			return
		}

		// 再应用指标中间件
		metricsMiddleware(c)
	}
}
