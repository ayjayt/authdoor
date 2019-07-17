package authdoor

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

var seed = rand.NewSource(time.Now().UnixNano())
var seededRand = rand.New(seed)

// I have a an array that may or may not get sorted away
// We know what order it should get calld i
type testInstancesSortableRaw struct {
	name     string
	priority int
	expRet   AuthFuncReturn
	expErr   error
}

// Negative and ascending to -1 is important
var sortableInstances = []testInstancesSortableRaw{
	{"Alpha", -15, AuthFuncReturn{Auth: AuthFailed, Resp: Ignored}, nil},
	{"Beta", -14, AuthFuncReturn{Auth: AuthDenied, Resp: Ignored}, nil},
	{"Gamma", -13, AuthFuncReturn{Auth: AuthGranted, Resp: Ignored}, nil},
	{"Delta", -12, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Epsilon", -11, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Zeta", -10, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Eta", -9, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Theta", -8, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Iota", -7, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Kappa", -6, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Lambda", -5, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Mu", -4, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Nu", -3, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Xi", -2, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
	{"Omicron", -1, AuthFuncReturn{Auth: AuthFailed, Resp: Answered}, nil},
}

// makeInstances takes our test table and turns it into instances.
func makeInstances(t testing.TB, raw []testInstancesSortableRaw) ([]AuthFuncInstance, []*MockImpl) {
	mockImpls := make([]*MockImpl, len(sortableInstances))
	authInstances := make([]AuthFuncInstance, len(sortableInstances))
	for i := range raw {
		dut := AuthFuncInstance{}
		mockImpl := newMockAuthFunc(t, raw[i].expRet, raw[i].expErr)
		dut.Init(raw[i].name, mockImpl.mockAuthFunc, raw[i].priority, nil)
		mockImpls[i] = mockImpl
		authInstances[i] = dut
	}
	return authInstances, mockImpls
}

// checkOrder determines if the instance list is in the correct order
func checkOrder(list *AuthFuncList) (bool, AuthFuncInstance) {
	for i, v := range list.funcList {
		// Let's make sure it's mapped properly
		if i != list.funcMap[v.name] {
			return false, list.funcList[i]
		}
		if i != 0 && (list.funcList[i].priority <= list.funcList[i-1].priority) {
			return false, list.funcList[i]
		}
	}
	return true, AuthFuncInstance{}
}

// TestAuthFunc
func TestAuthFuncListInit(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
	for i := range list.funcList {
		// This wouldn't work if priorities weren't distinct
		require.Equal(t, list.funcList[i].name, sortableInstances[i].name)
	}
}

// Test sort
func TestAuthFuncListSort(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
	for j := 0; (j < 20) && (ordered == true); j++ {
		for i := 0; i < seededRand.Intn(200); i++ {
			list.Swap(seededRand.Intn(15), seededRand.Intn(15))
		}
	}
	ordered, _ = checkOrder(list)
	require.False(t, ordered, "list was ordered after attempting to randomize")
	list.sort()
	ordered, errorList = checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
}

// AuthFuncList will test one call
func TestAuthFuncListCall(t *testing.T) {
	instances, mocks := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
	for i, v := range sortableInstances {
		ret, err := list.Call(nil, nil, v.name)
		require.Equal(t, v.expErr, err)
		require.Equal(t, v.expRet.Auth, ret.Auth)
		require.Equal(t, v.expRet.Resp, ret.Resp)
		mocks[i].RequireCalled()
	}
}

func TestAuthFuncListCallAll(t *testing.T) {
	instances, mocks := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
	ret, err := list.CallAll(nil, nil)
	// One in our test table must return this!
	require.Nil(t, err)
	if ret.Auth == AuthFailed && ret.Resp == Ignored {
		require.FailNowf(t, "Unexpected return from CallAll.", "Returned: %v, %v", ret.Auth, ret.Resp) // I don't like how we test this
	}
	for i := 0; i < list.funcMap[ret.Info.name]; i++ {
		mocks[i].RequireCalled()
	}
}

func TestAuthFuncListAddInstances(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init()
	require.Equal(t, 0, len(list.funcList))
	require.Equal(t, 0, len(list.funcMap))
	list.AddInstances(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
}

func TestAuthFuncListRemoveInstances(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	ordered, errorList := checkOrder(list)
	require.True(t, ordered, "list value returned: %v", errorList)
	for _, v := range instances {
		list.RemoveInstances(v.name)
	}
	require.Equal(t, 0, len(list.funcList))
	require.Equal(t, 0, len(list.funcMap))
}

func BenchmarkAuthFuncListInit(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	b.ResetTimer()
	b.Run("Init", func(b *testing.B) {
		for i := 0; i <= b.N; i++ {
			list := new(AuthFuncList)
			list.Init(instances...)
		}
	})
}

func BenchmarkAuthFuncListSort(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	b.Run("Randomizer", func(b *testing.B) {
		for i := 0; i <= b.N; i++ {
			for i := 0; i < seededRand.Intn(200); i++ {
				list.Swap(seededRand.Intn(15), seededRand.Intn(15))
			}
		}
	})
	b.Run("Sort (w/ randomizer)", func(b *testing.B) {
		for i := 0; i <= b.N; i++ {
			for i := 0; i < seededRand.Intn(200); i++ {
				list.Swap(seededRand.Intn(15), seededRand.Intn(15))
			}
			list.sort()
		}
	})
}

// AuthFuncList will test one call
func BenchmarkAuthFuncListCall(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.Call(nil, nil, instances[0].name)
	}
}

// CallAll
func BenchmarkAuthFuncListCallAll(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.CallAll(nil, nil)
	}
}

// AddInstances
func BenchmarkAuthFuncListAddRemoveInstances(b *testing.B) {
	b.Logf("This test adds and removes the same instance because of practical constrains")
	b.Logf("Would probably benefit to add/remove more than one, though")
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init()
	for i := 0; i < b.N; i++ {
		list.AddInstances(instances[0])
		list.RemoveInstances(instances[0].name)
	}
}
