package entity

import "time"

// Role 角色实体
type Role struct {
	ID          int64         `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string        `json:"name" gorm:"type:varchar(50);not null;uniqueIndex:idx_name"`
	Description string        `json:"description" gorm:"type:varchar(255)"`
	Status      int8          `json:"status" gorm:"type:tinyint(4);not null;default:1"`
	CreateTime  time.Time     `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time     `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
	Permissions []*Permission `json:"permissions" gorm:"-"` // 不存储到数据库中，只用于传输
}

// TableName 指定表名
func (Role) TableName() string {
	return "role"
}
