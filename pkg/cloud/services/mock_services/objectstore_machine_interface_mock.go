/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by MockGen. DO NOT EDIT.
// Source: sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/services (interfaces: ObjectStore)

// Package mock_services is a generated GoMock package.
package mock_services

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	scope "sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/scope"
)

// MockObjectStore is a mock of ObjectStore interface
type MockObjectStore struct {
	ctrl     *gomock.Controller
	recorder *MockObjectStoreMockRecorder
}

// MockObjectStoreMockRecorder is the mock recorder for MockObjectStore
type MockObjectStoreMockRecorder struct {
	mock *MockObjectStore
}

// NewMockObjectStore creates a new mock instance
func NewMockObjectStore(ctrl *gomock.Controller) *MockObjectStore {
	mock := &MockObjectStore{ctrl: ctrl}
	mock.recorder = &MockObjectStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockObjectStore) EXPECT() *MockObjectStoreMockRecorder {
	return m.recorder
}

// Create mocks base method
func (m *MockObjectStore) Create(arg0 *scope.MachineScope, arg1 []byte) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", arg0, arg1)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create
func (mr *MockObjectStoreMockRecorder) Create(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockObjectStore)(nil).Create), arg0, arg1)
}

// Delete mocks base method
func (m *MockObjectStore) Delete(arg0 *scope.MachineScope) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockObjectStoreMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockObjectStore)(nil).Delete), arg0)
}
