// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
import mock "github.com/stretchr/testify/mock"
import types "k8s.io/apimachinery/pkg/types"
import v1 "k8s.io/api/core/v1"
import watch "k8s.io/apimachinery/pkg/watch"

// ConfigMapInterface is an autogenerated mock type for the ConfigMapInterface type
type ConfigMapInterface struct {
	mock.Mock
}

// Create provides a mock function with given fields: _a0
func (_m *ConfigMapInterface) Create(_a0 *v1.ConfigMap) (*v1.ConfigMap, error) {
	ret := _m.Called(_a0)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(*v1.ConfigMap) *v1.ConfigMap); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.ConfigMap) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name, options
func (_m *ConfigMapInterface) Delete(name string, options *metav1.DeleteOptions) error {
	ret := _m.Called(name, options)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *metav1.DeleteOptions) error); ok {
		r0 = rf(name, options)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteCollection provides a mock function with given fields: options, listOptions
func (_m *ConfigMapInterface) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	ret := _m.Called(options, listOptions)

	var r0 error
	if rf, ok := ret.Get(0).(func(*metav1.DeleteOptions, metav1.ListOptions) error); ok {
		r0 = rf(options, listOptions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: name, options
func (_m *ConfigMapInterface) Get(name string, options metav1.GetOptions) (*v1.ConfigMap, error) {
	ret := _m.Called(name, options)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(string, metav1.GetOptions) *v1.ConfigMap); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, metav1.GetOptions) error); ok {
		r1 = rf(name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: opts
func (_m *ConfigMapInterface) List(opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	ret := _m.Called(opts)

	var r0 *v1.ConfigMapList
	if rf, ok := ret.Get(0).(func(metav1.ListOptions) *v1.ConfigMapList); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMapList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metav1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Patch provides a mock function with given fields: name, pt, data, subresources
func (_m *ConfigMapInterface) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (*v1.ConfigMap, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, name, pt, data)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(string, types.PatchType, []byte, ...string) *v1.ConfigMap); ok {
		r0 = rf(name, pt, data, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, types.PatchType, []byte, ...string) error); ok {
		r1 = rf(name, pt, data, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: _a0
func (_m *ConfigMapInterface) Update(_a0 *v1.ConfigMap) (*v1.ConfigMap, error) {
	ret := _m.Called(_a0)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(*v1.ConfigMap) *v1.ConfigMap); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1.ConfigMap) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: opts
func (_m *ConfigMapInterface) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(metav1.ListOptions) watch.Interface); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(metav1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
