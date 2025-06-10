package user

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
)

type UserService struct {
	grpc.UnimplementedUserServiceServer
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *grpc.CreateUserRequest) (*grpc.CreateUserResponse, error) {
	// 检查用户名是否已存在
	var existingUser models.User
	err := s.db.Where("username = ?", req.GetUsername()).First(&existingUser).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}

	// 检查邮箱是否已存在
	if req.GetEmail() != "" {
		err := s.db.Where("email = ?", req.GetEmail()).First(&existingUser).Error
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "email already exists")
		}
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// 创建用户
	user := models.User{
		ID:           uuid.New().String(),
		Username:     req.GetUsername(),
		Email:        req.GetEmail(),
		Password:     string(hashedPassword),
		RealName:     req.GetRealName(),
		Phone:        req.GetPhone(),
		DepartmentID: req.GetDepartmentId(),
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.db.Create(&user).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &grpc.CreateUserResponse{
		User: s.modelToGrpcUser(&user),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *grpc.GetUserRequest) (*grpc.GetUserResponse, error) {
	var user models.User
	err := s.db.Preload("Department").Preload("Roles").
		Where("id = ?", req.GetId()).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &grpc.GetUserResponse{
		User: s.modelToGrpcUser(&user),
	}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *grpc.ListUsersRequest) (*grpc.ListUsersResponse, error) {
	var users []models.User
	query := s.db.Preload("Department").Preload("Roles")

	// 应用过滤器
	if req.GetDepartmentId() != "" {
		query = query.Where("department_id = ?", req.GetDepartmentId())
	}
	if req.GetStatus() != "" {
		query = query.Where("status = ?", req.GetStatus())
	}
	if req.GetKeyword() != "" {
		keyword := "%" + req.GetKeyword() + "%"
		query = query.Where("username LIKE ? OR real_name LIKE ? OR email LIKE ?", keyword, keyword, keyword)
	}

	// 分页
	var total int64
	query.Model(&models.User{}).Count(&total)

	// 计算偏移量
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	size := req.GetSize()
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	if size > 0 {
		query = query.Limit(int(size))
	}
	if offset > 0 {
		query = query.Offset(int(offset))
	}

	err := query.Find(&users).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	var grpcUsers []*grpc.User
	for _, user := range users {
		grpcUsers = append(grpcUsers, s.modelToGrpcUser(&user))
	}

	return &grpc.ListUsersResponse{
		Users: grpcUsers,
		Total: total,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *grpc.UpdateUserRequest) (*grpc.UpdateUserResponse, error) {
	var user models.User
	err := s.db.Where("id = ?", req.GetId()).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	// 检查邮箱是否被其他用户使用
	if req.GetEmail() != "" && req.GetEmail() != user.Email {
		var existingUser models.User
		err := s.db.Where("email = ? AND id != ?", req.GetEmail(), req.GetId()).First(&existingUser).Error
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "email already exists")
		}
	}

	// 更新用户信息
	updates := map[string]interface{}{
		"email":         req.GetEmail(),
		"real_name":     req.GetRealName(),
		"phone":         req.GetPhone(),
		"avatar":        req.GetAvatar(),
		"department_id": req.GetDepartmentId(),
		"status":        req.GetStatus(),
		"updated_at":    time.Now(),
	}

	err = s.db.Model(&user).Updates(updates).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// 更新用户角色
	if len(req.GetRoleIds()) > 0 {
		// 删除现有角色关联
		s.db.Where("user_id = ?", req.GetId()).Delete(&models.UserRole{})

		// 创建新的角色关联
		for _, roleID := range req.GetRoleIds() {
			userRole := models.UserRole{
				UserID: req.GetId(),
				RoleID: roleID,
			}
			s.db.Create(&userRole)
		}
	}

	// 重新加载用户信息
	s.db.Preload("Department").Preload("Roles").Where("id = ?", req.GetId()).First(&user)

	return &grpc.UpdateUserResponse{
		User: s.modelToGrpcUser(&user),
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *grpc.DeleteUserRequest) (*grpc.DeleteUserResponse, error) {
	var user models.User
	err := s.db.Where("id = ?", req.GetId()).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	// 软删除用户
	err = s.db.Delete(&user).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &grpc.DeleteUserResponse{
		Success: true,
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, req *grpc.ChangePasswordRequest) (*grpc.ChangePasswordResponse, error) {
	var user models.User
	err := s.db.Where("id = ?", req.GetUserId()).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	// 验证旧密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetOldPassword()))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "old password is incorrect")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetNewPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %v", err)
	}

	// 更新密码
	err = s.db.Model(&user).Update("password", string(hashedPassword)).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update password: %v", err)
	}

	return &grpc.ChangePasswordResponse{
		Success: true,
	}, nil
}

// 模型转换
func (s *UserService) modelToGrpcUser(user *models.User) *grpc.User {
	grpcUser := &grpc.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		RealName:  user.RealName,
		Phone:     user.Phone,
		Avatar:    user.Avatar,
		Status:    string(user.Status),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}

	if user.DepartmentID != "" {
		grpcUser.DepartmentId = user.DepartmentID
	}

	// 添加部门信息
	if user.Department != nil {
		grpcUser.Department = &grpc.Department{
			Id:   user.Department.ID,
			Name: user.Department.Name,
			Code: user.Department.Code,
		}
	}

	// 添加角色信息
	for _, role := range user.Roles {
		grpcUser.Roles = append(grpcUser.Roles, &grpc.Role{
			Id:   role.ID,
			Name: role.Name,
			Code: role.Code,
		})
	}

	return grpcUser
}

func (s *UserService) AssignUserRoles(ctx context.Context, req *grpc.AssignUserRolesRequest) (*grpc.AssignUserRolesResponse, error) {
	// 查找用户
	var user models.User
	err := s.db.First(&user, "id = ?", req.GetUserId()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 清除用户现有的角色关联
	if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserRole{}).Error; err != nil {
		tx.Rollback()
		return nil, status.Errorf(codes.Internal, "failed to clear user roles: %v", err)
	}

	// 添加新的角色关联
	for _, roleID := range req.GetRoleIds() {
		// 验证角色是否存在
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return nil, status.Errorf(codes.NotFound, "role not found: %s", roleID)
			}
			return nil, status.Errorf(codes.Internal, "failed to find role: %v", err)
		}

		// 创建用户角色关联
		userRole := &models.UserRole{
			UserID: user.ID,
			RoleID: roleID,
		}

		if err := tx.Create(userRole).Error; err != nil {
			tx.Rollback()
			return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, status.Errorf(codes.Internal, "failed to commit transaction: %v", err)
	}

	return &grpc.AssignUserRolesResponse{
		Success: true,
	}, nil
}
