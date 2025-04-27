package mysql

import (
	"fmt"
	"time"

	"github.com/distributedJob/internal/config"
	"github.com/distributedJob/internal/store"
	"github.com/distributedJob/internal/store/mysql/repository"
	"github.com/distributedJob/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// MySQLManager MySQL存储管理器实现
type MySQLManager struct {
	db             *gorm.DB
	taskRepo       store.TaskRepository
	departmentRepo store.DepartmentRepository
	userRepo       store.UserRepository
	roleRepo       store.RoleRepository
	permissionRepo store.PermissionRepository
}

// NewMySQLManager 创建一个新的MySQL存储管理器
func NewMySQLManager(cfg *config.Config) (store.RepositoryManager, error) {
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.URL,
		cfg.Database.Schema)

	// 配置GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		// 关闭外键约束（由应用层处理）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConn)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("Connected to MySQL database")

	// 创建并返回管理器实例
	manager := &MySQLManager{
		db: db,
	}

	// 初始化各个存储库
	manager.taskRepo = repository.NewTaskRepository(db)
	manager.departmentRepo = repository.NewDepartmentRepository(db)
	manager.userRepo = repository.NewUserRepository(db)
	manager.roleRepo = repository.NewRoleRepository(db)
	manager.permissionRepo = repository.NewPermissionRepository(db)

	return manager, nil
}

// Task 返回任务存储库
func (m *MySQLManager) Task() store.TaskRepository {
	return m.taskRepo
}

// Department 返回部门存储库
func (m *MySQLManager) Department() store.DepartmentRepository {
	return m.departmentRepo
}

// User 返回用户存储库
func (m *MySQLManager) User() store.UserRepository {
	return m.userRepo
}

// Role 返回角色存储库
func (m *MySQLManager) Role() store.RoleRepository {
	return m.roleRepo
}

// Permission 返回权限存储库
func (m *MySQLManager) Permission() store.PermissionRepository {
	return m.permissionRepo
}

// Close 关闭数据库连接
func (m *MySQLManager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
