package broadcaster

import (
	"context"
	"go-job/internal/models"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"go-job/pkg/websocket"
	"time"

	"gorm.io/gorm"
)

// StatsBroadcaster 统计数据广播器
type StatsBroadcaster struct {
	db     *gorm.DB
	wsHub  *websocket.Hub
	ticker *time.Ticker
	done   chan struct{}
}

// NewStatsBroadcaster 创建统计数据广播器
func NewStatsBroadcaster(wsHub *websocket.Hub) *StatsBroadcaster {
	return &StatsBroadcaster{
		db:    database.GetDB(),
		wsHub: wsHub,
		done:  make(chan struct{}),
	}
}

// Start 启动广播服务
func (sb *StatsBroadcaster) Start(ctx context.Context) {
	sb.ticker = time.NewTicker(10 * time.Second) // 每10秒广播一次
	defer sb.ticker.Stop()

	logger.Info("统计数据广播器已启动")

	for {
		select {
		case <-ctx.Done():
			return
		case <-sb.done:
			return
		case <-sb.ticker.C:
			sb.broadcastStats()
		}
	}
}

// Stop 停止广播服务
func (sb *StatsBroadcaster) Stop() {
	close(sb.done)
	logger.Info("统计数据广播器已停止")
}

// broadcastStats 广播统计数据
func (sb *StatsBroadcaster) broadcastStats() {
	stats := sb.collectStats()
	sb.wsHub.BroadcastStats(*stats)
}

// collectStats 收集统计数据
func (sb *StatsBroadcaster) collectStats() *websocket.StatsMessage {
	stats := &websocket.StatsMessage{}

	// 总任务数
	sb.db.Model(&models.Job{}).Count(&stats.TotalJobs)

	// 活跃任务数
	sb.db.Model(&models.Job{}).Where("enabled = ?", true).Count(&stats.ActiveJobs)

	// 在线工作节点数
	sb.db.Model(&models.Worker{}).Where("status = ?", models.WorkerStatusOnline).Count(&stats.OnlineWorkers)

	// 成功率计算
	var totalExecutions, successExecutions int64
	sb.db.Model(&models.JobExecution{}).Count(&totalExecutions)
	sb.db.Model(&models.JobExecution{}).Where("status = ?", models.ExecutionStatusSuccess).Count(&successExecutions)

	if totalExecutions > 0 {
		stats.SuccessRate = float64(successExecutions) / float64(totalExecutions) * 100
	}

	return stats
}

// BroadcastLog 广播日志消息
func (sb *StatsBroadcaster) BroadcastLog(level, service, message string, data map[string]interface{}) {
	sb.wsHub.BroadcastLog(level, service, message, data)
}
