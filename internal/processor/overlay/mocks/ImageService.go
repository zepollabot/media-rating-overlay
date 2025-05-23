// Code generated by mockery. DO NOT EDIT.

package overlay_mocks

import (
	image "image"

	gg "github.com/fogleman/gg"

	mock "github.com/stretchr/testify/mock"
)

// ImageService is an autogenerated mock type for the ImageService type
type ImageService struct {
	mock.Mock
}

type ImageService_Expecter struct {
	mock *mock.Mock
}

func (_m *ImageService) EXPECT() *ImageService_Expecter {
	return &ImageService_Expecter{mock: &_m.Mock}
}

// CreateContext provides a mock function with given fields: width, height
func (_m *ImageService) CreateContext(width int, height int) *gg.Context {
	ret := _m.Called(width, height)

	if len(ret) == 0 {
		panic("no return value specified for CreateContext")
	}

	var r0 *gg.Context
	if rf, ok := ret.Get(0).(func(int, int) *gg.Context); ok {
		r0 = rf(width, height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gg.Context)
		}
	}

	return r0
}

// ImageService_CreateContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateContext'
type ImageService_CreateContext_Call struct {
	*mock.Call
}

// CreateContext is a helper method to define mock.On call
//   - width int
//   - height int
func (_e *ImageService_Expecter) CreateContext(width interface{}, height interface{}) *ImageService_CreateContext_Call {
	return &ImageService_CreateContext_Call{Call: _e.mock.On("CreateContext", width, height)}
}

func (_c *ImageService_CreateContext_Call) Run(run func(width int, height int)) *ImageService_CreateContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int), args[1].(int))
	})
	return _c
}

func (_c *ImageService_CreateContext_Call) Return(_a0 *gg.Context) *ImageService_CreateContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *ImageService_CreateContext_Call) RunAndReturn(run func(int, int) *gg.Context) *ImageService_CreateContext_Call {
	_c.Call.Return(run)
	return _c
}

// OpenImage provides a mock function with given fields: filePath
func (_m *ImageService) OpenImage(filePath string) (image.Image, error) {
	ret := _m.Called(filePath)

	if len(ret) == 0 {
		panic("no return value specified for OpenImage")
	}

	var r0 image.Image
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (image.Image, error)); ok {
		return rf(filePath)
	}
	if rf, ok := ret.Get(0).(func(string) image.Image); ok {
		r0 = rf(filePath)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(image.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filePath)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ImageService_OpenImage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OpenImage'
type ImageService_OpenImage_Call struct {
	*mock.Call
}

// OpenImage is a helper method to define mock.On call
//   - filePath string
func (_e *ImageService_Expecter) OpenImage(filePath interface{}) *ImageService_OpenImage_Call {
	return &ImageService_OpenImage_Call{Call: _e.mock.On("OpenImage", filePath)}
}

func (_c *ImageService_OpenImage_Call) Run(run func(filePath string)) *ImageService_OpenImage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *ImageService_OpenImage_Call) Return(_a0 image.Image, _a1 error) *ImageService_OpenImage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ImageService_OpenImage_Call) RunAndReturn(run func(string) (image.Image, error)) *ImageService_OpenImage_Call {
	_c.Call.Return(run)
	return _c
}

// NewImageService creates a new instance of ImageService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewImageService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ImageService {
	mock := &ImageService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
