package repository

import (
	"fmt"
	"time"

	"distributedJob/internal/model/entity"
	"distributedJob/pkg/logger"

	"gorm.io/gorm"
)

// TaskRepository MySQL实现的任务存储库
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务存储库实例
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetAllTasks 获取所有任务
func (r *TaskRepository) GetAllTasks() ([]*entity.Task, error) {
	var tasks []*entity.Task
	result := r.db.Find(&tasks)
	return tasks, result.Error
}

// GetTaskByID 根据ID获取任务
func (r *TaskRepository) GetTaskByID(id int64) (*entity.Task, error) {
	var task entity.Task
	result := r.db.First(&task, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

// GetTasksByDepartmentID 根据部门ID获取任务列表
func (r *TaskRepository) GetTasksByDepartmentID(departmentID int64, page, size int) ([]*entity.Task, int64, error) {
	var tasks []*entity.Task
	var total int64

	// 查询总数
	if err := r.db.Model(&entity.Task{}).Where("department_id = ?", departmentID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	result := r.db.Where("department_id = ?", departmentID).Offset(offset).Limit(size).Find(&tasks)

	return tasks, total, result.Error
}

// CreateTask 创建任务
func (r *TaskRepository) CreateTask(task *entity.Task) (int64, error) {
	result := r.db.Create(task)
	if result.Error != nil {
		return 0, result.Error
	}
	return task.ID, nil
}

// UpdateTask 更新任务
func (r *TaskRepository) UpdateTask(task *entity.Task) error {
	return r.db.Save(task).Error
}

// DeleteTask 删除任务
func (r *TaskRepository) DeleteTask(id int64) error {
	return r.db.Delete(&entity.Task{}, id).Error
}

// UpdateTaskStatus 更新任务状态
func (r *TaskRepository) UpdateTaskStatus(id int64, status int8) error {
	return r.db.Model(&entity.Task{}).Where("id = ?", id).Update("status", status).Error
}

// SaveTaskRecord 保存任务执行记录
func (r *TaskRepository) SaveTaskRecord(record *entity.Record) error {
	// 获取表名，根据年月动态确定
	tableName := entity.GetTableNameByYearMonth(
		record.CreateTime.Year(),
		int(record.CreateTime.Month()),
	)

	// 检查表是否存在，不存在则创建
	if err := r.ensureRecordTableExists(tableName); err != nil {
		return err
	}

	// 设置表名
	result := r.db.Table(tableName).Create(record)
	return result.Error
}

// GetRecords 获取任务执行记录
func (r *TaskRepository) GetRecords(year, month int, taskID, departmentID *int64, success *int8, page, size int) ([]*entity.Record, int64, error) {
	var records []*entity.Record
	var total int64

	// 获取表名
	tableName := entity.GetTableNameByYearMonth(year, month)

	// 构建查询
	query := r.db.Table(tableName)

	// 添加过滤条件
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}
	if success != nil {
		query = query.Where("success = ?", *success)
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		// 表不存在时不报错，返回空列表
		if err.Error() == "Error 1146: Table doesn't exist" {
			return []*entity.Record{}, 0, nil
		}
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	result := query.Order("id DESC").Offset(offset).Limit(size).Find(&records)
	if result.Error != nil {
		// 表不存在时不报错，返回空列表
		if result.Error.Error() == "Error 1146: Table doesn't exist" {
			return []*entity.Record{}, 0, nil
		}
		return nil, 0, result.Error
	}

	return records, total, nil
}

// GetRecordByID 根据ID获取任务执行记录
func (r *TaskRepository) GetRecordByID(id int64, year, month int) (*entity.Record, error) {
	var record entity.Record

	// 获取表名
	tableName := entity.GetTableNameByYearMonth(year, month)

	// 查询记录
	result := r.db.Table(tableName).First(&record, id)
	if result.Error != nil {
		return nil, result.Error
	}

	return &record, nil
}

// GetRecordStats 获取任务执行统计
func (r *TaskRepository) GetRecordStats(year, month int, taskID, departmentID *int64) (map[string]interface{}, error) {
	// 获取表名
	tableName := entity.GetTableNameByYearMonth(year, month)

	// 检查表是否存在
	if !r.tableExists(tableName) {
		// 表不存在，返回空统计
		return map[string]interface{}{
			"totalCount":   0,
			"successCount": 0,
			"failCount":    0,
			"successRate":  0,
			"avgCostTime":  0,
			"dailyStats":   []interface{}{},
		}, nil
	}

	// 构建查询基础
	query := r.db.Table(tableName)

	// 添加过滤条件
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}

	// 统计总数
	var totalCount int64
	query.Count(&totalCount)

	// 统计成功数
	var successCount int64
	query.Where("success = 1").Count(&successCount)

	// 计算失败数
	failCount := totalCount - successCount

	// 计算成功率
	var successRate float64
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100
	}

	// 计算平均耗时
	var avgCostTime float64
	query.Select("AVG(cost_time)").Row().Scan(&avgCostTime)

	// 计算每日统计
	type DailyStat struct {
		Date         string `json:"date"`
		TotalCount   int64  `json:"totalCount"`
		SuccessCount int64  `json:"successCount"`
		FailCount    int64  `json:"failCount"`
	}

	var dailyStats []DailyStat

	// 获取所选月份的天数
	daysInMonth := 31 // 默认31天
	if month == 2 {
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			daysInMonth = 29 // 闰年2月
		} else {
			daysInMonth = 28 // 平年2月
		}
	} else if month == 4 || month == 6 || month == 9 || month == 11 {
		daysInMonth = 30
	}

	// 查询每日统计
	for day := 1; day <= daysInMonth; day++ {
		date := fmt.Sprintf("%d-%02d-%02d", year, month, day)
		dayStart := fmt.Sprintf("%s 00:00:00", date)
		dayEnd := fmt.Sprintf("%s 23:59:59", date)

		var dayTotalCount, daySuccessCount int64

		dayQuery := query.Where("create_time BETWEEN ? AND ?", dayStart, dayEnd)
		dayQuery.Count(&dayTotalCount)

		if dayTotalCount > 0 {
			dayQuery.Where("success = 1").Count(&daySuccessCount)

			dailyStats = append(dailyStats, DailyStat{
				Date:         date,
				TotalCount:   dayTotalCount,
				SuccessCount: daySuccessCount,
				FailCount:    dayTotalCount - daySuccessCount,
			})
		}
	}

	// 返回统计结果
	return map[string]interface{}{
		"totalCount":   totalCount,
		"successCount": successCount,
		"failCount":    failCount,
		"successRate":  successRate,
		"avgCostTime":  avgCostTime,
		"dailyStats":   dailyStats,
	}, nil
}

// GetRecordsByTimeRange retrieves records within a specific time range
func (r *TaskRepository) GetRecordsByTimeRange(year, month int, taskID, departmentID *int64, success *int8, page, size int, startTime, endTime time.Time) ([]*entity.Record, int64, error) {
	// Calculate table name based on year and month
	tableName := fmt.Sprintf("record_%04d_%02d", year, month)

	// Start building the query
	query := r.db.Table(tableName).Where("create_time BETWEEN ? AND ?", startTime, endTime)

	// Apply filters if provided
	if taskID != nil {
		query = query.Where("task_id = ?", *taskID)
	}

	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}

	if success != nil {
		query = query.Where("success = ?", *success)
	}

	// Count total records matching the criteria
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	offset := (page - 1) * size

	// Execute the query with pagination
	var records []*entity.Record
	if err := query.Order("id DESC").Offset(offset).Limit(size).Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// tableExists 检查表是否存在
func (r *TaskRepository) tableExists(tableName string) bool {
	var count int64
	r.db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?",
		tableName,
	).Scan(&count)
	return count > 0
}

// ensureRecordTableExists 确保记录表存在，不存在则创建
func (r *TaskRepository) ensureRecordTableExists(tableName string) error {
	// 检查表是否存在
	if r.tableExists(tableName) {
		return nil
	}

	// 创建表
	logger.Infof("Creating record table: %s", tableName)

	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
		  id bigint(20) NOT NULL AUTO_INCREMENT,
		  task_id bigint(20) NOT NULL COMMENT '任务ID',
		  task_name varchar(255) NOT NULL COMMENT '任务名称',
		  task_type varchar(20) NOT NULL DEFAULT 'HTTP' COMMENT '任务类型: HTTP、GRPC',
		  department_id bigint(20) NOT NULL COMMENT '所属部门ID',
		  url varchar(500) DEFAULT NULL COMMENT '调度URL',
		  http_method varchar(10) DEFAULT NULL COMMENT 'HTTP方法',
		  body text COMMENT '请求体',
		  headers text COMMENT '请求头',
		  grpc_service varchar(255) DEFAULT NULL COMMENT 'gRPC服务名',
		  grpc_method varchar(255) DEFAULT NULL COMMENT 'gRPC方法名',
		  grpc_params text COMMENT 'gRPC参数',
		  response text COMMENT '响应内容',
		  status_code int(11) DEFAULT NULL COMMENT 'HTTP状态码',
		  grpc_status int(11) DEFAULT NULL COMMENT 'gRPC状态码',
		  success tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否成功',
		  retry_times int(11) NOT NULL DEFAULT '0' COMMENT '重试次数',
		  use_fallback tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否使用了备用调用',
		  cost_time int(11) NOT NULL DEFAULT '0' COMMENT '耗时(毫秒)',
		  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		  PRIMARY KEY (id),
		  KEY idx_task_id (task_id),
		  KEY idx_department_id (department_id),
		  KEY idx_success (success),
		  KEY idx_create_time (create_time),
		  KEY idx_task_type (task_type)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行记录表'
	`, tableName)

	return r.db.Exec(sql).Error
}
