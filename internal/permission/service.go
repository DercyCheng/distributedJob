package permission

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

type PermissionService struct {
	grpc.UnimplementedPermissionServiceServer
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{
		db: db,
	}
}

func (s *PermissionService) CreatePermission(ctx context.Context, req *grpc.CreatePermissionRequest) (*grpc.CreatePermissionResponse, error) {
	// 检查权限代码是否已存在
	var existingPermission models.Permission
	err := s.db.Where("code = ?", req.GetCode()).First(&existingPermission).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "permission code already exists")
	}

	// 如果有父权限，检查父权限是否存在
	if req.GetParentId() != "" {
		var parentPermission models.Permission
		err := s.db.Where("id = ?", req.GetParentId()).First(&parentPermission).Error
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "parent permission not found")
		}
	}

	// 创建权限
	permission := models.Permission{
		ID:        uuid.New().String(),
		Name:      req.GetName(),
		Code:      req.GetCode(),
		Type:      models.PermType(req.GetType()),
		Resource:  req.GetResource(),
		Action:    req.GetAction(),
		Sort:      int(req.GetSort()),
		Status:    models.PermStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.GetParentId() != "" {
		permission.ParentID = &req.ParentId
	}

	err = s.db.Create(&permission).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create permission: %v", err)
	}

	return &grpc.CreatePermissionResponse{
		Permission: s.modelToGrpcPermission(&permission),
	}, nil
}

func (s *PermissionService) GetPermission(ctx context.Context, req *grpc.GetPermissionRequest) (*grpc.GetPermissionResponse, error) {
	var permission models.Permission
	err := s.db.Preload("Parent").Preload("Children").
		Where("id = ?", req.GetId()).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "permission not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get permission: %v", err)
	}

	return &grpc.GetPermissionResponse{
		Permission: s.modelToGrpcPermission(&permission),
	}, nil
}

func (s *PermissionService) ListPermissions(ctx context.Context, req *grpc.ListPermissionsRequest) (*grpc.ListPermissionsResponse, error) {
	var permissions []models.Permission
	query := s.db.Preload("Parent").Preload("Children")

	// 应用过滤器
	if req.GetType() != "" {
		query = query.Where("type = ?", req.GetType())
	}
	if req.GetStatus() != "" {
		query = query.Where("status = ?", req.GetStatus())
	}
	if req.GetKeyword() != "" {
		keyword := "%" + req.GetKeyword() + "%"
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}

	// 排序
	query = query.Order("sort ASC, created_at ASC")

	// 分页
	var total int64
	query.Model(&models.Permission{}).Count(&total)

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

	err := query.Find(&permissions).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	var grpcPermissions []*grpc.Permission
	for _, permission := range permissions {
		grpcPermissions = append(grpcPermissions, s.modelToGrpcPermission(&permission))
	}

	return &grpc.ListPermissionsResponse{
		Permissions: grpcPermissions,
		Total:       total,
	}, nil
}

func (s *PermissionService) UpdatePermission(ctx context.Context, req *grpc.UpdatePermissionRequest) (*grpc.UpdatePermissionResponse, error) {
	var permission models.Permission
	err := s.db.Where("id = ?", req.GetId()).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "permission not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find permission: %v", err)
	}

	// 检查权限代码是否被其他权限使用
	if req.GetCode() != permission.Code {
		var existingPermission models.Permission
		err := s.db.Where("code = ? AND id != ?", req.GetCode(), req.GetId()).First(&existingPermission).Error
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "permission code already exists")
		}
	}

	// 检查父权限循环引用
	if req.GetParentId() != "" && req.GetParentId() != req.GetId() {
		if s.hasCircularReference(req.GetId(), req.GetParentId()) {
			return nil, status.Errorf(codes.InvalidArgument, "circular reference detected")
		}
	}

	// 更新权限信息
	updates := map[string]interface{}{
		"name":       req.GetName(),
		"code":       req.GetCode(),
		"type":       req.GetType(),
		"resource":   req.GetResource(),
		"action":     req.GetAction(),
		"path":       req.GetPath(),
		"icon":       req.GetIcon(),
		"sort":       int(req.GetSort()),
		"status":     req.GetStatus(),
		"updated_at": time.Now(),
	}

	if req.GetParentId() != "" {
		updates["parent_id"] = req.GetParentId()
	} else {
		updates["parent_id"] = nil
	}

	err = s.db.Model(&permission).Updates(updates).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update permission: %v", err)
	}

	// 重新加载权限信息
	s.db.Preload("Parent").Preload("Children").Where("id = ?", req.GetId()).First(&permission)

	return &grpc.UpdatePermissionResponse{
		Permission: s.modelToGrpcPermission(&permission),
	}, nil
}

func (s *PermissionService) DeletePermission(ctx context.Context, req *grpc.DeletePermissionRequest) (*grpc.DeletePermissionResponse, error) {
	var permission models.Permission
	err := s.db.Where("id = ?", req.GetId()).First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "permission not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find permission: %v", err)
	}

	// 检查是否有子权限
	var childCount int64
	s.db.Model(&models.Permission{}).Where("parent_id = ?", req.GetId()).Count(&childCount)
	if childCount > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete permission with children")
	}

	// 检查是否有角色使用此权限
	var rolePermissionCount int64
	s.db.Model(&models.RolePermission{}).Where("permission_id = ?", req.GetId()).Count(&rolePermissionCount)
	if rolePermissionCount > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete permission assigned to roles")
	}

	// 软删除权限
	err = s.db.Delete(&permission).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete permission: %v", err)
	}

	return &grpc.DeletePermissionResponse{
		Success: true,
	}, nil
}

func (s *PermissionService) GetPermissionTree(ctx context.Context, req *grpc.GetPermissionTreeRequest) (*grpc.GetPermissionTreeResponse, error) {
	var permissions []models.Permission
	err := s.db.Preload("Parent").Preload("Children").
		Order("sort ASC, created_at ASC").Find(&permissions).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get permissions: %v", err)
	}

	// 构建树形结构
	tree := s.buildPermissionTree(permissions, nil)

	return &grpc.GetPermissionTreeResponse{
		Permissions: tree,
	}, nil
}

// 检查循环引用
func (s *PermissionService) hasCircularReference(permissionID, parentID string) bool {
	if permissionID == parentID {
		return true
	}

	var parent models.Permission
	err := s.db.Where("id = ?", parentID).First(&parent).Error
	if err != nil {
		return false
	}

	if parent.ParentID != nil {
		return s.hasCircularReference(permissionID, *parent.ParentID)
	}

	return false
}

// 构建权限树
func (s *PermissionService) buildPermissionTree(permissions []models.Permission, parentID *string) []*grpc.Permission {
	var tree []*grpc.Permission

	for _, permission := range permissions {
		// 检查是否是当前层级的权限
		if (parentID == nil && permission.ParentID == nil) ||
			(parentID != nil && permission.ParentID != nil && *permission.ParentID == *parentID) {

			grpcPermission := s.modelToGrpcPermission(&permission)

			// 递归构建子权限
			children := s.buildPermissionTree(permissions, &permission.ID)
			grpcPermission.Children = children

			tree = append(tree, grpcPermission)
		}
	}

	return tree
}

// 模型转换
func (s *PermissionService) modelToGrpcPermission(permission *models.Permission) *grpc.Permission {
	grpcPermission := &grpc.Permission{
		Id:        permission.ID,
		Name:      permission.Name,
		Code:      permission.Code,
		Type:      string(permission.Type),
		Resource:  permission.Resource,
		Action:    permission.Action,
		Path:      permission.Path,
		Icon:      permission.Icon,
		Sort:      int32(permission.Sort),
		Status:    string(permission.Status),
		CreatedAt: timestamppb.New(permission.CreatedAt),
		UpdatedAt: timestamppb.New(permission.UpdatedAt),
	}

	if permission.ParentID != nil {
		grpcPermission.ParentId = *permission.ParentID
	}

	// 添加父权限信息
	if permission.Parent != nil {
		grpcPermission.Parent = &grpc.Permission{
			Id:   permission.Parent.ID,
			Name: permission.Parent.Name,
			Code: permission.Parent.Code,
		}
	}

	// 添加子权限信息
	for _, child := range permission.Children {
		grpcPermission.Children = append(grpcPermission.Children, &grpc.Permission{
			Id:   child.ID,
			Name: child.Name,
			Code: child.Code,
		})
	}

	return grpcPermission
}
