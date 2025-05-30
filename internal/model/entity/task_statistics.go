package entity

// TaskStatistics 任务统计信息
type TaskStatistics struct {
	TaskCount        int                // 任务总数
	SuccessRate      float64            // 任务成功率
	AvgExecutionTime float64            // 平均执行时间(毫秒)
	ExecutionStats   map[string]float64 // 执行统计，可包含不同类型任务的统计数据
}
