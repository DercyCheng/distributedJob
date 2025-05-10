package repository

import (
	"distributedJob/internal/model/entity"
	"gorm.io/gorm"
)

// RoleRepository MySQL实现的角色存储库
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository 创建角色存储库实例
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// GetAllRoles 获取所有角色
func (r *RoleRepository) GetAllRoles() ([]*entity.Role, error) {
	var roles []*entity.Role
	result := r.db.Find(&roles)
	return roles, result.Error
}

// GetRoleByID 根据ID获取角色
func (r *RoleRepository) GetRoleByID(id int64) (*entity.Role, error) {
	var role entity.Role
	result := r.db.First(&role, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &role, nil
}

// GetRolesByKeyword 根据关键字搜索角色
func (r *RoleRepository) GetRolesByKeyword(keyword string, page, size int) ([]*entity.Role, int64, error) {
	var roles []*entity.Role
	var total int64

	query := r.db.Model(&entity.Role{})
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	result := query.Offset(offset).Limit(size).Find(&roles)

	return roles, total, result.Error
}

// CreateRole 创建角色
func (r *RoleRepository) CreateRole(role *entity.Role) (int64, error) {
	result := r.db.Create(role)
	if result.Error != nil {
		return 0, result.Error
	}
	return role.ID, nil
}

// UpdateRole 更新角色
func (r *RoleRepository) UpdateRole(role *entity.Role) error {
	return r.db.Save(role).Error
}

// DeleteRole 删除角色
func (r *RoleRepository) DeleteRole(id int64) error {
	// 开始事务
	tx := r.db.Begin()

	// 删除角色-权限关联
	if err := tx.Where("role_id = ?", id).Delete(&entity.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 删除角色
	if err := tx.Delete(&entity.Role{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	return tx.Commit().Error
}

// GetRolePermissions 获取角色的权限列表
func (r *RoleRepository) GetRolePermissions(roleID int64) ([]*entity.Permission, error) {
	var permissions []*entity.Permission

	// 通过连接查询获取角色的权限
	result := r.db.Table("permission").
		Joins("JOIN role_permission ON permission.id = role_permission.permission_id").
		Where("role_permission.role_id = ?", roleID).
		Find(&permissions)

	return permissions, result.Error
}

// SetRolePermissions 设置角色权限
func (r *RoleRepository) SetRolePermissions(roleID int64, permissionIDs []int64) error {
	// 开始事务
	tx := r.db.Begin()

	// 删除现有的角色-权限关联
	if err := tx.Where("role_id = ?", roleID).Delete(&entity.RolePermission{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 创建新的角色-权限关联
	for _, permissionID := range permissionIDs {
		rolePermission := &entity.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		if err := tx.Create(rolePermission).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	return tx.Commit().Error
}
