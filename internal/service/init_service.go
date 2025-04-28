package service

import (
	"github.com/distributedJob/internal/model/entity"
	"github.com/distributedJob/internal/store"
	"github.com/distributedJob/pkg/logger"
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