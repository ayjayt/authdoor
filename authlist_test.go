package authdoor

import (
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"
)

// seededRand will reseed our random number generator
var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

// sortableInstances must have a negative and ascending to -1, since this slice also acts as a reference for test assertions.
var sortableInstances = []testInstancesRaw{
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

// makeInstances takes our test table and turns it into actual instances w/ a mock function for each AuthFunc that can check to see if it's been called.
func makeInstances(t testing.TB, raw []testInstancesRaw) ([]AuthFuncInstance, []*MockImpl) {
	mockImpls := make([]*MockImpl, len(sortableInstances))
	authInstances := make([]AuthFuncInstance, len(sortableInstances))
	for i := range raw {
		dut := AuthFuncInstance{}
		mockImpl := newMockAuthFunc(raw[i].expRet, raw[i].expErr)
		dut.Init(raw[i].name, mockImpl.mockAuthFunc, raw[i].priority, nil)
		mockImpls[i] = mockImpl
		authInstances[i] = dut
	}
	return authInstances, mockImpls
}

// checkOrder determines if the instance list is in the correct order by comparing adjacent item's priorities and making sure the map corresponds to the correct index.
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

// TestAuthFuncListInit tests to make sure our Init() works.
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
		require.Equal(t, sortableInstances[i].name, list.funcList[i].name)
	}
}

// TestAuthFuncListListInstances tests that we can retreive a list of instance names
func TestAuthFuncListListInstances(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	names := list.ListInstances()
	for i, v := range names {
		// This wouldn't work if priorities weren't distinct
		require.Equal(t, sortableInstances[i].name, v)
	}
}

// TestAuthFuncListSort tests to see if our sort() method works.
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

// TestAuthFuncListCall will check to see if the Call()
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
		mocks[i].RequireCalled(t)
	}
}

// TestAuthFuncListCallAll will check to see if the CallAll() method works
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
		mocks[i].RequireCalled(t)
	}
}

// TestAuthFuncListAddInstance tests the AddInstance() method
func TestAuthFuncListAddInstance(t *testing.T) {
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

// TestAuthFuncListRemoveInstance tests the RemoveInstance() method
func TestAuthFuncListRemoveInstance(t *testing.T) {
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

// BenchmarkAuthFuncListInit benchmarks the Init() method
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

// BenchmarkAuthFuncListSort benchmarks the Sort() method
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

// BenchmarkAuthFuncListCall benchmarks the Call() method
func BenchmarkAuthFuncListCall(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.Call(nil, nil, instances[0].name)
	}
}

// BenchmarkAuthFuncListCallAll benchmarks the CallAll() method
func BenchmarkAuthFuncListCallAll(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.CallAll(nil, nil)
	}
}

// BenchmarkAuthFuncListInitAddRemoveInstances benchmarks the Add and Remove methods
func BenchmarkAuthFuncListAddRemoveInstances(b *testing.B) {
	if testing.Verbose() {
		b.Logf("This test adds and removes the same instance because of practical constrains")
		b.Logf("Would probably benefit to add/remove more than one, though")
	}
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncList)
	list.Init()
	for i := 0; i < b.N; i++ {
		list.AddInstances(instances[0])
		list.RemoveInstances(instances[0].name)
	}
}
