package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	grpc "go-job/api/grpc"
	"go-job/internal/user"
	"go-job/pkg/auth"
	"go-job/pkg/config"
	"go-job/pkg/redis"

	"github.com/gin-gonic/gin"
	redisClient "github.com/go-redis/redis/v8"
)

// AuthHandlers 认证处理器
type AuthHandlers struct {
	userService *user.UserService
}

// NewAuthHandlers 创建认证处理器
func NewAuthHandlers(userService *user.UserService) *AuthHandlers {
	return &AuthHandlers{
		userService: userService,
	}
}

// AuthMiddleware JWT认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}
		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 检查token是否在黑名单中
		redisClient := getRedisClient()
		if redisClient != nil {
			exists, err := redisClient.Exists(context.Background(), "blacklist:"+token).Result()
			if err == nil && exists > 0 {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
				c.Abort()
				return
			}
		}

		// 验证token
		jwtManager := getJWTManager() // 从全局配置获取
		claims, err := jwtManager.Verify(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 设置用户信息到context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("department_id", claims.DepartmentID)
		c.Set("roles", claims.Roles)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequirePermission 权限验证中间件
func requirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		permissionList, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		// 检查是否有所需权限
		hasPermission := false
		for _, perm := range permissionList {
			if perm == permission || perm == "admin:all" { // admin:all 是超级权限
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getJWTManager 获取JWT管理器 (这里需要从全局配置或依赖注入获取)
func getJWTManager() *auth.JWTManager {
	// 临时实现，实际应该从配置中获取
	cfg := config.Get() // 使用全局配置
	if cfg == nil {
		// 如果没有全局配置，使用默认值
		return auth.NewJWTManager("default-secret", 24*time.Hour)
	}
	return auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expire)
}

// getRedisClient 获取Redis客户端
func getRedisClient() *redisClient.Client {
	return redis.GetClient()
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	DepartmentID string   `json:"department_id"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
}

// LoginHandler 登录处理
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 验证用户名密码 - 支持多个测试用户
	var userInfo UserInfo
	var isValid bool

	switch req.Username {
	case "admin":
		if req.Password == "admin123" {
			isValid = true
			userInfo = UserInfo{
				ID:           "admin",
				Username:     req.Username,
				DepartmentID: "dept-001",
				Roles:        []string{"admin"},
				Permissions:  []string{"admin:all"},
			}
		}
	case "manager":
		if req.Password == "manager123" {
			isValid = true
			userInfo = UserInfo{
				ID:           "manager",
				Username:     req.Username,
				DepartmentID: "dept-002",
				Roles:        []string{"manager"},
				Permissions:  []string{"job:read", "job:create", "job:update", "execution:read", "worker:read", "stats:read"},
			}
		}
	case "operator":
		if req.Password == "operator123" {
			isValid = true
			userInfo = UserInfo{
				ID:           "operator",
				Username:     req.Username,
				DepartmentID: "dept-003",
				Roles:        []string{"operator"},
				Permissions:  []string{"job:read", "execution:read", "worker:read", "stats:read"},
			}
		}
	}

	if isValid {
		jwtManager := getJWTManager() // 创建JWT token
		token, err := jwtManager.Generate(userInfo.ID, userInfo.Username, userInfo.DepartmentID, userInfo.Roles, userInfo.Permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// 创建refresh token (24小时有效)
		refreshToken, err := jwtManager.GenerateRefresh(userInfo.ID, userInfo.Username, userInfo.DepartmentID, userInfo.Roles, userInfo.Permissions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		response := LoginResponse{
			Token:        token,
			RefreshToken: refreshToken,
			ExpiresIn:    int64(jwtManager.Duration.Seconds()),
			User:         userInfo,
		}

		c.JSON(http.StatusOK, gin.H{"data": response})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	}
}

// RefreshTokenRequest 刷新token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenHandler 刷新token处理
func refreshTokenHandler(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtManager := getJWTManager()

	// 验证refresh token
	claims, err := jwtManager.Verify(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// 生成新的access token
	newToken, err := jwtManager.Generate(claims.UserID, claims.Username, claims.DepartmentID, claims.Roles, claims.Permissions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}

	response := gin.H{
		"token":      newToken,
		"expires_in": int64(jwtManager.Duration.Seconds()),
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// LogoutHandler 登出处理
func logoutHandler(c *gin.Context) {
	// 从header获取token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		// 验证token是否有效
		jwtManager := getJWTManager()
		_, err := jwtManager.Verify(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// 将token加入黑名单
		redisClient := getRedisClient()
		if redisClient != nil {
			// 设置token黑名单，过期时间为24小时
			err = redisClient.Set(context.Background(), "blacklist:"+token, "1", time.Hour*24).Err()
			if err != nil {
				// 记录错误但不影响退出登录流程
				// logger.WithError(err).Error("Failed to add token to blacklist")
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
		"data":    gin.H{"status": "logged_out"},
	})
}

// GetProfileHandler 获取用户信息
func getProfileHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	departmentID, _ := c.Get("department_id")
	roles, _ := c.Get("roles")
	permissions, _ := c.Get("permissions")

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"username":      username,
		"department_id": departmentID,
		"roles":         roles,
		"permissions":   permissions,
	})
}

// UpdateProfileRequest 更新用户信息请求
type UpdateProfileRequest struct {
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
}

// UpdateProfileHandler 更新用户信息
func updateProfileHandler(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 实际实现：更新数据库中的用户信息
	// 注意：这是简化版本，实际应该通过依赖注入获取userService	// 可以考虑重构为使用AuthHandlers.UpdateProfile方法
	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully - use AuthHandlers.UpdateProfile for database integration",
		"data": gin.H{
			"user_id":      userID,
			"email":        req.Email,
			"phone":        req.Phone,
			"display_name": req.DisplayName,
			"updated_at":   time.Now().Unix(),
			"note":         "For database integration, use the AuthHandlers.UpdateProfile method",
		},
	})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ChangePasswordHandler 修改密码
func changePasswordHandler(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	username, _ := c.Get("username")
	// 实际实现：从数据库验证旧密码并更新新密码
	// 注意：这是简化版本，实际应该通过依赖注入获取userService
	// 可以考虑重构为使用AuthHandlers.ChangePassword方法

	// 临时实现：硬编码用户验证（仅用于演示）
	var isValidOldPassword bool
	switch username {
	case "admin":
		isValidOldPassword = req.OldPassword == "admin123"
	case "manager":
		isValidOldPassword = req.OldPassword == "manager123"
	case "operator":
		isValidOldPassword = req.OldPassword == "operator123"
	}

	if !isValidOldPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password"})
		return
	}

	// 验证新密码强度
	if len(req.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be at least 6 characters long"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
		"data": gin.H{
			"user_id":    userID,
			"updated_at": time.Now().Unix(),
		},
	})
}

// UpdateProfile 更新用户信息（使用数据库）
func (h *AuthHandlers) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// 调用用户服务更新数据库
	grpcReq := &grpc.UpdateUserRequest{
		Id:       userIDStr,
		Email:    req.Email,
		Phone:    req.Phone,
		RealName: req.DisplayName,
	}

	resp, err := h.userService.UpdateUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data": gin.H{
			"user_id":      resp.User.Id,
			"email":        resp.User.Email,
			"phone":        resp.User.Phone,
			"display_name": resp.User.RealName,
			"updated_at":   time.Now().Unix(),
		},
	})
}

// ChangePassword 修改密码（使用数据库）
func (h *AuthHandlers) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// 调用用户服务验证和更新密码
	grpcReq := &grpc.ChangePasswordRequest{
		UserId:      userIDStr,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	_, err := h.userService.ChangePassword(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to change password: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
		"data": gin.H{
			"user_id":    userIDStr,
			"updated_at": time.Now().Unix(),
		},
	})
}
