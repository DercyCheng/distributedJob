package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	AuthorizationHeader = "authorization"
	AuthorizationBearer = "bearer"
)

type AuthInterceptor struct {
	jwtManager      *JWTManager
	accessibleRoles map[string][]string
	publicMethods   map[string]bool
}

func NewAuthInterceptor(jwtManager *JWTManager, accessibleRoles map[string][]string, publicMethods []string) *AuthInterceptor {
	publicMethodsMap := make(map[string]bool)
	for _, method := range publicMethods {
		publicMethodsMap[method] = true
	}

	return &AuthInterceptor{
		jwtManager:      jwtManager,
		accessibleRoles: accessibleRoles,
		publicMethods:   publicMethodsMap,
	}
}

func (interceptor *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 检查是否为公开方法
		if interceptor.publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// 验证访问令牌
		claims, err := interceptor.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		// 将用户信息添加到上下文
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "department_id", claims.DepartmentID)
		ctx = context.WithValue(ctx, "roles", claims.Roles)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		return handler(ctx, req)
	}
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) (*Claims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md[AuthorizationHeader]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	accessToken := values[0]
	if !strings.HasPrefix(strings.ToLower(accessToken), AuthorizationBearer) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization token format")
	}

	accessToken = accessToken[len(AuthorizationBearer)+1:]
	claims, err := interceptor.jwtManager.Verify(accessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	// 检查角色权限
	if accessibleRoles, ok := interceptor.accessibleRoles[method]; ok {
		hasAccess := false
		for _, role := range accessibleRoles {
			if interceptor.jwtManager.HasRole(claims, role) {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			return nil, status.Errorf(codes.PermissionDenied, "no permission to access this RPC")
		}
	}

	return claims, nil
}

// 权限检查中间件
func RequirePermission(permission string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		permissions, ok := ctx.Value("permissions").([]string)
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "permissions not found in context")
		}

		hasPermission := false
		for _, p := range permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return nil, status.Errorf(codes.PermissionDenied, "permission %s required", permission)
		}

		return handler(ctx, req)
	}
}

// 部门权限检查
func RequireDepartmentAccess() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		userDeptID, ok := ctx.Value("department_id").(string)
		if !ok {
			return nil, status.Errorf(codes.PermissionDenied, "department_id not found in context")
		}

		// 这里可以添加具体的部门权限逻辑
		// 例如：检查请求的资源是否属于用户所在部门
		_ = userDeptID // 暂时忽略未使用的变量

		return handler(ctx, req)
	}
}

// 从上下文获取用户信息的辅助函数
func GetUserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

func GetUsernameFromContext(ctx context.Context) (string, error) {
	username, ok := ctx.Value("username").(string)
	if !ok {
		return "", fmt.Errorf("username not found in context")
	}
	return username, nil
}

func GetDepartmentIDFromContext(ctx context.Context) (string, error) {
	departmentID, ok := ctx.Value("department_id").(string)
	if !ok {
		return "", fmt.Errorf("department_id not found in context")
	}
	return departmentID, nil
}

func GetRolesFromContext(ctx context.Context) ([]string, error) {
	roles, ok := ctx.Value("roles").([]string)
	if !ok {
		return nil, fmt.Errorf("roles not found in context")
	}
	return roles, nil
}

func GetPermissionsFromContext(ctx context.Context) ([]string, error) {
	permissions, ok := ctx.Value("permissions").([]string)
	if !ok {
		return nil, fmt.Errorf("permissions not found in context")
	}
	return permissions, nil
}
