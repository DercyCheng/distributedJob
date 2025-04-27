package entity

import "time"

// RolePermission 角色权限关联实体
type RolePermission struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	RoleID       int64     `json:"roleId" gorm:"column:role_id;not null;uniqueIndex:idx_role_permission,priority:1"`
	PermissionID int64     `json:"permissionId" gorm:"column:permission_id;not null;uniqueIndex:idx_role_permission,priority:2"`
	CreateTime   time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permission"
}
