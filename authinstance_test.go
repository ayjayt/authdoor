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

// TestAuthFuncInstance tests the AuthFuncInstance object constructor and its methods
func TestAuthFuncInstance(t *testing.T) {
	testCases := []struct {
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
	for i, _ := range testCases {
		mockImpl := newMockAuthFunc(t, testCases[i].expStat, testCases[i].expResp)
		dut := NewAuthFuncInstance(testCases[i].name, mockImpl.mockAuthFunc, testCases[i].priority)

		require.IsType(t, authFuncInstance{}, dut)            // we created a new authFunc
		require.Equal(t, testCases[i].name, dut.name)         // it has the proper value
		require.Equal(t, testCases[i].priority, dut.priority) // it has the proper value

		status, response := dut.call(nil, nil)           // it is called
		mockImpl.RequireCalled()                         // the authfunc we passed is called
		require.Equal(t, testCases[i].expStat, status)   // the result we expect and received
		require.Equal(t, testCases[i].expResp, response) // the result we expect and received
	}
}
