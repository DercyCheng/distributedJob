package service

import (
	"errors"
	"testing"

	mocks "distributedJob/internal/mocks/service"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Define errors for testing
var ErrModelNotFound = errors.New("model not found")

func TestModelContextService_GetModelContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModelContext := mocks.NewMockModelContextService(ctrl)

	// 测试用例1：成功获取模型上下文
	t.Run("Get Model Context Success", func(t *testing.T) {
		expectedContext := "Context for model model-1"

		mockModelContext.EXPECT().
			GetModelContext("model-1").
			Return(expectedContext, nil)
		ctx, err := mockModelContext.GetModelContext("model-1")
		assert.NoError(t, err)
		assert.Equal(t, expectedContext, ctx)
	})

	// 测试用例2：模型不存在
	t.Run("Model Not Found", func(t *testing.T) {
		mockModelContext.EXPECT().
			GetModelContext("nonexistent-model").
			Return("", ErrModelNotFound)

		_, err := mockModelContext.GetModelContext("nonexistent-model")
		assert.ErrorIs(t, err, ErrModelNotFound)
	})
}
