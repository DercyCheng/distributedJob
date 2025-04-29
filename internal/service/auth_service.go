package service

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/distributedJob/internal/model/entity"
	"github.com/distributedJob/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const (
	// User status constants
	UserStatusEnabled  int8 = 1
	UserStatusDisabled int8 = 2
)

// AuthService 认证服务接口
type AuthService interface {
	// 用户认证
	Login(username, password string) (string, error)
	RefreshToken(userID int64) (string, error)
	ValidateToken(tokenString string) (*Claims, error)

	// 用户管理
	GetUserList(departmentID int64, page, size int) ([]*entity.User, int64, error)
	GetUserByID(id int64) (*entity.User, error)
	CreateUser(user *entity.User) (int64, error)
	UpdateUser(user *entity.User) error
	DeleteUser(id int64) error

	// 角色管理
	GetRoleList(page, size int) ([]*entity.Role, int64, error)
	GetRoleByID(id int64) (*entity.Role, error)
	CreateRole(role *entity.Role) (int64, error)
	UpdateRole(role *entity.Role) error
	DeleteRole(id int64) error

	// 部门管理
	GetDepartmentList(page, size int) ([]*entity.Department, int64, error)
	GetDepartmentByID(id int64) (*entity.Department, error)
	CreateDepartment(dept *entity.Department) (int64, error)
	UpdateDepartment(dept *entity.Department) error
	DeleteDepartment(id int64) error

	// 权限管理
	GetPermissionList() ([]*entity.Permission, error)
	HasPermission(userID int64, permissionCode string) (bool, error)
}

// Claims 自定义JWT声明
type Claims struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	DepartmentID int64  `json:"department_id"`
	RoleID       int64  `json:"role_id"`
	jwt.StandardClaims
}

// authService 认证服务实现
type authService struct {
	userRepo       store.UserRepository
	roleRepo       store.RoleRepository
	deptRepo       store.DepartmentRepository
	permissionRepo store.PermissionRepository
	jwtSecret      []byte
	tokenExpire    time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo store.UserRepository,
	roleRepo store.RoleRepository,
	deptRepo store.DepartmentRepository,
	permissionRepo store.PermissionRepository,
	jwtSecret string,
	tokenExpire time.Duration,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		deptRepo:       deptRepo,
		permissionRepo: permissionRepo,
		jwtSecret:      []byte(jwtSecret),
		tokenExpire:    tokenExpire,
	}
}

// Login 用户登录
func (s *authService) Login(username, password string) (string, error) {
	// 查询用户
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != UserStatusEnabled {
		return "", errors.New("user is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid password")
	}

	// 生成Token
	claims := &Claims{
		UserID:       user.ID,
		Username:     user.Username,
		DepartmentID: user.DepartmentID,
		RoleID:       user.RoleID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenExpire).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// RefreshToken 刷新用户令牌
func (s *authService) RefreshToken(userID int64) (string, error) {
	// 查询用户
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}

	// 检查用户状态
	if user.Status != UserStatusEnabled {
		return "", errors.New("user is disabled")
	}

	// 生成新Token
	claims := &Claims{
		UserID:       user.ID,
		Username:     user.Username,
		DepartmentID: user.DepartmentID,
		RoleID:       user.RoleID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(s.tokenExpire).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken 验证Token
func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	// 验证Token
	if err != nil {
		return nil, err
	}

	// 验证Token是否有效
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// GetUserList 获取用户列表
func (s *authService) GetUserList(departmentID int64, page, size int) ([]*entity.User, int64, error) {
	return s.userRepo.GetUsersByDepartmentID(departmentID, page, size)
}

// GetUserByID 获取用户详情
func (s *authService) GetUserByID(id int64) (*entity.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 清除敏感信息
	user.Password = ""

	return user, nil
}

// CreateUser 创建用户
func (s *authService) CreateUser(user *entity.User) (int64, error) {
	// 检查用户名是否已存在
	existUser, err := s.userRepo.GetUserByUsername(user.Username)
	if err != nil {
		return 0, err
	}
	if existUser != nil {
		return 0, errors.New("username already exists")
	}

	// 检查部门是否存在
	dept, err := s.deptRepo.GetDepartmentByID(user.DepartmentID)
	if err != nil {
		return 0, err
	}
	if dept == nil {
		return 0, errors.New("department not found")
	}

	// 检查角色是否存在
	role, err := s.roleRepo.GetRoleByID(user.RoleID)
	if err != nil {
		return 0, err
	}
	if role == nil {
		return 0, errors.New("role not found")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = string(hashedPassword)

	// 创建用户
	return s.userRepo.CreateUser(user)
}

// UpdateUser 更新用户
func (s *authService) UpdateUser(user *entity.User) error {
	// 检查用户是否存在
	existUser, err := s.userRepo.GetUserByID(user.ID)
	if err != nil {
		return err
	}
	if existUser == nil {
		return errors.New("user not found")
	}

	// 如果更新用户名，检查是否与其他用户重复
	if user.Username != existUser.Username {
		otherUser, err := s.userRepo.GetUserByUsername(user.Username)
		if err != nil {
			return err
		}
		if otherUser != nil && otherUser.ID != user.ID {
			return errors.New("username already exists")
		}
	}

	// 检查部门是否存在
	if user.DepartmentID > 0 {
		dept, err := s.deptRepo.GetDepartmentByID(user.DepartmentID)
		if err != nil {
			return err
		}
		if dept == nil {
			return errors.New("department not found")
		}
	}

	// 检查角色是否存在
	if user.RoleID > 0 {
		role, err := s.roleRepo.GetRoleByID(user.RoleID)
		if err != nil {
			return err
		}
		if role == nil {
			return errors.New("role not found")
		}
	}

	// 如果更新密码，则加密
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// 不更新密码
		user.Password = existUser.Password
	}

	// 更新用户
	return s.userRepo.UpdateUser(user)
}

// DeleteUser 删除用户
func (s *authService) DeleteUser(id int64) error {
	// 检查用户是否存在
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 删除用户
	return s.userRepo.DeleteUser(id)
}

// GetRoleList 获取角色列表
func (s *authService) GetRoleList(page, size int) ([]*entity.Role, int64, error) {
	return s.roleRepo.GetRolesByKeyword("", page, size)
}

// GetRoleByID 获取角色详情
func (s *authService) GetRoleByID(id int64) (*entity.Role, error) {
	role, err := s.roleRepo.GetRoleByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	// 获取角色权限
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(id)
	if err != nil {
		return nil, err
	}

	role.Permissions = permissions
	return role, nil
}

// CreateRole 创建角色
func (s *authService) CreateRole(role *entity.Role) (int64, error) {
	// 检查角色名是否已存在
	existingRoles, _, err := s.roleRepo.GetRolesByKeyword(role.Name, 1, 1)
	if err != nil {
		return 0, err
	}
	if len(existingRoles) > 0 && existingRoles[0].Name == role.Name {
		return 0, errors.New("role name already exists")
	}

	// 创建角色
	id, err := s.roleRepo.CreateRole(role)
	if err != nil {
		return 0, err
	}

	// 设置角色权限
	if len(role.Permissions) > 0 {
		permissionIDs := make([]int64, len(role.Permissions))
		for i, p := range role.Permissions {
			permissionIDs[i] = p.ID
		}

		if err := s.roleRepo.SetRolePermissions(id, permissionIDs); err != nil {
			return id, err
		}
	}

	return id, nil
}

// UpdateRole 更新角色
func (s *authService) UpdateRole(role *entity.Role) error {
	// 检查角色是否存在
	existRole, err := s.roleRepo.GetRoleByID(role.ID)
	if err != nil {
		return err
	}
	if existRole == nil {
		return errors.New("role not found")
	}

	// 如果更新角色名，检查是否与其他角色重复
	if role.Name != existRole.Name {
		existingRoles, _, err := s.roleRepo.GetRolesByKeyword(role.Name, 1, 1)
		if err != nil {
			return err
		}
		if len(existingRoles) > 0 && existingRoles[0].Name == role.Name && existingRoles[0].ID != role.ID {
			return errors.New("role name already exists")
		}
	}

	// 更新角色
	if err := s.roleRepo.UpdateRole(role); err != nil {
		return err
	}

	// 更新角色权限
	if role.Permissions != nil {
		// 设置新权限
		if len(role.Permissions) > 0 {
			permissionIDs := make([]int64, len(role.Permissions))
			for i, p := range role.Permissions {
				permissionIDs[i] = p.ID
			}

			if err := s.roleRepo.SetRolePermissions(role.ID, permissionIDs); err != nil {
				return err
			}
		} else {
			// 清空权限
			if err := s.roleRepo.SetRolePermissions(role.ID, []int64{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteRole 删除角色
func (s *authService) DeleteRole(id int64) error {
	// 检查角色是否存在
	role, err := s.roleRepo.GetRoleByID(id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	// 检查角色是否有用户使用
	users, count, err := s.userRepo.GetUsersByKeyword("", 1, 1)
	if err != nil {
		return err
	}

	// 检查是否有用户使用该角色
	if count > 0 {
		for _, user := range users {
			if user.RoleID == id {
				return errors.New("role is in use by users")
			}
		}
	}

	// 创建空权限列表
	if err := s.roleRepo.SetRolePermissions(id, []int64{}); err != nil {
		return err
	}

	// 删除角色
	return s.roleRepo.DeleteRole(id)
}

// GetDepartmentList 获取部门列表
func (s *authService) GetDepartmentList(page, size int) ([]*entity.Department, int64, error) {
	depts, err := s.deptRepo.GetAllDepartments()
	if err != nil {
		return nil, 0, err
	}

	// 手动分页
	total := int64(len(depts))
	start := (page - 1) * size
	end := start + size
	if start >= int(total) {
		return []*entity.Department{}, total, nil
	}
	if end > int(total) {
		end = int(total)
	}

	return depts[start:end], total, nil
}

// GetDepartmentByID 获取部门详情
func (s *authService) GetDepartmentByID(id int64) (*entity.Department, error) {
	dept, err := s.deptRepo.GetDepartmentByID(id)
	if err != nil {
		return nil, err
	}
	if dept == nil {
		return nil, errors.New("department not found")
	}
	return dept, nil
}

// CreateDepartment 创建部门
func (s *authService) CreateDepartment(dept *entity.Department) (int64, error) {
	// 检查部门名是否已存在
	depts, err := s.deptRepo.GetDepartmentsByKeyword(dept.Name)
	if err != nil {
		return 0, err
	}

	for _, existDept := range depts {
		if existDept.Name == dept.Name {
			return 0, errors.New("department name already exists")
		}
	}

	// 创建部门
	return s.deptRepo.CreateDepartment(dept)
}

// UpdateDepartment 更新部门
func (s *authService) UpdateDepartment(dept *entity.Department) error {
	// 检查部门是否存在
	existDept, err := s.deptRepo.GetDepartmentByID(dept.ID)
	if err != nil {
		return err
	}
	if existDept == nil {
		return errors.New("department not found")
	}

	// 如果更新部门名，检查是否与其他部门重复
	if dept.Name != existDept.Name {
		depts, err := s.deptRepo.GetDepartmentsByKeyword(dept.Name)
		if err != nil {
			return err
		}

		for _, otherDept := range depts {
			if otherDept.Name == dept.Name && otherDept.ID != dept.ID {
				return errors.New("department name already exists")
			}
		}
	}

	// 更新部门
	return s.deptRepo.UpdateDepartment(dept)
}

// DeleteDepartment 删除部门
func (s *authService) DeleteDepartment(id int64) error {
	// 检查部门是否存在
	dept, err := s.deptRepo.GetDepartmentByID(id)
	if err != nil {
		return err
	}
	if dept == nil {
		return errors.New("department not found")
	}

	// 检查部门是否有用户使用
	_, count, err := s.userRepo.GetUsersByDepartmentID(id, 1, 1)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("department is in use by users")
	}

	// 删除部门
	return s.deptRepo.DeleteDepartment(id)
}

// GetPermissionList 获取权限列表
func (s *authService) GetPermissionList() ([]*entity.Permission, error) {
	return s.permissionRepo.GetAllPermissions()
}

// HasPermission 检查用户是否拥有指定权限
func (s *authService) HasPermission(userID int64, permissionCode string) (bool, error) {
	// 获取用户信息
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not found")
	}

	// 根据角色ID获取权限列表
	permissions, err := s.permissionRepo.GetPermissionsByRoleID(user.RoleID)
	if err != nil {
		return false, err
	}

	// 检查是否包含指定权限
	for _, p := range permissions {
		if p.Code == permissionCode {
			return true, nil
		}
	}

	return false, nil
}
