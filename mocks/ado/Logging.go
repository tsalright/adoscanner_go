// Code generated by mockery v2.0.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Logging is an autogenerated mock type for the Logging type
type Logging struct {
	mock.Mock
}

// LogError provides a mock function with given fields: err
func (_m *Logging) LogError(err error) {
	_m.Called(err)
}

// LogFatal provides a mock function with given fields: err
func (_m *Logging) LogFatal(err error) {
	_m.Called(err)
}

// LogInfo provides a mock function with given fields: msg
func (_m *Logging) LogInfo(msg string) {
	_m.Called(msg)
}

// LogWarning provides a mock function with given fields: msg
func (_m *Logging) LogWarning(msg string) {
	_m.Called(msg)
}
