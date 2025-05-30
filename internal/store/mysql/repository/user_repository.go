package repository

import (
	"distributedJob/internal/model/entity"

	"gorm.io/gorm"
)

// UserRepository MySQL实现的用户存储库
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户存储库实例
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id int64) (*entity.User, error) {
	var user entity.User
	result := r.db.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *UserRepository) GetUserByUsername(username string) (*entity.User, error) {
	var user entity.User
	// Use raw SQL to select by username without ORDER BY, leveraging username index
	sql := "SELECT id, username, password, real_name, email, phone, department_id, role_id, status, create_time, update_time " +
		"FROM `user` WHERE username = ? LIMIT 1"
	result := r.db.Raw(sql, username).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &user, nil
}

// GetUsersByDepartmentID 根据部门ID获取用户列表
func (r *UserRepository) GetUsersByDepartmentID(departmentID int64, page, size int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	// 查询总数
	if err := r.db.Model(&entity.User{}).Where("department_id = ?", departmentID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	result := r.db.Where("department_id = ?", departmentID).Offset(offset).Limit(size).Find(&users)

	return users, total, result.Error
}

// GetUsersByKeyword 根据关键字搜索用户
func (r *UserRepository) GetUsersByKeyword(keyword string, page, size int) ([]*entity.User, int64, error) {
	var users []*entity.User
	var total int64

	query := r.db.Model(&entity.User{})
	if keyword != "" {
		query = query.Where("username LIKE ? OR real_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * size
	result := query.Offset(offset).Limit(size).Find(&users)

	return users, total, result.Error
}

// CreateUser 创建用户
func (r *UserRepository) CreateUser(user *entity.User) (int64, error) {
	result := r.db.Create(user)
	if result.Error != nil {
		return 0, result.Error
	}
	return user.ID, nil
}

// UpdateUser 更新用户
func (r *UserRepository) UpdateUser(user *entity.User) error {
	// 不更新密码
	return r.db.Model(user).Omit("password").Updates(user).Error
}

// DeleteUser 删除用户
func (r *UserRepository) DeleteUser(id int64) error {
	return r.db.Delete(&entity.User{}, id).Error
}

// UpdateUserPassword 更新用户密码
func (r *UserRepository) UpdateUserPassword(id int64, password string) error {
	return r.db.Model(&entity.User{}).Where("id = ?", id).Update("password", password).Error
}
