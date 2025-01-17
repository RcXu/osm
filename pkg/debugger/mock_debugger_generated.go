// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/openservicemesh/osm/pkg/debugger (interfaces: MeshCatalogDebugger,XDSDebugger)

// Package debugger is a generated GoMock package.
package debugger

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	envoy "github.com/openservicemesh/osm/pkg/envoy"
	identity "github.com/openservicemesh/osm/pkg/identity"
	v1alpha3 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/access/v1alpha3"
	v1alpha4 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/specs/v1alpha4"
	v1alpha2 "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha2"
)

// MockMeshCatalogDebugger is a mock of MeshCatalogDebugger interface.
type MockMeshCatalogDebugger struct {
	ctrl     *gomock.Controller
	recorder *MockMeshCatalogDebuggerMockRecorder
}

// MockMeshCatalogDebuggerMockRecorder is the mock recorder for MockMeshCatalogDebugger.
type MockMeshCatalogDebuggerMockRecorder struct {
	mock *MockMeshCatalogDebugger
}

// NewMockMeshCatalogDebugger creates a new mock instance.
func NewMockMeshCatalogDebugger(ctrl *gomock.Controller) *MockMeshCatalogDebugger {
	mock := &MockMeshCatalogDebugger{ctrl: ctrl}
	mock.recorder = &MockMeshCatalogDebuggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMeshCatalogDebugger) EXPECT() *MockMeshCatalogDebuggerMockRecorder {
	return m.recorder
}

// ListSMIPolicies mocks base method.
func (m *MockMeshCatalogDebugger) ListSMIPolicies() ([]*v1alpha2.TrafficSplit, []identity.K8sServiceAccount, []*v1alpha4.HTTPRouteGroup, []*v1alpha3.TrafficTarget) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSMIPolicies")
	ret0, _ := ret[0].([]*v1alpha2.TrafficSplit)
	ret1, _ := ret[1].([]identity.K8sServiceAccount)
	ret2, _ := ret[2].([]*v1alpha4.HTTPRouteGroup)
	ret3, _ := ret[3].([]*v1alpha3.TrafficTarget)
	return ret0, ret1, ret2, ret3
}

// ListSMIPolicies indicates an expected call of ListSMIPolicies.
func (mr *MockMeshCatalogDebuggerMockRecorder) ListSMIPolicies() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSMIPolicies", reflect.TypeOf((*MockMeshCatalogDebugger)(nil).ListSMIPolicies))
}

// MockXDSDebugger is a mock of XDSDebugger interface.
type MockXDSDebugger struct {
	ctrl     *gomock.Controller
	recorder *MockXDSDebuggerMockRecorder
}

// MockXDSDebuggerMockRecorder is the mock recorder for MockXDSDebugger.
type MockXDSDebuggerMockRecorder struct {
	mock *MockXDSDebugger
}

// NewMockXDSDebugger creates a new mock instance.
func NewMockXDSDebugger(ctrl *gomock.Controller) *MockXDSDebugger {
	mock := &MockXDSDebugger{ctrl: ctrl}
	mock.recorder = &MockXDSDebuggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockXDSDebugger) EXPECT() *MockXDSDebuggerMockRecorder {
	return m.recorder
}

// GetXDSLog mocks base method.
func (m *MockXDSDebugger) GetXDSLog() map[string]map[envoy.TypeURI][]time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetXDSLog")
	ret0, _ := ret[0].(map[string]map[envoy.TypeURI][]time.Time)
	return ret0
}

// GetXDSLog indicates an expected call of GetXDSLog.
func (mr *MockXDSDebuggerMockRecorder) GetXDSLog() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetXDSLog", reflect.TypeOf((*MockXDSDebugger)(nil).GetXDSLog))
}
