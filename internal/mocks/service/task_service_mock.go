// Code generated by MockGen. DO NOT EDIT.
// Source: internal/service/task_service.go

// Package service is a generated GoMock package.
package service

import (
	entity "distributedJob/internal/model/entity"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
)

// MockTaskService is a mock of TaskService interface.
type MockTaskService struct {
	ctrl     *gomock.Controller
	recorder *MockTaskServiceMockRecorder
}

// MockTaskServiceMockRecorder is the mock recorder for MockTaskService.
type MockTaskServiceMockRecorder struct {
	mock *MockTaskService
}

// NewMockTaskService creates a new mock instance.
func NewMockTaskService(ctrl *gomock.Controller) *MockTaskService {
	mock := &MockTaskService{ctrl: ctrl}
	mock.recorder = &MockTaskServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTaskService) EXPECT() *MockTaskServiceMockRecorder {
	return m.recorder
}

// CreateGRPCTask mocks base method.
func (m *MockTaskService) CreateGRPCTask(task *entity.Task) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateGRPCTask", task)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateGRPCTask indicates an expected call of CreateGRPCTask.
func (mr *MockTaskServiceMockRecorder) CreateGRPCTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateGRPCTask", reflect.TypeOf((*MockTaskService)(nil).CreateGRPCTask), task)
}

// CreateHTTPTask mocks base method.
func (m *MockTaskService) CreateHTTPTask(task *entity.Task) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateHTTPTask", task)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateHTTPTask indicates an expected call of CreateHTTPTask.
func (mr *MockTaskServiceMockRecorder) CreateHTTPTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateHTTPTask", reflect.TypeOf((*MockTaskService)(nil).CreateHTTPTask), task)
}

// DeleteTask mocks base method.
func (m *MockTaskService) DeleteTask(id int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteTask", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteTask indicates an expected call of DeleteTask.
func (mr *MockTaskServiceMockRecorder) DeleteTask(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteTask", reflect.TypeOf((*MockTaskService)(nil).DeleteTask), id)
}

// GetRecordByID mocks base method.
func (m *MockTaskService) GetRecordByID(id int64, year, month int) (*entity.Record, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecordByID", id, year, month)
	ret0, _ := ret[0].(*entity.Record)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecordByID indicates an expected call of GetRecordByID.
func (mr *MockTaskServiceMockRecorder) GetRecordByID(id, year, month interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordByID", reflect.TypeOf((*MockTaskService)(nil).GetRecordByID), id, year, month)
}

// GetRecordList mocks base method.
func (m *MockTaskService) GetRecordList(year, month int, taskID, departmentID *int64, success *int8, page, size int) ([]*entity.Record, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecordList", year, month, taskID, departmentID, success, page, size)
	ret0, _ := ret[0].([]*entity.Record)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetRecordList indicates an expected call of GetRecordList.
func (mr *MockTaskServiceMockRecorder) GetRecordList(year, month, taskID, departmentID, success, page, size interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordList", reflect.TypeOf((*MockTaskService)(nil).GetRecordList), year, month, taskID, departmentID, success, page, size)
}

// GetRecordListByTimeRange mocks base method.
func (m *MockTaskService) GetRecordListByTimeRange(year, month int, taskID, departmentID *int64, success *int8, page, size int, startTime, endTime time.Time) ([]*entity.Record, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecordListByTimeRange", year, month, taskID, departmentID, success, page, size, startTime, endTime)
	ret0, _ := ret[0].([]*entity.Record)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetRecordListByTimeRange indicates an expected call of GetRecordListByTimeRange.
func (mr *MockTaskServiceMockRecorder) GetRecordListByTimeRange(year, month, taskID, departmentID, success, page, size, startTime, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordListByTimeRange", reflect.TypeOf((*MockTaskService)(nil).GetRecordListByTimeRange), year, month, taskID, departmentID, success, page, size, startTime, endTime)
}

// GetRecordStats mocks base method.
func (m *MockTaskService) GetRecordStats(year, month int, taskID, departmentID *int64) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRecordStats", year, month, taskID, departmentID)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRecordStats indicates an expected call of GetRecordStats.
func (mr *MockTaskServiceMockRecorder) GetRecordStats(year, month, taskID, departmentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRecordStats", reflect.TypeOf((*MockTaskService)(nil).GetRecordStats), year, month, taskID, departmentID)
}

// GetTaskByID mocks base method.
func (m *MockTaskService) GetTaskByID(id int64) (*entity.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskByID", id)
	ret0, _ := ret[0].(*entity.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTaskByID indicates an expected call of GetTaskByID.
func (mr *MockTaskServiceMockRecorder) GetTaskByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskByID", reflect.TypeOf((*MockTaskService)(nil).GetTaskByID), id)
}

// GetTaskList mocks base method.
func (m *MockTaskService) GetTaskList(departmentID int64, page, size int) ([]*entity.Task, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskList", departmentID, page, size)
	ret0, _ := ret[0].([]*entity.Task)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetTaskList indicates an expected call of GetTaskList.
func (mr *MockTaskServiceMockRecorder) GetTaskList(departmentID, page, size interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskList", reflect.TypeOf((*MockTaskService)(nil).GetTaskList), departmentID, page, size)
}

// GetTaskRecords mocks base method.
func (m *MockTaskService) GetTaskRecords(taskID int64, startTime, endTime time.Time, limit, offset int) ([]*entity.Record, int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskRecords", taskID, startTime, endTime, limit, offset)
	ret0, _ := ret[0].([]*entity.Record)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetTaskRecords indicates an expected call of GetTaskRecords.
func (mr *MockTaskServiceMockRecorder) GetTaskRecords(taskID, startTime, endTime, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskRecords", reflect.TypeOf((*MockTaskService)(nil).GetTaskRecords), taskID, startTime, endTime, limit, offset)
}

// GetTaskStatistics mocks base method.
func (m *MockTaskService) GetTaskStatistics(departmentID int64, startTime, endTime time.Time) (*entity.TaskStatistics, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskStatistics", departmentID, startTime, endTime)
	ret0, _ := ret[0].(*entity.TaskStatistics)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTaskStatistics indicates an expected call of GetTaskStatistics.
func (mr *MockTaskServiceMockRecorder) GetTaskStatistics(departmentID, startTime, endTime interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskStatistics", reflect.TypeOf((*MockTaskService)(nil).GetTaskStatistics), departmentID, startTime, endTime)
}

// SetMetrics mocks base method.
func (m *MockTaskService) SetMetrics(metrics interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetMetrics", metrics)
}

// SetMetrics indicates an expected call of SetMetrics.
func (mr *MockTaskServiceMockRecorder) SetMetrics(metrics interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMetrics", reflect.TypeOf((*MockTaskService)(nil).SetMetrics), metrics)
}

// SetTracer mocks base method.
func (m *MockTaskService) SetTracer(tracer interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTracer", tracer)
}

// SetTracer indicates an expected call of SetTracer.
func (mr *MockTaskServiceMockRecorder) SetTracer(tracer interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTracer", reflect.TypeOf((*MockTaskService)(nil).SetTracer), tracer)
}

// UpdateGRPCTask mocks base method.
func (m *MockTaskService) UpdateGRPCTask(task *entity.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateGRPCTask", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateGRPCTask indicates an expected call of UpdateGRPCTask.
func (mr *MockTaskServiceMockRecorder) UpdateGRPCTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateGRPCTask", reflect.TypeOf((*MockTaskService)(nil).UpdateGRPCTask), task)
}

// UpdateHTTPTask mocks base method.
func (m *MockTaskService) UpdateHTTPTask(task *entity.Task) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateHTTPTask", task)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateHTTPTask indicates an expected call of UpdateHTTPTask.
func (mr *MockTaskServiceMockRecorder) UpdateHTTPTask(task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateHTTPTask", reflect.TypeOf((*MockTaskService)(nil).UpdateHTTPTask), task)
}

// UpdateTaskStatus mocks base method.
func (m *MockTaskService) UpdateTaskStatus(id int64, status int8) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateTaskStatus", id, status)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateTaskStatus indicates an expected call of UpdateTaskStatus.
func (mr *MockTaskServiceMockRecorder) UpdateTaskStatus(id, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateTaskStatus", reflect.TypeOf((*MockTaskService)(nil).UpdateTaskStatus), id, status)
}
