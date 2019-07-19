package authdoor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestAuthFuncListSafeInit tests the Init() method
func TestAuthFuncListTemplateInit(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListTemplate)
	err := list.Init("test", instances...)
	require.Nil(t, err)
	require.Equal(t, len(sortableInstances), len(list.funcList))
	require.Equal(t, len(sortableInstances), len(list.funcMap))
	require.NotNil(t, list.handlerMutex)
	require.Equal(t, list.name, "test")
	ordered, errorList := checkOrder(&list.AuthFuncListSafe.AuthFuncList)
	require.True(t, ordered, "list value returned: %v", errorList)
	for i := range list.funcList {
		// This wouldn't work if priorities weren't distinct
		require.Equal(t, list.AuthFuncListSafe.AuthFuncList.funcList[i].name, sortableInstances[i].name)
	}
}

// TestAuthFuncListTemplateAddHandlers will test the AddHandler() method
func TestAuthFuncListTemplateAddHandlers(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListTemplate)
	list.Init("test", instances...)
	// Create Handler
	handler := new(AuthHandler)
	handler.Init(nil)
	list.AddHandler(handler)
	require.Equal(t, handler, list.handlers[0])
	require.Equal(t, 1, len(list.handlers))
}

// TestAuthFuncListTemplateRemoveHandler will test the RemoveHandler() method
func TestAuthFuncListTemplateRemoveHandler(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListTemplate)
	list.Init("test", instances...)
	handler := new(AuthHandler)
	handler.Init(nil)
	handler2 := new(AuthHandler)
	handler2.Init(nil)
	handler3 := new(AuthHandler)
	handler3.Init(nil)

	list.AddHandler(handler)
	list.AddHandler(handler2)
	list.AddHandler(handler3)

	require.Equal(t, handler, list.handlers[0])
	require.Equal(t, handler2, list.handlers[1])
	require.Equal(t, handler3, list.handlers[2])
	require.Equal(t, 3, len(list.handlers))
	list.RemoveHandler(handler2)
	require.Equal(t, handler, list.handlers[0])
	require.Equal(t, handler3, list.handlers[1])
	require.Equal(t, 2, len(list.handlers))
	list.RemoveHandler(handler)
	require.Equal(t, handler3, list.handlers[0])
	require.Equal(t, 1, len(list.handlers))
	list.RemoveHandler(handler3)
	require.Equal(t, 0, len(list.handlers))
}

// It would be better I guess if AuthFuncListTemplate handlers took an interface, that if the interface defined most of what was needed to access the data strucutre. Then I could test with mocks. It tests still need to sometimes access things that wont be exposed by the interface. Either way, I will rewrite this once UpdateHandler is done.

// TestAuthFuncListUpdateHandler will test the UpdateHandler() method
func TestAuthFuncListTemplateUpdateHandler(t *testing.T) {
	instances, _ := makeInstances(t, sortableInstances)
	list := new(AuthFuncListTemplate)
	list.Init("test", instances...)
	handler := new(AuthHandler)
	handler.Init(nil)
	handler2 := new(AuthHandler)
	handler2.Init(nil)
	list.AddHandler(handler)
	list.AddHandler(handler2)
	ch, total := list.UpdateHandlers()
	require.Equal(t, 2, total)
	require.IsType(t, ch, make(chan int, 2))
	list.BlockForUpdate(ch, total)
}

// Test UpdateHandler
// Test BlockForUpdate

// BenchmarkAuthFuncListSafeInit will benchmark the Init() method
func BenchmarkAuthFuncListTemplateInit(b *testing.B) {
	instances, _ := makeInstances(b, sortableInstances)
	b.ResetTimer()
	b.Run("Init", func(b *testing.B) {
		for i := 0; i <= b.N; i++ {
			list := new(AuthFuncListTemplate)
			list.Init("benchmark", instances...)
		}
	})
}

// BenchmarkAuthFuncListSafeAddRemoveInstance will benchmark the AddRemove() method
func BenchmarkAuthFuncListTemplateAddRemoveHandler(b *testing.B) {
	if testing.Verbose() {
		b.Logf("This test adds and removes the same instance because of practical constrains")
		b.Logf("Would probably benefit to add/remove more than one, though")
	}
	instances, _ := makeInstances(b, sortableInstances)
	list := new(AuthFuncListTemplate)
	list.Init("benchmark", instances...)
	handler := new(AuthHandler)
	handler.Init(nil)
	for i := 0; i < b.N; i++ {
		list.AddHandler(handler)
		list.RemoveHandler(handler)
	}
}
