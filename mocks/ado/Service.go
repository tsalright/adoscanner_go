// Code generated by mockery v2.0.3. DO NOT EDIT.

package mocks

import (
	core "github.com/microsoft/azure-devops-go-api/azuredevops/core"
	git "github.com/microsoft/azure-devops-go-api/azuredevops/git"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// CreateConnection provides a mock function with given fields: orgURL, pat
func (_m *Service) CreateConnection(orgURL string, pat string) error {
	ret := _m.Called(orgURL, pat)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(orgURL, pat)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAdditionalProjects provides a mock function with given fields: continuationToken
func (_m *Service) GetAdditionalProjects(continuationToken string) (*core.GetProjectsResponseValue, error) {
	ret := _m.Called(continuationToken)

	var r0 *core.GetProjectsResponseValue
	if rf, ok := ret.Get(0).(func(string) *core.GetProjectsResponseValue); ok {
		r0 = rf(continuationToken)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.GetProjectsResponseValue)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(continuationToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetItemContent provides a mock function with given fields: projectName, repoName, path
func (_m *Service) GetItemContent(projectName string, repoName string, path string) (io.ReadCloser, error) {
	ret := _m.Called(projectName, repoName, path)

	var r0 io.ReadCloser
	if rf, ok := ret.Get(0).(func(string, string, string) io.ReadCloser); ok {
		r0 = rf(projectName, repoName, path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(projectName, repoName, path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetItems provides a mock function with given fields: projectName, repoName
func (_m *Service) GetItems(projectName string, repoName string) (*[]git.GitItem, error) {
	ret := _m.Called(projectName, repoName)

	var r0 *[]git.GitItem
	if rf, ok := ret.Get(0).(func(string, string) *[]git.GitItem); ok {
		r0 = rf(projectName, repoName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]git.GitItem)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(projectName, repoName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetProjects provides a mock function with given fields:
func (_m *Service) GetProjects() (*core.GetProjectsResponseValue, error) {
	ret := _m.Called()

	var r0 *core.GetProjectsResponseValue
	if rf, ok := ret.Get(0).(func() *core.GetProjectsResponseValue); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.GetProjectsResponseValue)
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

// GetRepositories provides a mock function with given fields: projectName
func (_m *Service) GetRepositories(projectName string) (*[]git.GitRepository, error) {
	ret := _m.Called(projectName)

	var r0 *[]git.GitRepository
	if rf, ok := ret.Get(0).(func(string) *[]git.GitRepository); ok {
		r0 = rf(projectName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]git.GitRepository)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(projectName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}