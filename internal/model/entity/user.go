package entity

import "time"

// User 用户实体
type User struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"type:varchar(50);not null;uniqueIndex:idx_username"`
	Password     string    `json:"password" gorm:"type:varchar(100);not null"`
	RealName     string    `json:"realName" gorm:"column:real_name;type:varchar(50);not null"`
	Email        string    `json:"email" gorm:"type:varchar(100)"`
	Phone        string    `json:"phone" gorm:"type:varchar(20)"`
	DepartmentID int64     `json:"departmentId" gorm:"column:department_id;not null"`
	RoleID       int64     `json:"roleId" gorm:"column:role_id;not null"`
	Status       int8      `json:"status" gorm:"type:tinyint(4);not null;default:1"`
	CreateTime   time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdateTime   time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}
