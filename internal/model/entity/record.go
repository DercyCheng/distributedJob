package entity

import (
	"fmt"
	"time"
)

// Record 任务执行记录实体
// 注意：这个实体对应的表名是按年月分表的，例如 record_202501
type Record struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID       int64     `json:"taskId" gorm:"column:task_id;not null;index:idx_task_id"`
	TaskName     string    `json:"taskName" gorm:"column:task_name;type:varchar(255);not null"`
	TaskType     string    `json:"taskType" gorm:"column:task_type;type:varchar(20);not null;default:HTTP;index:idx_task_type"`
	DepartmentID int64     `json:"departmentId" gorm:"column:department_id;not null;index:idx_department_id"`
	URL          string    `json:"url" gorm:"type:varchar(500)"`
	HTTPMethod   string    `json:"httpMethod" gorm:"column:http_method;type:varchar(10)"`
	Body         string    `json:"body" gorm:"type:text"`
	Headers      string    `json:"headers" gorm:"type:text"`
	GrpcService  string    `json:"grpcService" gorm:"column:grpc_service;type:varchar(255)"`
	GrpcMethod   string    `json:"grpcMethod" gorm:"column:grpc_method;type:varchar(255)"`
	GrpcParams   string    `json:"grpcParams" gorm:"column:grpc_params;type:text"`
	Response     string    `json:"response" gorm:"type:text"`
	StatusCode   *int      `json:"statusCode" gorm:"column:status_code;type:int"`
	GrpcStatus   *int      `json:"grpcStatus" gorm:"column:grpc_status;type:int"`
	Success      int8      `json:"success" gorm:"type:tinyint(1);not null;default:0;index:idx_success"`
	RetryTimes   int       `json:"retryTimes" gorm:"column:retry_times;type:int;not null;default:0"`
	UseFallback  int8      `json:"useFallback" gorm:"column:use_fallback;type:tinyint(1);not null;default:0"`
	CostTime     int       `json:"costTime" gorm:"column:cost_time;type:int;not null;default:0"`
	CreateTime   time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_create_time"`
}

// TableName 返回表名（动态计算当前年月）
func (r Record) TableName() string {
	return "record_" + time.Now().Format("200601")
}

// GetTableNameByYearMonth 根据年月获取表名
func GetTableNameByYearMonth(year, month int) string {
	return fmt.Sprintf("record_%d%02d", year, month)
}
