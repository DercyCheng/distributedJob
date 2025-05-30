package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"distributedJob/internal/api"
	"distributedJob/internal/config"
	"distributedJob/internal/model/entity"

	"github.com/gin-gonic/gin"
)

// mockAuthService implements api.AuthService interface for testing Login
// only Login method used in tests
type mockAuthService struct {
	loginFunc func(username, password string) (string, string, *entity.User, error)
}

func (m *mockAuthService) SetTracer(interface{}) {}
func (m *mockAuthService) Login(username, password string) (string, string, *entity.User, error) {
	return m.loginFunc(username, password)
}
func (m *mockAuthService) GenerateTokens(*entity.User) (string, string, error) { return "", "", nil }
func (m *mockAuthService) RefreshToken(string) (string, string, error)         { return "", "", nil }
func (m *mockAuthService) ValidateToken(string) (int64, error)                 { return 0, nil }
func (m *mockAuthService) ValidateRefreshToken(string) (int64, error)          { return 0, nil }
func (m *mockAuthService) RevokeToken(string) error                            { return nil }
func (m *mockAuthService) IsTokenRevoked(string) bool                          { return false }
func (m *mockAuthService) GetUserList(int64, int, int) ([]*entity.User, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthService) GetUserByID(int64) (*entity.User, error)             { return nil, nil }
func (m *mockAuthService) CreateUser(*entity.User) (int64, error)              { return 0, nil }
func (m *mockAuthService) UpdateUser(*entity.User) error                       { return nil }
func (m *mockAuthService) DeleteUser(int64) error                              { return nil }
func (m *mockAuthService) GetUserPermissions(int64) ([]string, error)          { return nil, nil }
func (m *mockAuthService) GetRoleList(int, int) ([]*entity.Role, int64, error) { return nil, 0, nil }
func (m *mockAuthService) GetRoleByID(int64) (*entity.Role, error)             { return nil, nil }
func (m *mockAuthService) CreateRole(*entity.Role) (int64, error)              { return 0, nil }
func (m *mockAuthService) UpdateRole(*entity.Role) error                       { return nil }
func (m *mockAuthService) DeleteRole(int64) error                              { return nil }
func (m *mockAuthService) GetDepartmentList(int, int) ([]*entity.Department, int64, error) {
	return nil, 0, nil
}
func (m *mockAuthService) GetDepartmentByID(int64) (*entity.Department, error) { return nil, nil }
func (m *mockAuthService) CreateDepartment(*entity.Department) (int64, error)  { return 0, nil }
func (m *mockAuthService) UpdateDepartment(*entity.Department) error           { return nil }
func (m *mockAuthService) DeleteDepartment(int64) error                        { return nil }
func (m *mockAuthService) GetPermissionList() ([]*entity.Permission, error)    { return nil, nil }
func (m *mockAuthService) HasPermission(int64, string) (bool, error)           { return false, nil }

// setupRouter constructs a gin Engine with the login route for testing
func setupRouter(authService *mockAuthService) *gin.Engine {
	// use default config with JwtExpireMinutes set
	cfg := &config.Config{Auth: config.AuthConfig{JwtExpireMinutes: 15}}
	return api.CreateLoginTestRouter(cfg, authService)
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]string
		mockLoginFunc      func(username, password string) (string, string, *entity.User, error)
		expectedHTTPStatus int
		expectedCode       int
		expectedField      string
	}{
		{
			name:        "success",
			requestBody: map[string]string{"username": "user1", "password": "password1"},
			mockLoginFunc: func(u, p string) (string, string, *entity.User, error) {
				return "access-token", "refresh-token", &entity.User{ID: 1, Username: "user1", RealName: "User One", DepartmentID: 2, RoleID: 3}, nil
			},
			expectedHTTPStatus: http.StatusOK,
			expectedCode:       api.ErrorCodes.Success,
		}, {
			name:               "missing fields",
			requestBody:        map[string]string{},
			mockLoginFunc:      nil,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedCode:       api.ErrorCodes.ValidationFailed,
		}, {
			name:               "short username",
			requestBody:        map[string]string{"username": "ab", "password": "abcdef"},
			mockLoginFunc:      nil,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedCode:       api.ErrorCodes.ValidationFailed,
			expectedField:      "username",
		}, {
			name:               "short password",
			requestBody:        map[string]string{"username": "abc", "password": "123"},
			mockLoginFunc:      nil,
			expectedHTTPStatus: http.StatusBadRequest,
			expectedCode:       api.ErrorCodes.ValidationFailed,
			expectedField:      "password",
		}, {
			name:        "user not found",
			requestBody: map[string]string{"username": "nouser", "password": "password"},
			mockLoginFunc: func(u, p string) (string, string, *entity.User, error) {
				return "", "", nil, fmt.Errorf("user not found")
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedCode:       api.ErrorCodes.Unauthorized,
			expectedField:      "username",
		}, {
			name:        "invalid password",
			requestBody: map[string]string{"username": "user2", "password": "wrongpass"},
			mockLoginFunc: func(u, p string) (string, string, *entity.User, error) {
				return "", "", nil, fmt.Errorf("invalid password")
			},
			expectedHTTPStatus: http.StatusUnauthorized,
			expectedCode:       api.ErrorCodes.Unauthorized,
			expectedField:      "password",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// prepare mock
			var authService *mockAuthService
			if tc.mockLoginFunc != nil {
				authService = &mockAuthService{loginFunc: tc.mockLoginFunc}
			} else {
				authService = &mockAuthService{loginFunc: func(u, p string) (string, string, *entity.User, error) { return "", "", nil, nil }}
			}

			r := setupRouter(authService)

			// prepare request
			bodyBytes, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// execute
			r.ServeHTTP(w, req)

			// check status
			if w.Code != tc.expectedHTTPStatus {
				t.Errorf("expected status %d, got %d", tc.expectedHTTPStatus, w.Code)
			}

			// parse response
			var resp api.ResponseBody
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			if err != nil {
				t.Fatalf("invalid JSON response: %v", err)
			}

			// check code
			if resp.Code != tc.expectedCode {
				t.Errorf("expected code %d, got %d", tc.expectedCode, resp.Code)
			}

			// check error field when expected
			if tc.expectedField != "" && resp.ErrorField != tc.expectedField {
				t.Errorf("expected errorField '%s', got '%s'", tc.expectedField, resp.ErrorField)
			}
		})
	}
}
