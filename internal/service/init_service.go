package service

import (
	"time"

	"distributedJob/internal/config"
	"distributedJob/internal/job"
	"distributedJob/internal/model/entity"
	"distributedJob/internal/store"
	"distributedJob/internal/store/redis"
	"distributedJob/pkg/logger"
	"distributedJob/pkg/memory"
	"golang.org/x/crypto/bcrypt"
)

// InitializeSystem 系统初始化函数
func InitializeSystem(
	userRepo store.UserRepository,
	roleRepo store.RoleRepository,
	deptRepo store.DepartmentRepository,
	permissionRepo store.PermissionRepository,
) error {
	// 检查是否已有用户，如果没有则创建默认管理员用户
	_, total, err := userRepo.GetUsersByKeyword("", 1, 1)
	if err != nil {
		return err
	}

	if total == 0 {
		logger.Info("No users found in database, creating default admin user")

		// 创建默认部门
		defaultDept := &entity.Department{
			Name:        "管理部门",
			Description: "系统默认管理部门",
			Status:      1, // 启用
		}

		deptID, err := deptRepo.CreateDepartment(defaultDept)
		if err != nil {
			return err
		}

		// 获取所有权限
		allPerms, err := permissionRepo.GetAllPermissions()
		if err != nil {
			return err
		}

		// 创建管理员角色
		adminRole := &entity.Role{
			Name:        "系统管理员",
			Description: "拥有所有权限的系统管理员角色",
			Status:      1, // 启用
		}

		roleID, err := roleRepo.CreateRole(adminRole)
		if err != nil {
			return err
		}

		// 为管理员角色分配所有权限
		if len(allPerms) > 0 {
			permIDs := make([]int64, len(allPerms))
			for i, p := range allPerms {
				permIDs[i] = p.ID
			}

			if err := roleRepo.SetRolePermissions(roleID, permIDs); err != nil {
				return err
			}
		}

		// 创建管理员用户
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminUser := &entity.User{
			Username:     "admin",
			RealName:     "系统管理员",
			Password:     string(hashedPassword),
			Email:        "admin@example.com",
			Phone:        "13800000000",
			DepartmentID: deptID,
			RoleID:       roleID,
			Status:       1, // 启用
		}

		_, err = userRepo.CreateUser(adminUser)
		if err != nil {
			return err
		}

		logger.Info("Default admin user created successfully")
	}

	return nil
}

// InitServices 初始化所有服务
func InitServices(
	config *config.Config,
	scheduler *job.Scheduler,
	userRepo store.UserRepository,
	roleRepo store.RoleRepository,
	deptRepo store.DepartmentRepository,
	permissionRepo store.PermissionRepository,
	taskRepo store.TaskRepository,
) (AuthService, TaskService, *redis.Manager, store.TokenRevoker, error) {

	// 初始化Redis管理器（如果启用）
	var redisManager *redis.Manager
	var tokenRevoker store.TokenRevoker

	// 检查Redis URL是否配置，作为判断是否启用Redis的依据
	if config.Redis.URL != "" {
		var err error
		redisManager, err = redis.NewManager(config)
		if err != nil {
			logger.Errorf("Failed to initialize Redis: %v", err)
			// 继续执行，使用内存模式
		} else {
			logger.Info("Connected to Redis successfully")

			// 创建基于Redis的令牌撤销器
			if config.Auth.TokenRevocationStrategy == "redis" {
				tokenRevoker = redisManager.CreateTokenRevoker()
				logger.Info("Token revocation enabled using Redis")
			}
		}
	}

	// 如果Redis不可用或者配置为使用内存模式，使用内存模式的令牌撤销器
	if tokenRevoker == nil && config.Auth.TokenRevocationStrategy == "memory" {
		tokenRevoker = memory.NewMemoryTokenRevoker()
		logger.Info("Token revocation enabled using in-memory storage")
	}

	// 计算令牌过期时间
	accessTokenExpire := time.Duration(config.Auth.JwtExpireMinutes) * time.Minute
	refreshTokenExpire := time.Duration(config.Auth.JwtRefreshExpireDays) * 24 * time.Hour

	// 初始化认证服务
	authService := NewAuthService(
		userRepo,
		roleRepo,
		deptRepo,
		permissionRepo,
		config.Auth.JwtSecret,
		config.Auth.JwtRefreshSecret,
		accessTokenExpire,
		refreshTokenExpire,
		tokenRevoker,
	)

	// 初始化任务服务
	taskService := NewTaskService(taskRepo, scheduler)

	// 初始化系统数据
	if err := InitializeSystem(userRepo, roleRepo, deptRepo, permissionRepo); err != nil {
		return nil, nil, nil, nil, err
	}

	return authService, taskService, redisManager, tokenRevoker, nil
}
