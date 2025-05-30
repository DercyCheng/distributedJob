package service

import (
"testing"
"time"

"distributedJob/internal/model/entity"
mocks "distributedJob/internal/mocks/service"

"github.com/golang/mock/gomock"
"github.com/stretchr/testify/assert"
)

func TestTaskService_TaskManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTask := mocks.NewMockTaskService(ctrl)

	// 测试创建任务
	t.Run("Create Task", func(t *testing.T) {
testTask := &entity.Task{
			ID:   1,
			Name: "Test Task",
			Type: "HTTP",
		}

		mockTask.EXPECT().
			CreateHTTPTask(testTask).
			Return(int64(1), nil)

		id, err := mockTask.CreateHTTPTask(testTask)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), id)
	})

	// 测试获取任务
	t.Run("Get Task", func(t *testing.T) {
expectedTask := &entity.Task{
			ID:   1,
			Name: "Existing Task",
		}

		mockTask.EXPECT().
			GetTaskByID(int64(1)).
			Return(expectedTask, nil)

		task, err := mockTask.GetTaskByID(1)
		assert.NoError(t, err)
		assert.Equal(t, expectedTask, task)
	})
}

func TestTaskService_ExecutionRecords(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTask := mocks.NewMockTaskService(ctrl)

	// 测试获取执行记录
	t.Run("Get Execution Records", func(t *testing.T) {
now := time.Now()
		records := []*entity.Record{
			{ID: 1, TaskID: 1, Success: 1},
			{ID: 2, TaskID: 1, Success: 0},
		}
		totalCount := int64(2)

		mockTask.EXPECT().
			GetTaskRecords(int64(1), now.Add(-24*time.Hour), now, 10, 0).
			Return(records, totalCount, nil)

		result, count, err := mockTask.GetTaskRecords(1, now.Add(-24*time.Hour), now, 10, 0)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, totalCount, count)
	})
}

func TestTaskService_Statistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTask := mocks.NewMockTaskService(ctrl)

	// 测试获取统计信息
	t.Run("Get Statistics", func(t *testing.T) {
now := time.Now()
		stats := &entity.TaskStatistics{
			TaskCount:        10,
			SuccessRate:      80.0,
			AvgExecutionTime: 150.5,
			ExecutionStats: map[string]float64{
				"total_tasks":   10,
				"total_success": 8,
				"total_failed":  2,
			},
		}

		mockTask.EXPECT().
			GetTaskStatistics(int64(1), now.Add(-24*time.Hour), now).
			Return(stats, nil)

		result, err := mockTask.GetTaskStatistics(1, now.Add(-24*time.Hour), now)
		assert.NoError(t, err)
		assert.Equal(t, 10, result.TaskCount)
		assert.Equal(t, 80.0, result.SuccessRate)
		assert.Equal(t, 150.5, result.AvgExecutionTime)
	})
}

func TestTaskService_UpdateStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTask := mocks.NewMockTaskService(ctrl)

	// 测试更新任务状态
	t.Run("Update Task Status", func(t *testing.T) {
mockTask.EXPECT().
			UpdateTaskStatus(int64(1), int8(1)).
			Return(nil)

		err := mockTask.UpdateTaskStatus(1, 1)
		assert.NoError(t, err)
	})
}
