package authdoor

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockImpl struct {
	mock.Mock
	t      testing.TB
	expRet AuthFuncReturn
	expErr error
}

// mockAuthFunc will check to make sure it is returning the correct things
func (m *MockImpl) mockAuthFunc(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error) {
	args := m.Called(w, r)
	expRet := args.Get(0).(AuthFuncReturn)
	expErr := args.Error(1)
	return expRet, expErr
}

// This should have just taken t.
func (m *MockImpl) RequireCalled() {
	m.AssertCalled(m.t, "mockAuthFunc", nil, (*http.Request)(nil))
}

func newMockAuthFunc(t testing.TB, expRet AuthFuncReturn, expErr error) *MockImpl {
	mockImpl := &MockImpl{t: t, expRet: expRet, expErr: expErr}
	mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(expRet, expErr)
	return mockImpl
}

// testInstancesRaw is a to-be-read-only global for tests that reflects all possible combinations of return values for an AuthFuncInstance/AuthFunc
var testInstancesRaw = []struct {
	name     string
	priority int
	expRet   AuthFuncReturn
	expErr   error
}{
	{
		"AuthFailed and Ignored",
		0,
		AuthFuncReturn{
			Auth: AuthFailed,
			Resp: Ignored,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
	{
		"AuthDenied and Ignored",
		0,
		AuthFuncReturn{
			Auth: AuthDenied,
			Resp: Ignored,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
	{
		"AuthGranted and Ignored",
		0,
		AuthFuncReturn{
			Auth: AuthGranted,
			Resp: Ignored,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
	{"AuthFailed and Answered",
		0,
		AuthFuncReturn{
			Auth: AuthFailed,
			Resp: Answered,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
	{
		"AuthDenied and Answered",
		0,
		AuthFuncReturn{
			Auth: AuthFailed,
			Resp: Answered,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
	{
		"AuthGranted and Answered",
		0,
		AuthFuncReturn{
			Auth: AuthFailed,
			Resp: Answered,
			Info: InstanceReturnInfo{Info: make(json.RawMessage, 1)},
		},
		nil,
	},
}

// TestInit tests the AuthFuncInstance object constructor
func TestAuthFuncInstanceInit(t *testing.T) {
	for i := range testInstancesRaw {
		mockImpl := newMockAuthFunc(t, testInstancesRaw[i].expRet, testInstancesRaw[i].expErr)
		dut := new(AuthFuncInstance)
		dut.Init(testInstancesRaw[i].name, mockImpl.mockAuthFunc, testInstancesRaw[i].priority, nil)

		require.IsType(t, &AuthFuncInstance{}, dut)                  // we created a new AuthFunc
		require.Equal(t, testInstancesRaw[i].name, dut.name)         // it has the proper value
		require.Equal(t, testInstancesRaw[i].priority, dut.priority) // it has the proper value
	}
}

// TestCall tests the AuthFuncInstance call
func TestAuthFuncInstanceCall(t *testing.T) {
	for i := range testInstancesRaw {
		mockImpl := newMockAuthFunc(t, testInstancesRaw[i].expRet, testInstancesRaw[i].expErr)
		dut := new(AuthFuncInstance)
		dut.Init(testInstancesRaw[i].name, mockImpl.mockAuthFunc, testInstancesRaw[i].priority, nil)
		info, err := dut.call(nil, nil) // it is called
		require.Equal(t, testInstancesRaw[i].expErr, err)
		mockImpl.RequireCalled() // the authfunc we passed is called
		require.Equal(t, testInstancesRaw[i].expRet.Auth, info.Auth)
		require.Equal(t, testInstancesRaw[i].expRet.Resp, info.Resp)
		require.Equal(t, testInstancesRaw[i].name, info.Info.name)
	}
}

// blankAuthFunc is an authfunc that does nothing but return. This is different than the mockAuthFunc because we cannot check if it has been called, and we can set it's return values.
func blankAuthFunc(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error) {
	return AuthFuncReturn{
		Auth: AuthDenied,
		Resp: Ignored,
		Info: InstanceReturnInfo{name: "blank"},
	}, nil
}

// BenchmarkNewAuthFuncInstance will test how fast and efficient we are in creating new `AuthFuncInstance`s
func BenchmarkAuthFuncInstanceInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dut := new(AuthFuncInstance)
		dut.Init("benchmark", blankAuthFunc, 0, &EmptyLogger{})
	}
}

// BenchmarkAuthFuncInstanceCall tests how fast and efficiently we call the authfunc member of the struct
func BenchmarkAuthFuncInstanceCall(b *testing.B) {
	dut := new(AuthFuncInstance)
	dut.Init("benchmark", blankAuthFunc, 0, &EmptyLogger{})
	for i := 0; i < b.N; i++ {
		dut.call(nil, nil)
	}
}
