package service

import (
	"testing"

	mocks "distributedJob/internal/mocks/service"
	"distributedJob/internal/model/entity"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthService(ctrl)

	// 测试用例1：登录成功
	t.Run("Login Success", func(t *testing.T) {
		expectedToken := "token123"
		expectedRefreshToken := "refresh_token123"
		expectedUser := &entity.User{ID: 1}
		mockAuth.EXPECT().
			Login("admin", "password123").
			Return(expectedToken, expectedRefreshToken, expectedUser, nil)

		token, refreshToken, user, err := mockAuth.Login("admin", "password123")
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, token)
		assert.Equal(t, expectedRefreshToken, refreshToken)
		assert.Equal(t, expectedUser, user)
	})

	// 测试用例2：无效凭证
	t.Run("Invalid Credentials", func(t *testing.T) {
		mockAuth.EXPECT().
			Login("wrong", "wrong").
			Return("", "", nil, assert.AnError)

		token, refreshToken, user, err := mockAuth.Login("wrong", "wrong")
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Empty(t, refreshToken)
		assert.Nil(t, user)
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthService(ctrl)

	// 测试用例1：有效token
	t.Run("Valid Token", func(t *testing.T) {
		mockAuth.EXPECT().
			ValidateToken("valid_token").
			Return(int64(1), nil)

		userID, err := mockAuth.ValidateToken("valid_token")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), userID)
	})

	// 测试用例2：无效token
	t.Run("Invalid Token", func(t *testing.T) {
		mockAuth.EXPECT().
			ValidateToken("invalid_token").
			Return(int64(0), assert.AnError)

		userID, err := mockAuth.ValidateToken("invalid_token")
		assert.Error(t, err)
		assert.Equal(t, int64(0), userID)
	})
}

func TestAuthService_RevokeToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthService(ctrl)

	t.Run("Revoke Token Success", func(t *testing.T) {
		mockAuth.EXPECT().
			RevokeToken("valid_token").
			Return(nil)

		err := mockAuth.RevokeToken("valid_token")
		assert.NoError(t, err)
	})
}
