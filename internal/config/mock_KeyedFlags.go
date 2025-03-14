// Code generated by mockery v2.7.4. DO NOT EDIT.

package config

import mock "github.com/stretchr/testify/mock"

// MockKeyedFlags is an autogenerated mock type for the KeyedFlags type
type MockKeyedFlags struct {
	mock.Mock
}

// Bool provides a mock function with given fields: k
func (_m *MockKeyedFlags) Bool(k string) bool {
	ret := _m.Called(k)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(k)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Int provides a mock function with given fields: k
func (_m *MockKeyedFlags) Int(k string) int {
	ret := _m.Called(k)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(k)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// IsSet provides a mock function with given fields: k
func (_m *MockKeyedFlags) IsSet(k string) bool {
	ret := _m.Called(k)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(k)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// String provides a mock function with given fields: k
func (_m *MockKeyedFlags) String(k string) string {
	ret := _m.Called(k)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(k)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
