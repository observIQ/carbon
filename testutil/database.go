// Code generated by mockery v1.0.0. DO NOT EDIT.

package testutil

import (
	mock "github.com/stretchr/testify/mock"
	bbolt "go.etcd.io/bbolt"
)

// Database is an autogenerated mock type for the Database type
type Database struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Database) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Sync provides a mock function with given fields:
func (_m *Database) Sync() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0
func (_m *Database) Update(_a0 func(*bbolt.Tx) error) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(func(*bbolt.Tx) error) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// View provides a mock function with given fields: _a0
func (_m *Database) View(_a0 func(*bbolt.Tx) error) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(func(*bbolt.Tx) error) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
