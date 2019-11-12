// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/solo-io/valet/cli/internal/ensure/cmd (interfaces: Runner)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	cmd "github.com/solo-io/valet/cli/internal/ensure/cmd"
	http "net/http"
	reflect "reflect"
)

// MockRunner is a mock of Runner interface
type MockRunner struct {
	ctrl     *gomock.Controller
	recorder *MockRunnerMockRecorder
}

// MockRunnerMockRecorder is the mock recorder for MockRunner
type MockRunnerMockRecorder struct {
	mock *MockRunner
}

// NewMockRunner creates a new mock instance
func NewMockRunner(ctrl *gomock.Controller) *MockRunner {
	mock := &MockRunner{ctrl: ctrl}
	mock.recorder = &MockRunnerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRunner) EXPECT() *MockRunnerMockRecorder {
	return m.recorder
}

// Output mocks base method
func (m *MockRunner) Output(arg0 context.Context, arg1 *cmd.Command) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Output", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Output indicates an expected call of Output
func (mr *MockRunnerMockRecorder) Output(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Output", reflect.TypeOf((*MockRunner)(nil).Output), arg0, arg1)
}

// Request mocks base method
func (m *MockRunner) Request(arg0 context.Context, arg1 *http.Request) (string, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Request", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Request indicates an expected call of Request
func (mr *MockRunnerMockRecorder) Request(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Request", reflect.TypeOf((*MockRunner)(nil).Request), arg0, arg1)
}

// Run mocks base method
func (m *MockRunner) Run(arg0 context.Context, arg1 *cmd.Command) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Run", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Run indicates an expected call of Run
func (mr *MockRunnerMockRecorder) Run(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockRunner)(nil).Run), arg0, arg1)
}

// Stream mocks base method
func (m *MockRunner) Stream(arg0 context.Context, arg1 *cmd.Command) (*cmd.CommandStreamHandler, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stream", arg0, arg1)
	ret0, _ := ret[0].(*cmd.CommandStreamHandler)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stream indicates an expected call of Stream
func (mr *MockRunnerMockRecorder) Stream(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockRunner)(nil).Stream), arg0, arg1)
}
