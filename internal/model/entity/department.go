package entity

import "time"

// Department 部门实体
type Department struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Description string    `json:"description" gorm:"type:varchar(500)"`
	ParentID    *int64    `json:"parentId" gorm:"column:parent_id"`
	Status      int8      `json:"status" gorm:"type:tinyint(4);not null;default:1"`
	CreateTime  time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// TableName 指定表名
func (Department) TableName() string {
	return "department"
}
