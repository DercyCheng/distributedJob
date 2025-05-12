package entity

import (
	"strings"
	"time"
)

// Task 定时任务实体
type Task struct {
	ID                  int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string    `json:"name" gorm:"type:varchar(255);not null"`
	DepartmentID        int64     `json:"departmentId" gorm:"column:department_id;not null;index:idx_department_id"`
	TaskType            string    `json:"taskType" gorm:"column:task_type;type:varchar(20);not null;default:HTTP;index:idx_task_type"`
	Type                string    `json:"type" gorm:"column:type;type:varchar(20)"` // Adding Type field for backward compatibility
	Cron                string    `json:"cron" gorm:"column:cron_expression;type:varchar(100);not null"`
	URL                 string    `json:"url" gorm:"type:varchar(500)"`
	HTTPMethod          string    `json:"httpMethod" gorm:"column:http_method;type:varchar(10);default:GET"`
	Body                string    `json:"body" gorm:"type:text"`
	Headers             string    `json:"headers" gorm:"type:text"`
	GrpcService         string    `json:"grpcService" gorm:"column:grpc_service;type:varchar(255)"`
	GrpcMethod          string    `json:"grpcMethod" gorm:"column:grpc_method;type:varchar(255)"`
	GrpcParams          string    `json:"grpcParams" gorm:"column:grpc_params;type:text"`
	RetryCount          int       `json:"retryCount" gorm:"column:retry_count;type:int;not null;default:0"`
	RetryInterval       int       `json:"retryInterval" gorm:"column:retry_interval;type:int;not null;default:0"`
	Timeout             int       `json:"timeout" gorm:"column:timeout;type:int;not null;default:60"` // 超时时间，单位：秒
	FallbackURL         string    `json:"fallbackUrl" gorm:"column:fallback_url;type:varchar(500)"`
	FallbackGrpcService string    `json:"fallbackGrpcService" gorm:"column:fallback_grpc_service;type:varchar(255)"`
	FallbackGrpcMethod  string    `json:"fallbackGrpcMethod" gorm:"column:fallback_grpc_method;type:varchar(255)"`
	Status              int8      `json:"status" gorm:"type:tinyint(4);not null;default:1;index:idx_status"`
	CreateTime          time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime          time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	CreateBy            int64     `json:"createBy" gorm:"column:create_by;not null"`
	UpdateBy            *int64    `json:"updateBy" gorm:"column:update_by"`
	// Additional fields for task scheduling
	LastExecuteTime *time.Time `json:"lastExecuteTime" gorm:"-"`
	NextExecuteTime *time.Time `json:"nextExecuteTime" gorm:"-"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "task"
}

// SyncTypeFields 同步Type和TaskType字段，确保两者保持一致
func (t *Task) SyncTypeFields() {
	// 防止任一字段为nil导致的问题
	if t == nil {
		return
	}

	// 规范化类型值 - 确保类型始终是HTTP或GRPC（大写）
	if t.TaskType != "" {
		t.TaskType = normalizeTaskType(t.TaskType)
	}
	if t.Type != "" {
		t.Type = normalizeTaskType(t.Type)
	}

	// 如果Type字段为空，而TaskType有值，则Type使用TaskType的值
	if t.Type == "" && t.TaskType != "" {
		t.Type = t.TaskType
		return
	}

	// 如果TaskType字段为空，而Type有值，则TaskType使用Type的值
	if t.TaskType == "" && t.Type != "" {
		t.TaskType = t.Type
		return
	}

	// 如果两者都有值但不同，以TaskType为准
	if t.Type != t.TaskType {
		t.Type = t.TaskType
	}
}

// normalizeTaskType 规范化任务类型名称
func normalizeTaskType(taskType string) string {
	// 转换为大写
	upperType := strings.ToUpper(taskType)

	// 如果是HTTP或GRPC，直接返回
	if upperType == "HTTP" || upperType == "GRPC" {
		return upperType
	}

	// 处理可能的小写或混合大小写情况
	if strings.ToUpper(taskType) == "HTTP" {
		return "HTTP"
	} else if strings.ToUpper(taskType) == "GRPC" {
		return "GRPC"
	}

	// 默认返回原值
	return taskType
}
