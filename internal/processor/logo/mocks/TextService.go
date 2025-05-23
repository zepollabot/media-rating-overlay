// Code generated by mockery. DO NOT EDIT.

package logo_mocks

import (
	mock "github.com/stretchr/testify/mock"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

// TextService is an autogenerated mock type for the TextService type
type TextService struct {
	mock.Mock
}

type TextService_Expecter struct {
	mock *mock.Mock
}

func (_m *TextService) EXPECT() *TextService_Expecter {
	return &TextService_Expecter{mock: &_m.Mock}
}

// GetText provides a mock function with given fields: imageObject, areaWidth, areaHeight, text, numberOfDigits
func (_m *TextService) GetText(imageObject model.Image, areaWidth float64, areaHeight float64, text string, numberOfDigits int) (model.Text, error) {
	ret := _m.Called(imageObject, areaWidth, areaHeight, text, numberOfDigits)

	if len(ret) == 0 {
		panic("no return value specified for GetText")
	}

	var r0 model.Text
	var r1 error
	if rf, ok := ret.Get(0).(func(model.Image, float64, float64, string, int) (model.Text, error)); ok {
		return rf(imageObject, areaWidth, areaHeight, text, numberOfDigits)
	}
	if rf, ok := ret.Get(0).(func(model.Image, float64, float64, string, int) model.Text); ok {
		r0 = rf(imageObject, areaWidth, areaHeight, text, numberOfDigits)
	} else {
		r0 = ret.Get(0).(model.Text)
	}

	if rf, ok := ret.Get(1).(func(model.Image, float64, float64, string, int) error); ok {
		r1 = rf(imageObject, areaWidth, areaHeight, text, numberOfDigits)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TextService_GetText_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetText'
type TextService_GetText_Call struct {
	*mock.Call
}

// GetText is a helper method to define mock.On call
//   - imageObject model.Image
//   - areaWidth float64
//   - areaHeight float64
//   - text string
//   - numberOfDigits int
func (_e *TextService_Expecter) GetText(imageObject interface{}, areaWidth interface{}, areaHeight interface{}, text interface{}, numberOfDigits interface{}) *TextService_GetText_Call {
	return &TextService_GetText_Call{Call: _e.mock.On("GetText", imageObject, areaWidth, areaHeight, text, numberOfDigits)}
}

func (_c *TextService_GetText_Call) Run(run func(imageObject model.Image, areaWidth float64, areaHeight float64, text string, numberOfDigits int)) *TextService_GetText_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(model.Image), args[1].(float64), args[2].(float64), args[3].(string), args[4].(int))
	})
	return _c
}

func (_c *TextService_GetText_Call) Return(_a0 model.Text, _a1 error) *TextService_GetText_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *TextService_GetText_Call) RunAndReturn(run func(model.Image, float64, float64, string, int) (model.Text, error)) *TextService_GetText_Call {
	_c.Call.Return(run)
	return _c
}

// NewTextService creates a new instance of TextService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTextService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TextService {
	mock := &TextService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
