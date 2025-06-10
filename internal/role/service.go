package role

import (
	"context"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
)

type RoleService struct {
	grpc.UnimplementedRoleServiceServer
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{
		db: db,
	}
}

func (s *RoleService) CreateRole(ctx context.Context, req *grpc.CreateRoleRequest) (*grpc.CreateRoleResponse, error) {
	// 检查角色代码是否已存在
	var existingRole models.Role
	err := s.db.Where("code = ?", req.GetCode()).First(&existingRole).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "role code already exists")
	}

	// 创建角色
	role := models.Role{
		ID:          uuid.New().String(),
		Name:        req.GetName(),
		Code:        req.GetCode(),
		Description: req.GetDescription(),
		Status:      models.RoleStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = s.db.Create(&role).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	return &grpc.CreateRoleResponse{
		Role: s.modelToGrpcRole(&role),
	}, nil
}

func (s *RoleService) GetRole(ctx context.Context, req *grpc.GetRoleRequest) (*grpc.GetRoleResponse, error) {
	var role models.Role
	err := s.db.Preload("Permissions").Where("id = ?", req.GetId()).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "role not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get role: %v", err)
	}

	return &grpc.GetRoleResponse{
		Role: s.modelToGrpcRole(&role),
	}, nil
}

func (s *RoleService) ListRoles(ctx context.Context, req *grpc.ListRolesRequest) (*grpc.ListRolesResponse, error) {
	var roles []models.Role
	query := s.db.Preload("Permissions")

	// 应用过滤器
	if req.GetStatus() != "" {
		query = query.Where("status = ?", req.GetStatus())
	}
	if req.GetKeyword() != "" {
		keyword := "%" + req.GetKeyword() + "%"
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}

	// 分页
	var total int64
	query.Model(&models.Role{}).Count(&total)

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

	err := query.Find(&roles).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	var grpcRoles []*grpc.Role
	for _, role := range roles {
		grpcRoles = append(grpcRoles, s.modelToGrpcRole(&role))
	}

	return &grpc.ListRolesResponse{
		Roles: grpcRoles,
		Total: total,
	}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *grpc.UpdateRoleRequest) (*grpc.UpdateRoleResponse, error) {
	var role models.Role
	err := s.db.Where("id = ?", req.GetId()).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "role not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find role: %v", err)
	}

	// 检查角色代码是否被其他角色使用
	if req.GetCode() != role.Code {
		var existingRole models.Role
		err := s.db.Where("code = ? AND id != ?", req.GetCode(), req.GetId()).First(&existingRole).Error
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "role code already exists")
		}
	}

	// 更新角色信息
	updates := map[string]interface{}{
		"name":        req.GetName(),
		"code":        req.GetCode(),
		"description": req.GetDescription(),
		"updated_at":  time.Now(),
	}

	err = s.db.Model(&role).Updates(updates).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	// 重新加载角色信息
	s.db.Preload("Permissions").Where("id = ?", req.GetId()).First(&role)

	return &grpc.UpdateRoleResponse{
		Role: s.modelToGrpcRole(&role),
	}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *grpc.DeleteRoleRequest) (*grpc.DeleteRoleResponse, error) {
	var role models.Role
	err := s.db.Where("id = ?", req.GetId()).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "role not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find role: %v", err)
	}

	// 检查是否有用户使用此角色
	var userRoleCount int64
	s.db.Model(&models.UserRole{}).Where("role_id = ?", req.GetId()).Count(&userRoleCount)
	if userRoleCount > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete role with assigned users")
	}

	// 删除角色权限关联
	s.db.Where("role_id = ?", req.GetId()).Delete(&models.RolePermission{})

	// 软删除角色
	err = s.db.Delete(&role).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	return &grpc.DeleteRoleResponse{
		Success: true,
	}, nil
}

func (s *RoleService) AssignPermissions(ctx context.Context, req *grpc.AssignPermissionsRequest) (*grpc.AssignPermissionsResponse, error) {
	// 检查角色是否存在
	var role models.Role
	err := s.db.Where("id = ?", req.GetRoleId()).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "role not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find role: %v", err)
	}

	// 检查权限是否存在
	var permissions []models.Permission
	err = s.db.Where("id IN ?", req.GetPermissionIds()).Find(&permissions).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find permissions: %v", err)
	}
	if len(permissions) != len(req.GetPermissionIds()) {
		return nil, status.Errorf(codes.NotFound, "some permissions not found")
	}

	// 删除现有权限关联
	s.db.Where("role_id = ?", req.GetRoleId()).Delete(&models.RolePermission{})

	// 创建新的权限关联
	for _, permissionID := range req.GetPermissionIds() {
		rolePermission := models.RolePermission{
			RoleID:       req.GetRoleId(),
			PermissionID: permissionID,
		}
		s.db.Create(&rolePermission)
	}

	return &grpc.AssignPermissionsResponse{
		Success: true,
	}, nil
}

// 模型转换
func (s *RoleService) modelToGrpcRole(role *models.Role) *grpc.Role {
	grpcRole := &grpc.Role{
		Id:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      string(role.Status),
		CreatedAt:   timestamppb.New(role.CreatedAt),
		UpdatedAt:   timestamppb.New(role.UpdatedAt),
	}

	// 添加权限信息
	for _, permission := range role.Permissions {
		grpcRole.Permissions = append(grpcRole.Permissions, &grpc.Permission{
			Id:   permission.ID,
			Name: permission.Name,
			Code: permission.Code,
		})
	}

	return grpcRole
}
