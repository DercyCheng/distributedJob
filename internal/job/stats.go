package job

// GetRunningJobCount 获取当前正在运行的任务数
func (s *Scheduler) GetRunningJobCount() int {
	if s.metrics != nil {
		// Since we can't do type assertion on a struct, we'll use the metrics
		// we already have directly
		// Try to get the active workers gauge directly from metrics
		count := s.metrics.GetActiveWorkers()
		if count >= 0 {
			return int(count)
		}
	}

	// 如果没有指标或无法获取，则返回近似值
	// 这个是一个简单的实现，实际应该有更准确的计数机制
	return len(s.runningJobs)
}
