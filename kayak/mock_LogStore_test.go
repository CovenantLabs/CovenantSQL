// Code generated by mockery v1.0.0. DO NOT EDIT.
package kayak

import mock "github.com/stretchr/testify/mock"

// MockLogStore is an autogenerated mock type for the LogStore type
type MockLogStore struct {
	mock.Mock
}

// DeleteRange provides a mock function with given fields: min, max
func (_m *MockLogStore) DeleteRange(min uint64, max uint64) error {
	ret := _m.Called(min, max)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, uint64) error); ok {
		r0 = rf(min, max)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FirstIndex provides a mock function with given fields:
func (_m *MockLogStore) FirstIndex() (uint64, error) {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLog provides a mock function with given fields: index, l
func (_m *MockLogStore) GetLog(index uint64, l *Log) error {
	ret := _m.Called(index, l)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint64, *Log) error); ok {
		r0 = rf(index, l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// LastIndex provides a mock function with given fields:
func (_m *MockLogStore) LastIndex() (uint64, error) {
	ret := _m.Called()

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StoreLog provides a mock function with given fields: l
func (_m *MockLogStore) StoreLog(l *Log) error {
	ret := _m.Called(l)

	var r0 error
	if rf, ok := ret.Get(0).(func(*Log) error); ok {
		r0 = rf(l)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StoreLogs provides a mock function with given fields: logs
func (_m *MockLogStore) StoreLogs(logs []*Log) error {
	ret := _m.Called(logs)

	var r0 error
	if rf, ok := ret.Get(0).(func([]*Log) error); ok {
		r0 = rf(logs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}