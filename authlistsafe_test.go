package authdoor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthFuncListSafeInit tests the Init() method
func TestAuthFuncListSafeInit(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	require.NotNil(t, list.listMutex)
	ordered, errorList := checkOrder(&list.AuthFuncList)
	require.True(t, ordered, "list value returned: %v", errorList)
	for i := range list.funcList {
		// This wouldn't work if priorities weren't distinct
		require.Equal(t, list.AuthFuncList.funcList[i].name, sortableInstances[i].name)
	}
}

// TestAuthFuncListSafeCall will test the Call() method
func TestAuthFuncListSafeCall(t *testing.T) {
	instances, mocks := makeInstances(t, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	ordered, errorList := checkOrder(&list.AuthFuncList)
	require.True(t, ordered, "list value returned: %v", errorList)
	for i, v := range sortableInstances {
		ret, err := list.Call(nil, nil, v.name)
		require.Equal(t, v.expErr, err)
		require.Equal(t, v.expRet.Auth, ret.Auth)
		require.Equal(t, v.expRet.Resp, ret.Resp)
		mocks[i].RequireCalled(t)
	}
}

// TestAuthFuncListSafeCallAll will test the CallAll() method
func TestAuthFuncListSafeCallAll(t *testing.T) {
	instances, mocks := makeInstances(t, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	ordered, errorList := checkOrder(&list.AuthFuncList)
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

// TestAuthFuncListSafeAddInstance will test the AddInstance() method
func TestAuthFuncListSafeAddInstance(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init()
	require.Equal(t, 0, len(list.funcList))
	require.Equal(t, 0, len(list.funcMap))
	list.AddInstances(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	ordered, errorList := checkOrder(&list.AuthFuncList)
	require.True(t, ordered, "list value returned: %v", errorList)
}

// TestAuthFuncListSafeRemoveInstance will test the RemoveInstance method
func TestAuthFuncListSafeRemoveInstance(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	ordered, errorList := checkOrder(&list.AuthFuncList)
	require.True(t, ordered, "list value returned: %v", errorList)
	for _, v := range instances {
		list.RemoveInstances(v.name)
	}
	require.Equal(t, 0, len(list.funcList))
	require.Equal(t, 0, len(list.funcMap))
}

// BenchmarkAuthFuncListSafeInit will benchmark the Init() method
func BenchmarkAuthFuncListSafeInit(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	b.ResetTimer()
	b.Run("Init", func(b *testing.B) {
		for i := 0; i <= b.N; i++ {
			list := new(AuthFuncListSafe)
			list.Init(instances...)
		}
	})
}

// BenchmarkAuthFuncListSafeCall will benchmark the Call() method
func BenchmarkAuthFuncListSafeCall(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.Call(nil, nil, instances[0].name)
	}
}

// BenchmarkAuthFuncListSafeCallAll will benchmark the CallAll() method
func BenchmarkAuthFuncListSafeCallAll(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init(instances...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = list.CallAll(nil, nil)
	}
}

// BenchmarkAuthFuncListSafeAddRemoveInstance will benchmark the AddRemove() method
func BenchmarkAuthFuncListSafeAddRemoveInstance(b *testing.B) {
	if testing.Verbose() {
		b.Logf("This test adds and removes the same instance because of practical constrains")
		b.Logf("Would probably benefit to add/remove more than one, though")
	}
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncListSafe)
	list.Init()
	for i := 0; i < b.N; i++ {
		list.AddInstances(instances[0])
		list.RemoveInstances(instances[0].name)
	}
}
