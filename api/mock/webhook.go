// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/HyperGAI/serving-api/api (interfaces: Webhook)

// Package mockapi is a generated GoMock package.
package mockapi

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockWebhook is a mock of Webhook interface.
type MockWebhook struct {
	ctrl     *gomock.Controller
	recorder *MockWebhookMockRecorder
}

// MockWebhookMockRecorder is the mock recorder for MockWebhook.
type MockWebhookMockRecorder struct {
	mock *MockWebhook
}

// NewMockWebhook creates a new mock instance.
func NewMockWebhook(ctrl *gomock.Controller) *MockWebhook {
	mock := &MockWebhook{ctrl: ctrl}
	mock.recorder = &MockWebhookMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWebhook) EXPECT() *MockWebhookMockRecorder {
	return m.recorder
}

// GetTaskInfo mocks base method.
func (m *MockWebhook) GetTaskInfo(arg0 string) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTaskInfo", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTaskInfo indicates an expected call of GetTaskInfo.
func (mr *MockWebhookMockRecorder) GetTaskInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTaskInfo", reflect.TypeOf((*MockWebhook)(nil).GetTaskInfo), arg0)
}