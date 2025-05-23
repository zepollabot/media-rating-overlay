// Code generated by mockery. DO NOT EDIT.

package core_mocks

import (
	mock "github.com/stretchr/testify/mock"
	model "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
)

// MediaServiceModelFactory is an autogenerated mock type for the MediaServiceModelFactory type
type MediaServiceModelFactory struct {
	mock.Mock
}

type MediaServiceModelFactory_Expecter struct {
	mock *mock.Mock
}

func (_m *MediaServiceModelFactory) EXPECT() *MediaServiceModelFactory_Expecter {
	return &MediaServiceModelFactory_Expecter{mock: &_m.Mock}
}

// Create provides a mock function with given fields: mediaServiceType
func (_m *MediaServiceModelFactory) Create(mediaServiceType string) (model.MediaService, error) {
	ret := _m.Called(mediaServiceType)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 model.MediaService
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (model.MediaService, error)); ok {
		return rf(mediaServiceType)
	}
	if rf, ok := ret.Get(0).(func(string) model.MediaService); ok {
		r0 = rf(mediaServiceType)
	} else {
		r0 = ret.Get(0).(model.MediaService)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(mediaServiceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MediaServiceModelFactory_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type MediaServiceModelFactory_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - mediaServiceType string
func (_e *MediaServiceModelFactory_Expecter) Create(mediaServiceType interface{}) *MediaServiceModelFactory_Create_Call {
	return &MediaServiceModelFactory_Create_Call{Call: _e.mock.On("Create", mediaServiceType)}
}

func (_c *MediaServiceModelFactory_Create_Call) Run(run func(mediaServiceType string)) *MediaServiceModelFactory_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MediaServiceModelFactory_Create_Call) Return(_a0 model.MediaService, _a1 error) *MediaServiceModelFactory_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MediaServiceModelFactory_Create_Call) RunAndReturn(run func(string) (model.MediaService, error)) *MediaServiceModelFactory_Create_Call {
	_c.Call.Return(run)
	return _c
}

// NewMediaServiceModelFactory creates a new instance of MediaServiceModelFactory. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMediaServiceModelFactory(t interface {
	mock.TestingT
	Cleanup(func())
}) *MediaServiceModelFactory {
	mock := &MediaServiceModelFactory{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
