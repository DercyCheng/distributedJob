package repository

import (
	"distributedJob/internal/model/entity"
	"gorm.io/gorm"
)

// PermissionRepository MySQL实现的权限存储库
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository 创建权限存储库实例
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// GetAllPermissions 获取所有权限
func (r *PermissionRepository) GetAllPermissions() ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	result := r.db.Find(&permissions)
	return permissions, result.Error
}

// GetPermissionByID 根据ID获取权限
func (r *PermissionRepository) GetPermissionByID(id int64) (*entity.Permission, error) {
	var permission entity.Permission
	result := r.db.First(&permission, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &permission, nil
}

// GetPermissionsByRoleID 获取角色的权限列表
func (r *PermissionRepository) GetPermissionsByRoleID(roleID int64) ([]*entity.Permission, error) {
	var permissions []*entity.Permission

	// 通过连接查询获取角色的权限
	result := r.db.Table("permission").
		Joins("JOIN role_permission ON permission.id = role_permission.permission_id").
		Where("role_permission.role_id = ? AND permission.status = 1", roleID).
		Find(&permissions)

	return permissions, result.Error
}
