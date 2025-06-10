package department

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
)

type DepartmentService struct {
	grpc.UnimplementedDepartmentServiceServer
	db *gorm.DB
}

func NewDepartmentService(db *gorm.DB) *DepartmentService {
	return &DepartmentService{
		db: db,
	}
}

func (s *DepartmentService) CreateDepartment(ctx context.Context, req *grpc.CreateDepartmentRequest) (*grpc.CreateDepartmentResponse, error) {
	// 检查部门代码是否已存在
	var existingDept models.Department
	err := s.db.Where("code = ?", req.GetCode()).First(&existingDept).Error
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "department code already exists")
	}

	// 如果有父部门，检查父部门是否存在
	if req.GetParentId() != "" {
		var parentDept models.Department
		err := s.db.Where("id = ?", req.GetParentId()).First(&parentDept).Error
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "parent department not found")
		}
	}

	// 创建部门
	dept := models.Department{
		ID:          uuid.New().String(),
		Name:        req.GetName(),
		Code:        req.GetCode(),
		Description: req.GetDescription(),
		Status:      models.DeptStatusActive,
		Sort:        int(req.GetSort()),
	}

	if req.GetParentId() != "" {
		dept.ParentID = &req.ParentId
	}

	err = s.db.Create(&dept).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create department: %v", err)
	}

	return &grpc.CreateDepartmentResponse{
		Department: s.modelToGrpcDepartment(&dept),
	}, nil
}

func (s *DepartmentService) GetDepartment(ctx context.Context, req *grpc.GetDepartmentRequest) (*grpc.GetDepartmentResponse, error) {
	var dept models.Department
	err := s.db.Preload("Parent").Preload("Children").
		Where("id = ?", req.GetId()).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "department not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get department: %v", err)
	}

	return &grpc.GetDepartmentResponse{
		Department: s.modelToGrpcDepartment(&dept),
	}, nil
}

func (s *DepartmentService) ListDepartments(ctx context.Context, req *grpc.ListDepartmentsRequest) (*grpc.ListDepartmentsResponse, error) {
	var departments []models.Department
	query := s.db.Preload("Parent").Preload("Children")

	// 添加过滤条件
	if req.GetKeyword() != "" {
		keyword := "%" + req.GetKeyword() + "%"
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if req.GetStatus() != "" {
		query = query.Where("status = ?", req.GetStatus())
	}

	// 分页
	page := req.GetPage()
	if page <= 0 {
		page = 1
	}
	size := req.GetSize()
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 获取总数
	var total int64
	countQuery := s.db.Model(&models.Department{})
	if req.GetKeyword() != "" {
		keyword := "%" + req.GetKeyword() + "%"
		countQuery = countQuery.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", keyword, keyword, keyword)
	}
	if req.GetStatus() != "" {
		countQuery = countQuery.Where("status = ?", req.GetStatus())
	}
	countQuery.Count(&total)

	err := query.Offset(int(offset)).Limit(int(size)).Order("sort ASC, name ASC").Find(&departments).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list departments: %v", err)
	}

	// 转换为gRPC格式
	var grpcDepartments []*grpc.Department
	for _, dept := range departments {
		grpcDepartments = append(grpcDepartments, s.modelToGrpcDepartment(&dept))
	}

	return &grpc.ListDepartmentsResponse{
		Departments: grpcDepartments,
		Total:       total,
	}, nil
}

func (s *DepartmentService) UpdateDepartment(ctx context.Context, req *grpc.UpdateDepartmentRequest) (*grpc.UpdateDepartmentResponse, error) {
	var dept models.Department
	err := s.db.Where("id = ?", req.GetId()).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "department not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find department: %v", err)
	}

	// 检查部门代码是否被其他部门使用
	if req.GetCode() != dept.Code {
		var existingDept models.Department
		err := s.db.Where("code = ? AND id != ?", req.GetCode(), req.GetId()).First(&existingDept).Error
		if err == nil {
			return nil, status.Errorf(codes.AlreadyExists, "department code already exists")
		}
	}

	// 检查父部门循环引用
	if req.GetParentId() != "" && req.GetParentId() != req.GetId() {
		if s.hasCircularReference(req.GetId(), req.GetParentId()) {
			return nil, status.Errorf(codes.InvalidArgument, "circular reference detected")
		}
	}

	// 更新部门信息
	updates := map[string]interface{}{
		"name":        req.GetName(),
		"code":        req.GetCode(),
		"description": req.GetDescription(),
		"sort":        int(req.GetSort()),
	}

	if req.GetParentId() != "" {
		updates["parent_id"] = req.GetParentId()
	} else {
		updates["parent_id"] = nil
	}

	err = s.db.Model(&dept).Updates(updates).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update department: %v", err)
	}

	// 重新加载部门信息
	s.db.Preload("Parent").Preload("Children").Where("id = ?", req.GetId()).First(&dept)

	return &grpc.UpdateDepartmentResponse{
		Department: s.modelToGrpcDepartment(&dept),
	}, nil
}

func (s *DepartmentService) DeleteDepartment(ctx context.Context, req *grpc.DeleteDepartmentRequest) (*grpc.DeleteDepartmentResponse, error) {
	var dept models.Department
	err := s.db.Where("id = ?", req.GetId()).First(&dept).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "department not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find department: %v", err)
	}

	// 检查是否有子部门
	var childCount int64
	s.db.Model(&models.Department{}).Where("parent_id = ?", req.GetId()).Count(&childCount)
	if childCount > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete department with sub-departments")
	}

	// 检查是否有用户
	var userCount int64
	s.db.Model(&models.User{}).Where("department_id = ?", req.GetId()).Count(&userCount)
	if userCount > 0 {
		return nil, status.Errorf(codes.FailedPrecondition, "cannot delete department with users")
	}

	// 软删除部门
	err = s.db.Delete(&dept).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete department: %v", err)
	}

	return &grpc.DeleteDepartmentResponse{
		Success: true,
	}, nil
}

func (s *DepartmentService) GetDepartmentTree(ctx context.Context, req *grpc.GetDepartmentTreeRequest) (*grpc.GetDepartmentTreeResponse, error) {
	// 获取所有部门
	var departments []models.Department
	err := s.db.Where("status = ?", models.DeptStatusActive).Order("sort ASC, name ASC").Find(&departments).Error
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get departments: %v", err)
	}

	// 构建树形结构
	tree := s.buildDepartmentTree(departments, nil)

	return &grpc.GetDepartmentTreeResponse{
		Departments: tree,
	}, nil
}

// 检查循环引用
func (s *DepartmentService) hasCircularReference(deptID, parentID string) bool {
	var parent models.Department
	err := s.db.Where("id = ?", parentID).First(&parent).Error
	if err != nil {
		return false
	}

	if parent.ParentID == nil {
		return false
	}

	if *parent.ParentID == deptID {
		return true
	}

	return s.hasCircularReference(deptID, *parent.ParentID)
}

// 构建部门树
func (s *DepartmentService) buildDepartmentTree(departments []models.Department, parentID *string) []*grpc.Department {
	var tree []*grpc.Department

	for _, dept := range departments {
		// 检查是否为当前层级的部门
		if (parentID == nil && dept.ParentID == nil) || (parentID != nil && dept.ParentID != nil && *dept.ParentID == *parentID) {
			grpcDept := s.modelToGrpcDepartment(&dept)
			// 递归获取子部门
			grpcDept.Children = s.buildDepartmentTree(departments, &dept.ID)
			tree = append(tree, grpcDept)
		}
	}

	return tree
}

// 模型转换
func (s *DepartmentService) modelToGrpcDepartment(dept *models.Department) *grpc.Department {
	grpcDept := &grpc.Department{
		Id:          dept.ID,
		Name:        dept.Name,
		Code:        dept.Code,
		Description: dept.Description,
		Status:      string(dept.Status),
		Sort:        int32(dept.Sort),
		CreatedAt:   timestamppb.New(dept.CreatedAt),
		UpdatedAt:   timestamppb.New(dept.UpdatedAt),
	}

	if dept.ParentID != nil {
		grpcDept.ParentId = *dept.ParentID
	}

	// 添加父部门信息
	if dept.Parent != nil {
		grpcDept.Parent = &grpc.Department{
			Id:   dept.Parent.ID,
			Name: dept.Parent.Name,
			Code: dept.Parent.Code,
		}
	}

	// 添加子部门信息
	for _, child := range dept.Children {
		grpcDept.Children = append(grpcDept.Children, &grpc.Department{
			Id:   child.ID,
			Name: child.Name,
			Code: child.Code,
		})
	}

	return grpcDept
}
