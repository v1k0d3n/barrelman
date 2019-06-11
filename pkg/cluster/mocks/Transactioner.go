// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import cluster "github.com/charter-oss/barrelman/pkg/cluster"
import mock "github.com/stretchr/testify/mock"

// Transactioner is an autogenerated mock type for the Transactioner type
type Transactioner struct {
	mock.Mock
}

// Cancel provides a mock function with given fields:
func (_m *Transactioner) Cancel() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Canceled provides a mock function with given fields:
func (_m *Transactioner) Canceled() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Complete provides a mock function with given fields:
func (_m *Transactioner) Complete() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Completed provides a mock function with given fields:
func (_m *Transactioner) Completed() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Started provides a mock function with given fields:
func (_m *Transactioner) Started() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Versions provides a mock function with given fields:
func (_m *Transactioner) Versions() *cluster.Versions {
	ret := _m.Called()

	var r0 *cluster.Versions
	if rf, ok := ret.Get(0).(func() *cluster.Versions); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cluster.Versions)
		}
	}

	return r0
}

// WriteNewVersion provides a mock function with given fields:
func (_m *Transactioner) WriteNewVersion() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
