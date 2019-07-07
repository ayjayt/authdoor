package authdoor

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockImpl struct {
	mock.Mock
}

func (m *MockImpl) mockAuthFunc(w http.ResponseWriter, r *http.Request) (AuthStatus, ResponseStatus) {
	args := m.Called(w, r)
	return AuthStatus(args.Int(0)), ResponseStatus(args.Bool(1))
}

// empty authfunc might be good for call and mock

func TestNewAuthFuncInstance(t *testing.T) {
	mockImpl := &MockImpl{}
	mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(1, true)
	dut := NewAuthFuncInstance("test_name", mockImpl.mockAuthFunc, 0)
	require.IsType(t, authFuncInstance{}, dut)
	require.Equal(t, "test_name", dut.name)
	require.Equal(t, 0, dut.priority)
	status, response := dut.call(nil, nil)
	require.Equal(t, AuthGranted, status)
	require.Equal(t, Answered, response)

	// probably a good time to test call
}

// This also tests WriteMap, Sort, and AddCallables
func TestAuthFuncListCore(t *testing.T) {
	// Priority should be the negative index for this to test sorting
	tt := []struct {
		name       string
		authstatus int
		respstatus bool
		priority   int
	}{
		{"AuthFailed w/o Response", int(AuthFailed), bool(Ignored), 0},
		{"AuthGranted w/o Response", int(AuthGranted), bool(Ignored), -1},
		{"AuthDenied w/o Response", int(AuthDenied), bool(Ignored), -2},
		{"AuthFailed w/ Response", int(AuthFailed), bool(Answered), -3},
		{"AuthGranted w/ Response", int(AuthGranted), bool(Answered), -4},
		{"AuthDenied w/ Response", int(AuthDenied), bool(Answered), -5},
	}
	tt2 := []struct {
		name       string
		authstatus int
		respstatus bool
		priority   int
	}{
		{"AuthFailed w/o Response2", int(AuthFailed), bool(Ignored), 0},
		{"AuthGranted w/o Response2", int(AuthGranted), bool(Ignored), -1},
		{"AuthDenied w/o Response2", int(AuthDenied), bool(Ignored), -2},
		{"AuthFailed w/ Response2", int(AuthFailed), bool(Answered), -3},
		{"AuthGranted w/ Response2", int(AuthGranted), bool(Answered), -4},
		{"AuthDenied w/ Response2", int(AuthDenied), bool(Answered), -5},
	}
	duts := make([]authFuncCallable, len(tt))
	duts2 := make([]authFuncCallable, len(tt2))

	for i, _ := range tt {
		mockImpl := &MockImpl{}
		mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(tt[i].authstatus, tt[i].respstatus)
		temp := NewAuthFuncInstance(tt[i].name, mockImpl.mockAuthFunc, tt[i].priority)
		duts = append(duts, &temp)
	}
	for i, _ := range tt2 {
		mockImpl := &MockImpl{}
		mockImpl.On("mockAuthFunc", nil, (*http.Request)(nil)).Return(tt2[i].authstatus, tt2[i].respstatus)
		temp := NewAuthFuncInstance(tt2[i].name, mockImpl.mockAuthFunc, tt2[i].priority)
		duts2 = append(duts2, &temp)
	}
	// now we have an array of NewAuthFunInstances, and we're going to expload it into new
	dut, err := newAuthFuncListCore("dut", duts...)
	require.NoError(t, err)
	for i, _ := range tt {
		t.Logf("Name: %v, Priority: %v, Index: %v, Calced-Index: %v", tt[i].name, tt[i].priority, dut.funcMap[tt[i].name], len(tt)-1+tt[i].priority)
		// Check to see if the map translates to priority properly (-5 == 0)
		// The first argument basically "inverts" the priority
		// The second argument provides the index
		require.Equal(t, len(tt)-1+tt[i].priority, dut.funcMap[tt[i].name])
		// Check to see if the names come out in reverse order
		// The first argument inverts the tt order
		// the second argument provides the name
		require.Equal(t, tt[len(tt)-1-i].name, dut.funcList[i].name)
	}
	err = dut.AddCallables(&dut) // This should return an error, no?
	require.Error(t, ErrNameTaken, err)
	t.Logf("Error: %v", err)
	err = dut.AddCallables(duts2...) // This should return an error, no?
	lastPriority := -100000
	for i, _ := range dut.funcList {
		require.GreaterOrEqual(t, dut.funcList[i].priority, lastPriority)
		lastPriority = dut.funcList[i].priority
		require.Equal(t, i, dut.funcMap[dut.funcList[i].name])
	}
	// We want to check each return value - is what's returned what's expected
	// And then maybe we can remove it by name?
	// TODO: this is a serious hole that needs to be evaluated
	// RemoveCallables and test
	var names []string
	for _, v := range tt2 {
		names = append(names, v.name)
	}
	dut.RemoveCallables(names...)
	// TODO: can't copy dut, the Core
	for i, _ := range tt {
		t.Logf("Name: %v, Priority: %v, Index: %v, Calced-Index: %v", tt[i].name, tt[i].priority, dut.funcMap[tt[i].name], len(tt)-1+tt[i].priority)
		// Check to see if the map translates to priority properly (-5 == 0)
		// The first argument basically "inverts" the priority
		// The second argument provides the index
		require.Equal(t, len(tt)-1+tt[i].priority, dut.funcMap[tt[i].name])
		// Check to see if the names come out in reverse order
		// The first argument inverts the tt order
		// the second argument provides the name
		require.Equal(t, tt[len(tt)-1-i].name, dut.funcList[i].name)
	}
	dut2, err := NewAuthFuncList("dut2", &dut)
	for i, _ := range tt {
		t.Logf("Name: %v, Priority: %v, Index: %v, Calced-Index: %v", tt[i].name, tt[i].priority, dut2.funcMap[tt[i].name], len(tt)-1+tt[i].priority)
		// Check to see if the map translates to priority properly (-5 == 0)
		// The first argument basically "inverts" the priority
		// The second argument provides the index
		require.Equal(t, len(tt)-1+tt[i].priority, dut2.funcMap[tt[i].name])
		// Check to see if the names come out in reverse order
		// The first argument inverts the tt order
		// the second argument provides the name
		require.Equal(t, tt[len(tt)-1-i].name, dut2.funcList[i].name)
	}
	base := http.FileServer(http.Dir("/tmp"))
	sampleHandler := NewAuthHandler(base)
	require.Equal(t, base, sampleHandler.GetBase())
	sampleHandler.SetBase(nil)
	require.Equal(t, nil, sampleHandler.GetBase())
	sampleHandler.SetBase(base)

	dut2.addHandler(sampleHandler)
	require.Equal(t, len(dut2.handlers), 1)
	dut2.removeHandler(sampleHandler)
	require.Equal(t, len(dut2.handlers), 0)

	sampleHandler.AddInstances(duts2...)
}

// TestUpdateHandlers <-- broken until UpdateActiveList is fixed
// TestBlockForUpdate 3:00pm <-- broken until UpdateActiveList is fixed
// TestNewAuthHandler <-- broken until UpdateActiveList is fixed
// TestAddInstances -- do
// TestRemoveInstances -- d0
// TestUpdateHandler 4:00 pm <-- broken until UpdateActiveList is fixed
// TestAddLists -- do
// TestRemoveLists -- do
// TestStartLock 4:30 pm // TODO HARD
// TestEndLock // TODO HARD
// TestUpdateActiveList <-- broken
// TestServeHTTP 5:00 pm // do
