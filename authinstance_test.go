package authdoor

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ayjayt/ilog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockImpl is a structure providing shim functions that cna be checked to see if they were called
type MockImpl struct {
	mock.Mock
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

// RequireCalled will assert that function was called in testing
func (m *MockImpl) RequireCalled(tb testing.TB) {
	m.AssertCalled(tb, "mockAuthFunc", nil, (*http.Request)(nil))
}

// newMockAuthFunc is a generator for fake AuthFuncs
func newMockAuthFunc(expRet AuthFuncReturn, expErr error) *MockImpl {
	mockImpl := &MockImpl{expRet: expRet, expErr: expErr}
	mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(expRet, expErr)
	return mockImpl
}

// testInstancesRaw is a data table used to form authfuncs and authinstances for use in testing.
type testInstancesRaw struct {
	name     string
	priority int
	expRet   AuthFuncReturn
	expErr   error
}

// allReturnCombos is an instance of testInstancesRaw that provides all possible return values of an AuthFunc
var allReturnCombos = []testInstancesRaw{
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
	for i := range allReturnCombos {
		mockImpl := newMockAuthFunc(allReturnCombos[i].expRet, allReturnCombos[i].expErr)
		dut := new(AuthFuncInstance)
		dut.Init(allReturnCombos[i].name, mockImpl.mockAuthFunc, allReturnCombos[i].priority, nil)

		require.IsType(t, &AuthFuncInstance{}, dut)                 // we created a new AuthFunc
		require.Equal(t, allReturnCombos[i].name, dut.name)         // it has the proper value
		require.Equal(t, allReturnCombos[i].priority, dut.priority) // it has the proper value
	}
}

// TestCall tests the AuthFuncInstance call
func TestAuthFuncInstanceCall(t *testing.T) {
	for i := range allReturnCombos {
		mockImpl := newMockAuthFunc(allReturnCombos[i].expRet, allReturnCombos[i].expErr)
		dut := new(AuthFuncInstance)
		dut.Init(allReturnCombos[i].name, mockImpl.mockAuthFunc, allReturnCombos[i].priority, nil)
		info, err := dut.call(nil, nil) // it is called
		require.Equal(t, allReturnCombos[i].expErr, err)
		mockImpl.RequireCalled(t) // the authfunc we passed is called
		require.Equal(t, allReturnCombos[i].expRet.Auth, info.Auth)
		require.Equal(t, allReturnCombos[i].expRet.Resp, info.Resp)
		require.Equal(t, allReturnCombos[i].name, info.Info.name)
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
		dut.Init("benchmark", blankAuthFunc, 0, &ilog.EmptyLogger{})
	}
}

// BenchmarkAuthFuncInstanceCall tests how fast and efficiently we call the authfunc member of the struct
func BenchmarkAuthFuncInstanceCall(b *testing.B) {
	dut := new(AuthFuncInstance)
	dut.Init("benchmark", blankAuthFunc, 0, &ilog.EmptyLogger{})
	for i := 0; i < b.N; i++ {
		dut.call(nil, nil)
	}
}
