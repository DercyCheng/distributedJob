package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/distributedJob/internal/config"
	"github.com/gin-gonic/gin"
)

// JWTClaims 自定义JWT声明结构
type JWTClaims struct {
	UserID       int64    `json:"userId"`
	Username     string   `json:"username"`
	DepartmentID int64    `json:"departmentId"`
	RoleID       int64    `json:"roleId"`
	Permissions  []string `json:"permissions"`
	jwt.StandardClaims
}

// JWTAuth JWT认证中间件
func JWTAuth(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"code":    4001,
				"message": "authorization header is empty",
			})
			c.Abort()
			return
		}

		// 检查Authorization格式
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(401, gin.H{
				"code":    4001,
				"message": "authorization format is wrong. should be 'Bearer {token}'",
			})
			c.Abort()
			return
		}

		// 解析token
		claims, err := ParseToken(parts[1], config.Auth.JwtSecret)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    4001,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("departmentId", claims.DepartmentID)
		c.Set("roleId", claims.RoleID)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(
	userID int64,
	username string,
	departmentID int64,
	roleID int64,
	permissions []string,
	secret string,
	expireHours int,
) (string, error) {
	// 设置token有效期
	expireTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	// 创建声明
	claims := JWTClaims{
		UserID:       userID,
		Username:     username,
		DepartmentID: departmentID,
		RoleID:       roleID,
		Permissions:  permissions,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "distributed-job-system",
		},
	}

	// 创建token
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString([]byte(secret))
}

// ParseToken 解析JWT令牌
func ParseToken(token string, secret string) (*JWTClaims, error) {
	// 解析token
	tokenClaims, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if tokenClaims != nil {
		// 验证token是否有效
		if claims, ok := tokenClaims.Claims.(*JWTClaims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, errors.New("invalid token")
}
