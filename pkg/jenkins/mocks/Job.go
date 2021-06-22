// Code generated by mockery v2.7.5. DO NOT EDIT.

package mocks

import (
	context "context"

	gojenkins "github.com/bndr/gojenkins"

	mock "github.com/stretchr/testify/mock"
)

// Job is an autogenerated mock type for the Job type
type Job struct {
	mock.Mock
}

// GetJob provides a mock function with given fields:
func (_m *Job) GetJob() *gojenkins.Job {
	ret := _m.Called()

	var r0 *gojenkins.Job
	if rf, ok := ret.Get(0).(func() *gojenkins.Job); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gojenkins.Job)
		}
	}

	return r0
}

// InvokeSimple provides a mock function with given fields: ctx, params
func (_m *Job) InvokeSimple(ctx context.Context, params map[string]string) (int64, error) {
	ret := _m.Called(ctx, params)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, map[string]string) int64); ok {
		r0 = rf(ctx, params)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, map[string]string) error); ok {
		r1 = rf(ctx, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Poll provides a mock function with given fields: _a0
func (_m *Job) Poll(_a0 context.Context) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
