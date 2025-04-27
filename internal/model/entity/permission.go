package entity

import "time"

// Permission 权限实体
type Permission struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null"`
	Code        string    `json:"code" gorm:"type:varchar(50);not null;uniqueIndex:idx_code"`
	Description string    `json:"description" gorm:"type:varchar(255)"`
	Status      int8      `json:"status" gorm:"type:tinyint(4);not null;default:1"`
	CreateTime  time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permission"
}
