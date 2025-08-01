// Code generated by MockGen. DO NOT EDIT.
// Source: pkg/llm/types.go
//
// Generated by this command:
//
//	mockgen -source=pkg/llm/types.go -destination=pkg/mocks/llm/mock_llm.go -package=llm
//

// Package llm is a generated GoMock package.
package llm

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
	isgomock struct{}
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// CreateThread mocks base method.
func (m *MockInterface) CreateThread(project, version string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateThread", project, version)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateThread indicates an expected call of CreateThread.
func (mr *MockInterfaceMockRecorder) CreateThread(project, version any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateThread", reflect.TypeOf((*MockInterface)(nil).CreateThread), project, version)
}

// Elaborate mocks base method.
func (m *MockInterface) Elaborate(threadSlug, message string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Elaborate", threadSlug, message)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Elaborate indicates an expected call of Elaborate.
func (mr *MockInterfaceMockRecorder) Elaborate(threadSlug, message any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Elaborate", reflect.TypeOf((*MockInterface)(nil).Elaborate), threadSlug, message)
}

// Inject mocks base method.
func (m *MockInterface) Inject(project, version, message string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inject", project, version, message)
	ret0, _ := ret[0].(error)
	return ret0
}

// Inject indicates an expected call of Inject.
func (mr *MockInterfaceMockRecorder) Inject(project, version, message any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inject", reflect.TypeOf((*MockInterface)(nil).Inject), project, version, message)
}

// SendMessageToChat mocks base method.
func (m *MockInterface) SendMessageToChat(project, version, threadSlug, message string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessageToChat", project, version, threadSlug, message)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendMessageToChat indicates an expected call of SendMessageToChat.
func (mr *MockInterfaceMockRecorder) SendMessageToChat(project, version, threadSlug, message any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessageToChat", reflect.TypeOf((*MockInterface)(nil).SendMessageToChat), project, version, threadSlug, message)
}
