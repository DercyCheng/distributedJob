package auth

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
	DepartmentID string   `json:"department_id"`
	Roles        []string `json:"roles"`
	Permissions  []string `json:"permissions"`
	jwt.StandardClaims
}

type JWTManager struct {
	secretKey     string
	tokenDuration time.Duration
	Duration      time.Duration // Add public field for access
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     secretKey,
		tokenDuration: tokenDuration,
		Duration:      tokenDuration,
	}
}

func (manager *JWTManager) Generate(userID, username, departmentID string, roles, permissions []string) (string, error) {
	claims := Claims{
		UserID:       userID,
		Username:     username,
		DepartmentID: departmentID,
		Roles:        roles,
		Permissions:  permissions,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// GenerateRefresh generates a refresh token with longer duration
func (manager *JWTManager) GenerateRefresh(userID, username, departmentID string, roles, permissions []string) (string, error) {
	claims := Claims{
		UserID:       userID,
		Username:     username,
		DepartmentID: departmentID,
		Roles:        roles,
		Permissions:  permissions,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(manager.tokenDuration * 24).Unix(), // 24x longer
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

func (manager *JWTManager) Verify(accessToken string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}

			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func (manager *JWTManager) HasPermission(claims *Claims, permission string) bool {
	for _, p := range claims.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func (manager *JWTManager) HasRole(claims *Claims, role string) bool {
	for _, r := range claims.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (manager *JWTManager) InDepartment(claims *Claims, departmentID string) bool {
	return claims.DepartmentID == departmentID
}
