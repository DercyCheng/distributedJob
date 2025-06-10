package auth

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	grpc "go-job/api/grpc"
	"go-job/internal/models"
	"go-job/pkg/auth"
)

type AuthService struct {
	grpc.UnimplementedAuthServiceServer
	db         *gorm.DB
	jwtManager *auth.JWTManager
}

func NewAuthService(db *gorm.DB, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (s *AuthService) Login(ctx context.Context, req *grpc.LoginRequest) (*grpc.LoginResponse, error) {
	// 查找用户
	var user models.User
	err := s.db.Preload("Department").Preload("Roles.Permissions").
		Where("username = ? AND status = ?", req.GetUsername(), models.UserStatusActive).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword()))
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password")
	}

	// 收集角色和权限
	var roles []string
	var permissions []string
	permissionSet := make(map[string]bool)

	for _, role := range user.Roles {
		if role.Status == models.RoleStatusActive {
			roles = append(roles, role.Code)
			for _, permission := range role.Permissions {
				if permission.Status == models.PermStatusActive && !permissionSet[permission.Code] {
					permissions = append(permissions, permission.Code)
					permissionSet[permission.Code] = true
				}
			}
		}
	}

	// 生成访问令牌
	accessToken, err := s.jwtManager.Generate(user.ID, user.Username, user.DepartmentID, roles, permissions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	// 生成刷新令牌（这里简化处理，实际应该存储到Redis等）
	refreshToken, err := s.jwtManager.Generate(user.ID, user.Username, user.DepartmentID, roles, permissions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// 转换用户信息
	grpcUser := s.modelToGrpcUser(&user)

	// 转换权限信息
	var grpcPermissions []*grpc.Permission
	var allPermissions []models.Permission
	s.db.Where("code IN ?", permissions).Find(&allPermissions)
	for _, perm := range allPermissions {
		grpcPermissions = append(grpcPermissions, s.modelToGrpcPermission(&perm))
	}

	return &grpc.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1小时
		User:         grpcUser,
		Permissions:  grpcPermissions,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *grpc.LogoutRequest) (*grpc.LogoutResponse, error) {
	// 实际项目中应该将令牌加入黑名单或从Redis中删除
	return &grpc.LogoutResponse{
		Success: true,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *grpc.RefreshTokenRequest) (*grpc.RefreshTokenResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtManager.Verify(req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
	}

	// 生成新的访问令牌
	accessToken, err := s.jwtManager.Generate(claims.UserID, claims.Username, claims.DepartmentID, claims.Roles, claims.Permissions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token: %v", err)
	}

	// 生成新的刷新令牌
	refreshToken, err := s.jwtManager.Generate(claims.UserID, claims.Username, claims.DepartmentID, claims.Roles, claims.Permissions)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token: %v", err)
	}

	return &grpc.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
	}, nil
}

func (s *AuthService) GetUserInfo(ctx context.Context, req *grpc.GetUserInfoRequest) (*grpc.GetUserInfoResponse, error) {
	var user models.User
	err := s.db.Preload("Department").Preload("Roles.Permissions").
		Where("id = ? AND status = ?", req.GetUserId(), models.UserStatusActive).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	grpcUser := s.modelToGrpcUser(&user)
	return &grpc.GetUserInfoResponse{
		User: grpcUser,
	}, nil
}

func (s *AuthService) GetUserPermissions(ctx context.Context, req *grpc.GetUserPermissionsRequest) (*grpc.GetUserPermissionsResponse, error) {
	var user models.User
	err := s.db.Preload("Roles.Permissions").
		Where("id = ? AND status = ?", req.GetUserId(), models.UserStatusActive).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to find user: %v", err)
	}

	var permissions []*grpc.Permission
	permissionSet := make(map[string]bool)

	for _, role := range user.Roles {
		if role.Status == models.RoleStatusActive {
			for _, permission := range role.Permissions {
				if permission.Status == models.PermStatusActive && !permissionSet[permission.ID] {
					permissions = append(permissions, s.modelToGrpcPermission(&permission))
					permissionSet[permission.ID] = true
				}
			}
		}
	}

	return &grpc.GetUserPermissionsResponse{
		Permissions: permissions,
	}, nil
}

func (s *AuthService) modelToGrpcUser(user *models.User) *grpc.User {
	grpcUser := &grpc.User{
		Id:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RealName: user.RealName,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Status:   string(user.Status),
	}

	if user.Department != nil {
		grpcUser.Department = s.modelToGrpcDepartment(user.Department)
	}

	for _, role := range user.Roles {
		grpcUser.Roles = append(grpcUser.Roles, s.modelToGrpcRole(&role))
	}

	return grpcUser
}

func (s *AuthService) modelToGrpcDepartment(dept *models.Department) *grpc.Department {
	return &grpc.Department{
		Id:          dept.ID,
		Name:        dept.Name,
		Code:        dept.Code,
		Description: dept.Description,
		Status:      string(dept.Status),
		Sort:        int32(dept.Sort),
	}
}

func (s *AuthService) modelToGrpcRole(role *models.Role) *grpc.Role {
	grpcRole := &grpc.Role{
		Id:          role.ID,
		Name:        role.Name,
		Code:        role.Code,
		Description: role.Description,
		Status:      string(role.Status),
	}

	for _, permission := range role.Permissions {
		grpcRole.Permissions = append(grpcRole.Permissions, s.modelToGrpcPermission(&permission))
	}

	return grpcRole
}

func (s *AuthService) modelToGrpcPermission(perm *models.Permission) *grpc.Permission {
	return &grpc.Permission{
		Id:       perm.ID,
		Name:     perm.Name,
		Code:     perm.Code,
		Type:     string(perm.Type),
		Resource: perm.Resource,
		Action:   perm.Action,
		Path:     perm.Path,
		Icon:     perm.Icon,
		Sort:     int32(perm.Sort),
		Status:   string(perm.Status),
	}
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}
