package authdoor

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockImpl struct {
	mock.Mock
	t       *testing.T
	expStat AuthStatus
	expResp ResponseStatus
}

// mockAuthFunc will check to make sure it is returning the correct things
func (m *MockImpl) mockAuthFunc(w http.ResponseWriter, r *http.Request) (AuthStatus, ResponseStatus) {
	args := m.Called(w, r)
	actStat := args.Get(0).(AuthStatus)
	actResp := args.Get(1).(ResponseStatus)
	return actStat, actResp
}

func (m *MockImpl) RequireCalled() {
	m.AssertCalled(m.t, "mockAuthFunc", nil, (*http.Request)(nil))
}

func newMockAuthFunc(t *testing.T, stat AuthStatus, resp ResponseStatus) *MockImpl {
	mockImpl := &MockImpl{t: t, expStat: stat, expResp: resp}
	mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(stat, resp)
	return mockImpl
}

// testCases is a to-be-read-only global for tests that reflects all possible combinations of return values for an AuthFuncInstance/AuthFunc
var testCases = []struct {
	name     string
	expStat  AuthStatus
	expResp  ResponseStatus
	priority int
}{
	{"AuthFailed and Ignored", AuthFailed, Ignored, 0},
	{"AuthDenied and Ignored", AuthDenied, Ignored, 0},
	{"AuthGranted and Ignored", AuthGranted, Ignored, 0},
	{"AuthFailed and Answered", AuthFailed, Answered, 0},
	{"AuthDenied and Answered", AuthFailed, Answered, 0},
	{"AuthGranted and Answered", AuthFailed, Answered, 0},
}

// TestAuthFuncInstance tests the AuthFuncInstance object constructor and its methods
func TestNewAuthFuncInstance(t *testing.T) {
	for i, _ := range testCases {
		mockImpl := newMockAuthFunc(t, testCases[i].expStat, testCases[i].expResp)
		dut := new(AuthFuncInstance)
		dut.Init(testCases[i].name, mockImpl.mockAuthFunc, testCases[i].priority, nil)

		require.IsType(t, &AuthFuncInstance{}, dut)           // we created a new AuthFunc
		require.Equal(t, testCases[i].name, dut.name)         // it has the proper value
		require.Equal(t, testCases[i].priority, dut.priority) // it has the proper value
	}
}
func TestAuthFuncInstanceCall(t *testing.T) {
	for i, _ := range testCases {
		mockImpl := newMockAuthFunc(t, testCases[i].expStat, testCases[i].expResp)
		dut := new(AuthFuncInstance)
		dut.Init(testCases[i].name, mockImpl.mockAuthFunc, testCases[i].priority, nil)
		status, response := dut.call(nil, nil)           // it is called
		mockImpl.RequireCalled()                         // the authfunc we passed is called
		require.Equal(t, testCases[i].expStat, status)   // the result we expect and received
		require.Equal(t, testCases[i].expResp, response) // the result we expect and received
	}
}

// blankAuthFunc is an authfunc that does nothing but return. This is different than the mockAuthFunc because we cannot check if it has been called, and we can set it's return values.
func blankAuthFunc(w http.ResponseWriter, r *http.Request) (AuthStatus, ResponseStatus) {
	return AuthDenied, Ignored
}

// BenchmarkNewAuthFuncInstance will test how fast and efficient we are in creating new `AuthFuncInstance`s
func BenchmarkNewAuthFuncInstance(b *testing.B) {
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
