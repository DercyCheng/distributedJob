package entity

import "time"

// Task 定时任务实体
type Task struct {
	ID                  int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string    `json:"name" gorm:"type:varchar(255);not null"`
	DepartmentID        int64     `json:"departmentId" gorm:"column:department_id;not null;index:idx_department_id"`
	TaskType            string    `json:"taskType" gorm:"column:task_type;type:varchar(20);not null;default:HTTP;index:idx_task_type"`
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
