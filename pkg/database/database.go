package database

import (
	"fmt"
	"go-job/internal/models"
	"go-job/pkg/config"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.Config) error {
	dsn := cfg.GetMySQLDSN()

	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层数据库连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(cfg.Database.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MySQL.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.MySQL.ConnMaxLifetime * time.Second)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	// 自动迁移
	if err := migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 初始化默认数据
	if err := InitDefaultData(); err != nil {
		return fmt.Errorf("初始化默认数据失败: %w", err)
	}

	return nil
}

// migrate 自动迁移数据库表
func migrate() error {
	return db.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.Role{},
		&models.Permission{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.Job{},
		&models.JobExecution{},
		&models.Worker{},
		&models.JobSchedule{},
		&models.AISchedule{},
	)
}

// InitDefaultData 初始化默认数据
func InitDefaultData() error {
	// 检查是否已经初始化过
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return fmt.Errorf("检查用户数据失败: %w", err)
	}

	if count > 0 {
		// 已有数据，跳过初始化
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 创建默认部门
		rootDeptID := uuid.New().String()
		rootDept := &models.Department{
			ID:          rootDeptID,
			Name:        "总公司",
			Code:        "ROOT",
			Description: "根部门",
			Status:      models.DeptStatusActive,
			Sort:        0,
		}
		if err := tx.Create(rootDept).Error; err != nil {
			return fmt.Errorf("创建根部门失败: %w", err)
		}

		itDeptID := uuid.New().String()
		itDept := &models.Department{
			ID:          itDeptID,
			Name:        "信息技术部",
			Code:        "IT",
			Description: "负责系统开发和运维",
			ParentID:    &rootDeptID,
			Status:      models.DeptStatusActive,
			Sort:        1,
		}
		if err := tx.Create(itDept).Error; err != nil {
			return fmt.Errorf("创建IT部门失败: %w", err)
		}

		// 创建默认权限
		permissions := []models.Permission{
			{
				ID:       uuid.New().String(),
				Name:     "系统管理",
				Code:     "system:admin",
				Type:     models.PermTypeMenu,
				Resource: "system",
				Action:   "admin",
				Path:     "/system",
				Icon:     "system",
				Sort:     1,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "任务管理",
				Code:     "job:manage",
				Type:     models.PermTypeMenu,
				Resource: "job",
				Action:   "manage",
				Path:     "/jobs",
				Icon:     "job",
				Sort:     2,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "创建任务",
				Code:     "job:create",
				Type:     models.PermTypeButton,
				Resource: "job",
				Action:   "create",
				Sort:     1,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "编辑任务",
				Code:     "job:update",
				Type:     models.PermTypeButton,
				Resource: "job",
				Action:   "update",
				Sort:     2,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "删除任务",
				Code:     "job:delete",
				Type:     models.PermTypeButton,
				Resource: "job",
				Action:   "delete",
				Sort:     3,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "执行记录",
				Code:     "execution:view",
				Type:     models.PermTypeMenu,
				Resource: "execution",
				Action:   "view",
				Path:     "/executions",
				Icon:     "execution",
				Sort:     3,
				Status:   models.PermStatusActive,
			},
			{
				ID:       uuid.New().String(),
				Name:     "工作节点",
				Code:     "worker:view",
				Type:     models.PermTypeMenu,
				Resource: "worker",
				Action:   "view",
				Path:     "/workers",
				Icon:     "worker",
				Sort:     4,
				Status:   models.PermStatusActive,
			},
		}

		for _, perm := range permissions {
			if err := tx.Create(&perm).Error; err != nil {
				return fmt.Errorf("创建权限失败: %w", err)
			}
		}

		// 设置权限层级关系
		jobManagePermID := permissions[1].ID
		permissions[2].ParentID = &jobManagePermID // job:create
		permissions[3].ParentID = &jobManagePermID // job:update
		permissions[4].ParentID = &jobManagePermID // job:delete

		// 更新权限的父子关系
		for i := 2; i <= 4; i++ {
			if err := tx.Save(&permissions[i]).Error; err != nil {
				return fmt.Errorf("更新权限父子关系失败: %w", err)
			}
		}

		// 创建默认角色
		adminRoleID := uuid.New().String()
		adminRole := &models.Role{
			ID:          adminRoleID,
			Name:        "系统管理员",
			Code:        "admin",
			Description: "拥有所有权限的系统管理员",
			Status:      models.RoleStatusActive,
		}
		if err := tx.Create(adminRole).Error; err != nil {
			return fmt.Errorf("创建管理员角色失败: %w", err)
		}

		operatorRoleID := uuid.New().String()
		operatorRole := &models.Role{
			ID:          operatorRoleID,
			Name:        "任务操作员",
			Code:        "operator",
			Description: "可以管理任务和查看执行记录",
			Status:      models.RoleStatusActive,
		}
		if err := tx.Create(operatorRole).Error; err != nil {
			return fmt.Errorf("创建操作员角色失败: %w", err)
		}

		viewerRoleID := uuid.New().String()
		viewerRole := &models.Role{
			ID:          viewerRoleID,
			Name:        "只读用户",
			Code:        "viewer",
			Description: "只能查看任务和执行记录",
			Status:      models.RoleStatusActive,
		}
		if err := tx.Create(viewerRole).Error; err != nil {
			return fmt.Errorf("创建查看者角色失败: %w", err)
		}

		// 为管理员角色分配所有权限
		for _, perm := range permissions {
			rolePermission := &models.RolePermission{
				ID:           uuid.New().String(),
				RoleID:       adminRoleID,
				PermissionID: perm.ID,
			}
			if err := tx.Create(rolePermission).Error; err != nil {
				return fmt.Errorf("分配管理员权限失败: %w", err)
			}
		}

		// 为操作员角色分配部分权限
		operatorPermCodes := []string{"job:manage", "job:create", "job:update", "execution:view", "worker:view"}
		for _, permCode := range operatorPermCodes {
			for _, perm := range permissions {
				if perm.Code == permCode {
					rolePermission := &models.RolePermission{
						ID:           uuid.New().String(),
						RoleID:       operatorRoleID,
						PermissionID: perm.ID,
					}
					if err := tx.Create(rolePermission).Error; err != nil {
						return fmt.Errorf("分配操作员权限失败: %w", err)
					}
					break
				}
			}
		}

		// 为查看者角色分配查看权限
		viewerPermCodes := []string{"job:manage", "execution:view", "worker:view"}
		for _, permCode := range viewerPermCodes {
			for _, perm := range permissions {
				if perm.Code == permCode {
					rolePermission := &models.RolePermission{
						ID:           uuid.New().String(),
						RoleID:       viewerRoleID,
						PermissionID: perm.ID,
					}
					if err := tx.Create(rolePermission).Error; err != nil {
						return fmt.Errorf("分配查看者权限失败: %w", err)
					}
					break
				}
			}
		}

		// 创建默认管理员用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("密码加密失败: %w", err)
		}

		adminUserID := uuid.New().String()
		adminUser := &models.User{
			ID:           adminUserID,
			Username:     "admin",
			Email:        "admin@example.com",
			Password:     string(hashedPassword),
			RealName:     "系统管理员",
			Phone:        "13800138000",
			Status:       models.UserStatusActive,
			DepartmentID: itDeptID,
		}
		if err := tx.Create(adminUser).Error; err != nil {
			return fmt.Errorf("创建管理员用户失败: %w", err)
		}

		// 为管理员用户分配管理员角色
		userRole := &models.UserRole{
			ID:     uuid.New().String(),
			UserID: adminUserID,
			RoleID: adminRoleID,
		}
		if err := tx.Create(userRole).Error; err != nil {
			return fmt.Errorf("分配用户角色失败: %w", err)
		}

		// 创建演示任务操作员
		operatorUser := &models.User{
			ID:           uuid.New().String(),
			Username:     "operator",
			Email:        "operator@example.com",
			Password:     string(hashedPassword),
			RealName:     "任务操作员",
			Phone:        "13800138001",
			Status:       models.UserStatusActive,
			DepartmentID: itDeptID,
		}
		if err := tx.Create(operatorUser).Error; err != nil {
			return fmt.Errorf("创建操作员用户失败: %w", err)
		}

		operatorUserRole := &models.UserRole{
			ID:     uuid.New().String(),
			UserID: operatorUser.ID,
			RoleID: operatorRoleID,
		}
		if err := tx.Create(operatorUserRole).Error; err != nil {
			return fmt.Errorf("分配操作员角色失败: %w", err)
		}

		return nil
	})
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return db
}

// Close 关闭数据库连接
func Close() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// Transaction 执行事务
func Transaction(fn func(tx *gorm.DB) error) error {
	return db.Transaction(fn)
}

// IsConnected 检查数据库连接状态
func IsConnected() bool {
	if db == nil {
		return false
	}

	sqlDB, err := db.DB()
	if err != nil {
		return false
	}

	return sqlDB.Ping() == nil
}
