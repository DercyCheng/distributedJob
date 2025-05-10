package handler

import (
	"net/http"
	"time"

	"distributedJob/internal/infrastructure"

	"github.com/gin-gonic/gin"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	infra *infrastructure.Infrastructure
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(infra *infrastructure.Infrastructure) *HealthHandler {
	return &HealthHandler{
		infra: infra,
	}
}

// HealthStatus 健康状态结构体
type HealthStatus struct {
	Status      string            `json:"status"`
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Timestamp   string            `json:"timestamp"`
	Components  map[string]Status `json:"components"`
}

// Status 组件状态
type Status struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Check 检查系统健康状态
func (h *HealthHandler) Check(c *gin.Context) {
	status := HealthStatus{
		Status:      "UP",
		Version:     "1.0.0",      // 从配置获取
		Environment: "production", // 从配置获取
		Timestamp:   time.Now().Format(time.RFC3339),
		Components:  make(map[string]Status),
	}

	// 检查数据库
	dbStatus := Status{Status: "UP"}
	if h.infra.DB != nil {
		if err := h.infra.DB.Ping(); err != nil {
			dbStatus.Status = "DOWN"
			dbStatus.Message = err.Error()
			status.Status = "DOWN"
		}
	} else {
		dbStatus.Status = "UNKNOWN"
		dbStatus.Message = "Database component is not initialized"
	}
	status.Components["database"] = dbStatus
	// 检查Redis
	redisStatus := Status{Status: "UP"}
	if h.infra.Redis != nil {
		if err := h.infra.Redis.Ping(c.Request.Context()); err != nil {
			redisStatus.Status = "DOWN"
			redisStatus.Message = err.Error()
			status.Status = "DOWN"
		}
	} else {
		redisStatus.Status = "UNKNOWN"
		redisStatus.Message = "Redis component is not initialized"
	}
	status.Components["redis"] = redisStatus

	// 检查Kafka
	kafkaStatus := Status{Status: "UP"}
	if h.infra.Kafka != nil {
		if h.infra.Kafka.GetProducer() == nil {
			kafkaStatus.Status = "DOWN"
			kafkaStatus.Message = "Kafka producer is not initialized"
		}
	} else {
		kafkaStatus.Status = "UNKNOWN"
		kafkaStatus.Message = "Kafka component is not initialized"
	}
	status.Components["kafka"] = kafkaStatus

	// 检查ETCD
	etcdStatus := Status{Status: "UP"}
	if h.infra.Etcd != nil {
		_, err := h.infra.Etcd.Get(c, "/health-check")
		if err != nil {
			// 如果是找不到键，这不算故障
			if err.Error() != "key not found: /health-check" {
				etcdStatus.Status = "DOWN"
				etcdStatus.Message = err.Error()
				status.Status = "DOWN"
			}
		}
	} else {
		etcdStatus.Status = "UNKNOWN"
		etcdStatus.Message = "ETCD component is not initialized"
	}
	status.Components["etcd"] = etcdStatus

	c.JSON(http.StatusOK, status)
}
