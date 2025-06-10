package http

import (
	"net/http"
	"strings"
	"time"

	"go-job/pkg/auth"
	"go-job/pkg/config"

	"github.com/gin-gonic/gin"
)

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

// LoginHandler 登录处理
func loginHandler(c *gin.Context) {
	// TODO: 实现登录逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Login endpoint - to be implemented"})
}

// RefreshTokenHandler 刷新token处理
func refreshTokenHandler(c *gin.Context) {
	// TODO: 实现刷新token逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Refresh token endpoint - to be implemented"})
}

// LogoutHandler 登出处理
func logoutHandler(c *gin.Context) {
	// TODO: 实现登出逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
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

// UpdateProfileHandler 更新用户信息
func updateProfileHandler(c *gin.Context) {
	// TODO: 实现更新用户信息逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Update profile endpoint - to be implemented"})
}

// ChangePasswordHandler 修改密码
func changePasswordHandler(c *gin.Context) {
	// TODO: 实现修改密码逻辑
	c.JSON(http.StatusOK, gin.H{"message": "Change password endpoint - to be implemented"})
}
