// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"github.com/argoproj/argo-cd/v3/util/workloadidentity"
	mock "github.com/stretchr/testify/mock"
)

// NewTokenProvider creates a new instance of TokenProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTokenProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *TokenProvider {
	mock := &TokenProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// TokenProvider is an autogenerated mock type for the TokenProvider type
type TokenProvider struct {
	mock.Mock
}

type TokenProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *TokenProvider) EXPECT() *TokenProvider_Expecter {
	return &TokenProvider_Expecter{mock: &_m.Mock}
}

// GetToken provides a mock function for the type TokenProvider
func (_mock *TokenProvider) GetToken(scope string) (*workloadidentity.Token, error) {
	ret := _mock.Called(scope)

	if len(ret) == 0 {
		panic("no return value specified for GetToken")
	}

	var r0 *workloadidentity.Token
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(string) (*workloadidentity.Token, error)); ok {
		return returnFunc(scope)
	}
	if returnFunc, ok := ret.Get(0).(func(string) *workloadidentity.Token); ok {
		r0 = returnFunc(scope)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*workloadidentity.Token)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(string) error); ok {
		r1 = returnFunc(scope)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// TokenProvider_GetToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetToken'
type TokenProvider_GetToken_Call struct {
	*mock.Call
}

// GetToken is a helper method to define mock.On call
//   - scope string
func (_e *TokenProvider_Expecter) GetToken(scope interface{}) *TokenProvider_GetToken_Call {
	return &TokenProvider_GetToken_Call{Call: _e.mock.On("GetToken", scope)}
}

func (_c *TokenProvider_GetToken_Call) Run(run func(scope string)) *TokenProvider_GetToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 string
		if args[0] != nil {
			arg0 = args[0].(string)
		}
		run(
			arg0,
		)
	})
	return _c
}

func (_c *TokenProvider_GetToken_Call) Return(token *workloadidentity.Token, err error) *TokenProvider_GetToken_Call {
	_c.Call.Return(token, err)
	return _c
}

func (_c *TokenProvider_GetToken_Call) RunAndReturn(run func(scope string) (*workloadidentity.Token, error)) *TokenProvider_GetToken_Call {
	_c.Call.Return(run)
	return _c
}
