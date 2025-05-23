// Code generated by mockery. DO NOT EDIT.

package core_mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// SemaphoreWeighted is an autogenerated mock type for the SemaphoreWeighted type
type SemaphoreWeighted struct {
	mock.Mock
}

type SemaphoreWeighted_Expecter struct {
	mock *mock.Mock
}

func (_m *SemaphoreWeighted) EXPECT() *SemaphoreWeighted_Expecter {
	return &SemaphoreWeighted_Expecter{mock: &_m.Mock}
}

// Acquire provides a mock function with given fields: ctx, n
func (_m *SemaphoreWeighted) Acquire(ctx context.Context, n int64) error {
	ret := _m.Called(ctx, n)

	if len(ret) == 0 {
		panic("no return value specified for Acquire")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64) error); ok {
		r0 = rf(ctx, n)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SemaphoreWeighted_Acquire_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Acquire'
type SemaphoreWeighted_Acquire_Call struct {
	*mock.Call
}

// Acquire is a helper method to define mock.On call
//   - ctx context.Context
//   - n int64
func (_e *SemaphoreWeighted_Expecter) Acquire(ctx interface{}, n interface{}) *SemaphoreWeighted_Acquire_Call {
	return &SemaphoreWeighted_Acquire_Call{Call: _e.mock.On("Acquire", ctx, n)}
}

func (_c *SemaphoreWeighted_Acquire_Call) Run(run func(ctx context.Context, n int64)) *SemaphoreWeighted_Acquire_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64))
	})
	return _c
}

func (_c *SemaphoreWeighted_Acquire_Call) Return(_a0 error) *SemaphoreWeighted_Acquire_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SemaphoreWeighted_Acquire_Call) RunAndReturn(run func(context.Context, int64) error) *SemaphoreWeighted_Acquire_Call {
	_c.Call.Return(run)
	return _c
}

// Release provides a mock function with given fields: n
func (_m *SemaphoreWeighted) Release(n int64) {
	_m.Called(n)
}

// SemaphoreWeighted_Release_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Release'
type SemaphoreWeighted_Release_Call struct {
	*mock.Call
}

// Release is a helper method to define mock.On call
//   - n int64
func (_e *SemaphoreWeighted_Expecter) Release(n interface{}) *SemaphoreWeighted_Release_Call {
	return &SemaphoreWeighted_Release_Call{Call: _e.mock.On("Release", n)}
}

func (_c *SemaphoreWeighted_Release_Call) Run(run func(n int64)) *SemaphoreWeighted_Release_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64))
	})
	return _c
}

func (_c *SemaphoreWeighted_Release_Call) Return() *SemaphoreWeighted_Release_Call {
	_c.Call.Return()
	return _c
}

func (_c *SemaphoreWeighted_Release_Call) RunAndReturn(run func(int64)) *SemaphoreWeighted_Release_Call {
	_c.Run(run)
	return _c
}

// TryAcquire provides a mock function with given fields: n
func (_m *SemaphoreWeighted) TryAcquire(n int64) bool {
	ret := _m.Called(n)

	if len(ret) == 0 {
		panic("no return value specified for TryAcquire")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(int64) bool); ok {
		r0 = rf(n)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// SemaphoreWeighted_TryAcquire_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TryAcquire'
type SemaphoreWeighted_TryAcquire_Call struct {
	*mock.Call
}

// TryAcquire is a helper method to define mock.On call
//   - n int64
func (_e *SemaphoreWeighted_Expecter) TryAcquire(n interface{}) *SemaphoreWeighted_TryAcquire_Call {
	return &SemaphoreWeighted_TryAcquire_Call{Call: _e.mock.On("TryAcquire", n)}
}

func (_c *SemaphoreWeighted_TryAcquire_Call) Run(run func(n int64)) *SemaphoreWeighted_TryAcquire_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(int64))
	})
	return _c
}

func (_c *SemaphoreWeighted_TryAcquire_Call) Return(_a0 bool) *SemaphoreWeighted_TryAcquire_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SemaphoreWeighted_TryAcquire_Call) RunAndReturn(run func(int64) bool) *SemaphoreWeighted_TryAcquire_Call {
	_c.Call.Return(run)
	return _c
}

// NewSemaphoreWeighted creates a new instance of SemaphoreWeighted. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSemaphoreWeighted(t interface {
	mock.TestingT
	Cleanup(func())
}) *SemaphoreWeighted {
	mock := &SemaphoreWeighted{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
