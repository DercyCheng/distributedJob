package store

import (
	"time"

	"distributedJob/internal/model/entity"
)

// TaskRepository 定义任务数据仓库接口
type TaskRepository interface {
	GetAllTasks() ([]*entity.Task, error)
	GetTaskByID(id int64) (*entity.Task, error)
	GetTasksByDepartmentID(departmentID int64, page, size int) ([]*entity.Task, int64, error)
	CreateTask(task *entity.Task) (int64, error)
	UpdateTask(task *entity.Task) error
	DeleteTask(id int64) error
	UpdateTaskStatus(id int64, status int8) error
	SaveTaskRecord(record *entity.Record) error
	GetRecords(year, month int, taskID, departmentID *int64, success *int8, page, size int) ([]*entity.Record, int64, error)
	GetRecordByID(id int64, year, month int) (*entity.Record, error)
	GetRecordStats(year, month int, taskID, departmentID *int64) (map[string]interface{}, error)
	GetRecordsByTimeRange(year, month int, taskID, departmentID *int64, success *int8, page, size int, startTime, endTime time.Time) ([]*entity.Record, int64, error)
}

// DepartmentRepository 定义部门数据仓库接口
type DepartmentRepository interface {
	GetAllDepartments() ([]*entity.Department, error)
	GetDepartmentByID(id int64) (*entity.Department, error)
	GetDepartmentsByKeyword(keyword string) ([]*entity.Department, error)
	CreateDepartment(department *entity.Department) (int64, error)
	UpdateDepartment(department *entity.Department) error
	DeleteDepartment(id int64) error
}

// UserRepository 定义用户数据仓库接口
type UserRepository interface {
	GetUserByID(id int64) (*entity.User, error)
	GetUserByUsername(username string) (*entity.User, error)
	GetUsersByDepartmentID(departmentID int64, page, size int) ([]*entity.User, int64, error)
	GetUsersByKeyword(keyword string, page, size int) ([]*entity.User, int64, error)
	CreateUser(user *entity.User) (int64, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id int64) error
	UpdateUserPassword(id int64, password string) error
}

// RoleRepository 定义角色数据仓库接口
type RoleRepository interface {
	GetAllRoles() ([]*entity.Role, error)
	GetRoleByID(id int64) (*entity.Role, error)
	GetRolesByKeyword(keyword string, page, size int) ([]*entity.Role, int64, error)
	CreateRole(role *entity.Role) (int64, error)
	UpdateRole(role *entity.Role) error
	DeleteRole(id int64) error
	GetRolePermissions(roleID int64) ([]*entity.Permission, error)
	SetRolePermissions(roleID int64, permissionIDs []int64) error
}

// PermissionRepository 定义权限数据仓库接口
type PermissionRepository interface {
	GetAllPermissions() ([]*entity.Permission, error)
	GetPermissionByID(id int64) (*entity.Permission, error)
	GetPermissionsByRoleID(roleID int64) ([]*entity.Permission, error)
}

// RepositoryManager 存储库管理器接口
type RepositoryManager interface {
	Task() TaskRepository
	Department() DepartmentRepository
	User() UserRepository
	Role() RoleRepository
	Permission() PermissionRepository
	Ping() error
}
