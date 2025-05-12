package middleware

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"distributedJob/internal/config"
	"distributedJob/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims 访问令牌的JWT声明
type AccessClaims struct {
	UserID       int64    `json:"userId"`
	Username     string   `json:"username"`
	DepartmentID int64    `json:"departmentId"`
	RoleID       int64    `json:"roleId"`
	Permissions  []string `json:"permissions"`
	jwt.RegisteredClaims
}

// RefreshClaims 刷新令牌的JWT声明
type RefreshClaims struct {
	UserID int64 `json:"userId"`
	jwt.RegisteredClaims
}

// JWTAuth JWT认证中间件
func JWTAuth(cfg *config.Config, tokenRevoker store.TokenRevoker) gin.HandlerFunc {
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
		claims, err := ParseAccessToken(parts[1], cfg.Auth.JwtSecret, tokenRevoker)
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

// RefreshAuth 刷新令牌认证中间件 - 专门用于token刷新接口
func RefreshAuth(cfg *config.Config, tokenRevoker store.TokenRevoker) gin.HandlerFunc {
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

		// 解析刷新令牌
		claims, err := ParseRefreshToken(parts[1], cfg.Auth.JwtRefreshSecret, tokenRevoker)
		if err != nil {
			c.JSON(401, gin.H{
				"code":    4001,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("userId", claims.UserID)

		c.Next()
	}
}

// GenerateAccessToken 生成访问令牌
func GenerateAccessToken(
	userID int64,
	username string,
	departmentID int64,
	roleID int64,
	permissions []string,
	secret string,
	expireMinutes int,
) (string, error) {
	// 设置token有效期
	expireTime := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	// 创建声明
	claims := AccessClaims{
		UserID:       userID,
		Username:     username,
		DepartmentID: departmentID,
		RoleID:       roleID,
		Permissions:  permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        generateTokenID(userID),
			Issuer:    "distributed-job-system",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	// 创建token
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString([]byte(secret))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(
	userID int64,
	secret string,
	expireDays int,
) (string, error) {
	// 设置token有效期
	expireTime := time.Now().Add(time.Duration(expireDays) * 24 * time.Hour)

	// 创建声明
	claims := RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        generateRefreshTokenID(userID),
			Issuer:    "distributed-job-system",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	// 创建token
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenClaims.SignedString([]byte(secret))
}

// ParseAccessToken 解析访问令牌
func ParseAccessToken(tokenString string, secret string, tokenRevoker store.TokenRevoker) (*AccessClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		// 处理特定错误
		if strings.Contains(err.Error(), "token is expired") {
			return nil, errors.New("token has expired")
		}
		return nil, err
	}

	// 验证token是否有效
	if claims, ok := token.Claims.(*AccessClaims); ok && token.Valid {
		// 检查令牌是否被撤销
		if tokenRevoker != nil {
			jti := claims.ID
			if jti == "" {
				return nil, errors.New("invalid token id")
			}
			if tokenRevoker.IsRevoked(jti) {
				return nil, errors.New("token has been revoked")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ParseRefreshToken 解析刷新令牌
func ParseRefreshToken(tokenString string, secret string, tokenRevoker store.TokenRevoker) (*RefreshClaims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		// 处理令牌过期错误
		if strings.Contains(err.Error(), "token is expired") {
			return nil, errors.New("refresh token has expired")
		}
		return nil, err
	}
	// 验证token是否有效
	if claims, ok := token.Claims.(*RefreshClaims); ok && token.Valid {
		// 检查令牌是否被撤销
		if tokenRevoker != nil {
			jti := claims.ID
			if jti == "" {
				return nil, errors.New("invalid refresh token id")
			}
			if tokenRevoker.IsRevoked(jti) {
				return nil, errors.New("refresh token has been revoked")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid refresh token")
}

// generateTokenID 生成访问令牌ID
func generateTokenID(userID int64) string {
	return fmt.Sprintf("access_%d_%d", userID, time.Now().UnixNano())
}

// generateRefreshTokenID 生成刷新令牌ID
func generateRefreshTokenID(userID int64) string {
	return fmt.Sprintf("refresh_%d_%d", userID, time.Now().UnixNano())
}
