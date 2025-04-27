package repository

import (
	"github.com/distributedJob/internal/model/entity"
	"gorm.io/gorm"
)

// DepartmentRepository MySQL实现的部门存储库
type DepartmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository 创建部门存储库实例
func NewDepartmentRepository(db *gorm.DB) *DepartmentRepository {
	return &DepartmentRepository{db: db}
}

// GetAllDepartments 获取所有部门
func (r *DepartmentRepository) GetAllDepartments() ([]*entity.Department, error) {
	var departments []*entity.Department
	result := r.db.Find(&departments)
	return departments, result.Error
}

// GetDepartmentByID 根据ID获取部门
func (r *DepartmentRepository) GetDepartmentByID(id int64) (*entity.Department, error) {
	var department entity.Department
	result := r.db.First(&department, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &department, nil
}

// GetDepartmentsByKeyword 根据关键字搜索部门
func (r *DepartmentRepository) GetDepartmentsByKeyword(keyword string) ([]*entity.Department, error) {
	var departments []*entity.Department

	query := r.db.Where("status = 1")
	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	result := query.Find(&departments)
	return departments, result.Error
}

// CreateDepartment 创建部门
func (r *DepartmentRepository) CreateDepartment(department *entity.Department) (int64, error) {
	result := r.db.Create(department)
	if result.Error != nil {
		return 0, result.Error
	}
	return department.ID, nil
}

// UpdateDepartment 更新部门
func (r *DepartmentRepository) UpdateDepartment(department *entity.Department) error {
	return r.db.Save(department).Error
}

// DeleteDepartment 删除部门
func (r *DepartmentRepository) DeleteDepartment(id int64) error {
	// 判断是否有子部门
	var count int64
	r.db.Model(&entity.Department{}).Where("parent_id = ?", id).Count(&count)
	if count > 0 {
		return gorm.ErrInvalidTransaction // 有子部门不能删除
	}

	// 判断是否有关联的任务
	r.db.Model(&entity.Task{}).Where("department_id = ?", id).Count(&count)
	if count > 0 {
		return gorm.ErrInvalidTransaction // 有关联的任务不能删除
	}

	return r.db.Delete(&entity.Department{}, id).Error
}
