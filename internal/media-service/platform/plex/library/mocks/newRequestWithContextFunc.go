// Code generated by mockery. DO NOT EDIT.

package plex_mocks

import (
	context "context"
	http "net/http"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// newRequestWithContextFunc is an autogenerated mock type for the newRequestWithContextFunc type
type newRequestWithContextFunc struct {
	mock.Mock
}

type newRequestWithContextFunc_Expecter struct {
	mock *mock.Mock
}

func (_m *newRequestWithContextFunc) EXPECT() *newRequestWithContextFunc_Expecter {
	return &newRequestWithContextFunc_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, method, url, body
func (_m *newRequestWithContextFunc) Execute(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	ret := _m.Called(ctx, method, url, body)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 *http.Request
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, io.Reader) (*http.Request, error)); ok {
		return rf(ctx, method, url, body)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, io.Reader) *http.Request); ok {
		r0 = rf(ctx, method, url, body)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Request)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, io.Reader) error); ok {
		r1 = rf(ctx, method, url, body)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// newRequestWithContextFunc_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type newRequestWithContextFunc_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - method string
//   - url string
//   - body io.Reader
func (_e *newRequestWithContextFunc_Expecter) Execute(ctx interface{}, method interface{}, url interface{}, body interface{}) *newRequestWithContextFunc_Execute_Call {
	return &newRequestWithContextFunc_Execute_Call{Call: _e.mock.On("Execute", ctx, method, url, body)}
}

func (_c *newRequestWithContextFunc_Execute_Call) Run(run func(ctx context.Context, method string, url string, body io.Reader)) *newRequestWithContextFunc_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string), args[3].(io.Reader))
	})
	return _c
}

func (_c *newRequestWithContextFunc_Execute_Call) Return(_a0 *http.Request, _a1 error) *newRequestWithContextFunc_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *newRequestWithContextFunc_Execute_Call) RunAndReturn(run func(context.Context, string, string, io.Reader) (*http.Request, error)) *newRequestWithContextFunc_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// newNewRequestWithContextFunc creates a new instance of newRequestWithContextFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newNewRequestWithContextFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *newRequestWithContextFunc {
	mock := &newRequestWithContextFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
