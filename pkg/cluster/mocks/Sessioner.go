// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import cluster "github.com/charter-se/barrelman/pkg/cluster"
import io "io"
import mock "github.com/stretchr/testify/mock"

// Sessioner is an autogenerated mock type for the Sessioner type
type Sessioner struct {
	mock.Mock
}

// DeleteRelease provides a mock function with given fields: m
func (_m *Sessioner) DeleteRelease(m *cluster.DeleteMeta) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(*cluster.DeleteMeta) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteReleases provides a mock function with given fields: dm
func (_m *Sessioner) DeleteReleases(dm []*cluster.DeleteMeta) error {
	ret := _m.Called(dm)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*cluster.DeleteMeta) error); ok {
		r0 = rf(dm)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DiffManifests provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *Sessioner) DiffManifests(_a0 map[string]*cluster.MappingResult, _a1 map[string]*cluster.MappingResult, _a2 []string, _a3 int, _a4 io.Writer) bool {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]*cluster.MappingResult, map[string]*cluster.MappingResult, []string, int, io.Writer) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// DiffRelease provides a mock function with given fields: m
func (_m *Sessioner) DiffRelease(m *cluster.ReleaseMeta) (bool, []byte, error) {
	ret := _m.Called(m)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*cluster.ReleaseMeta) bool); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 []byte
	if rf, ok := ret.Get(1).(func(*cluster.ReleaseMeta) []byte); ok {
		r1 = rf(m)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]byte)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*cluster.ReleaseMeta) error); ok {
		r2 = rf(m)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetKubeConfig provides a mock function with given fields:
func (_m *Sessioner) GetKubeConfig() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetKubeContext provides a mock function with given fields:
func (_m *Sessioner) GetKubeContext() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Init provides a mock function with given fields:
func (_m *Sessioner) Init() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// InstallRelease provides a mock function with given fields: _a0
func (_m *Sessioner) InstallRelease(_a0 *cluster.ReleaseMeta) (string, string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(*cluster.ReleaseMeta) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(*cluster.ReleaseMeta) string); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(*cluster.ReleaseMeta) error); ok {
		r2 = rf(_a0)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ListReleases provides a mock function with given fields:
func (_m *Sessioner) ListReleases() ([]*cluster.Release, error) {
	ret := _m.Called()

	var r0 []*cluster.Release
	if rf, ok := ret.Get(0).(func() []*cluster.Release); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*cluster.Release)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Releases provides a mock function with given fields:
func (_m *Sessioner) Releases() (map[string]*cluster.ReleaseMeta, error) {
	ret := _m.Called()

	var r0 map[string]*cluster.ReleaseMeta
	if rf, ok := ret.Get(0).(func() map[string]*cluster.ReleaseMeta); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*cluster.ReleaseMeta)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetKubeConfig provides a mock function with given fields: c
func (_m *Sessioner) SetKubeConfig(c string) {
	_m.Called(c)
}

// SetKubeContext provides a mock function with given fields: c
func (_m *Sessioner) SetKubeContext(c string) {
	_m.Called(c)
}

// UpgradeRelease provides a mock function with given fields: m
func (_m *Sessioner) UpgradeRelease(m *cluster.ReleaseMeta) (string, error) {
	ret := _m.Called(m)

	var r0 string
	if rf, ok := ret.Get(0).(func(*cluster.ReleaseMeta) string); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*cluster.ReleaseMeta) error); ok {
		r1 = rf(m)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
