package authdoor

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// emptyHandler is a shim since we must have a non-func to test for quality with the prequire package
type emptyHandler struct{}

// ServeHTTP is the empty function completing the http.Handler interface for emptyHandler
func (h *emptyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

// TestAuthHandlerInit tests that the Init() method is working and everything is set properly
func TestAuthHandlerInit(t *testing.T) {
	handler := new(AuthHandler)
	emptyHandler := new(emptyHandler)
	err := handler.Init(emptyHandler)
	require.Nil(t, err)
	require.IsType(t, emptyHandler, handler.base)
	require.NotNil(t, handler.componentMutex)
	require.NotNil(t, handler.componentsList)
	item, ok := handler.componentsList[""]
	require.True(t, ok, "blank components list not found in map")
	require.IsType(t, new(AuthFuncListTemplate), item)
	require.NotNil(t, handler.activeMutex)
}

// TestAuthHandlerGetBase makes sure that the GetBase() returns the base set manually
func TestAuthHandlerGetBase(t *testing.T) {
	handler := new(AuthHandler)
	emptyHandler := new(emptyHandler)
	err := handler.Init(emptyHandler)
	require.Nil(t, err)
	require.IsType(t, emptyHandler, handler.GetBase())
}

// TestAuthHandlerSetBase makes sure that the SetBase() returns the base set manually
func TestAuthHandlerSetBase(t *testing.T) {
	handler := new(AuthHandler)
	emptyHandler := new(emptyHandler)
	err := handler.Init(nil)
	require.Nil(t, err)
	handler.SetBase(emptyHandler)
	require.IsType(t, emptyHandler, emptyHandler)
}

// TestAuthHandlerAddInstances will make sure you can add instances to a handler
func TestAuthHandlerAddInstances(t *testing.T) { // TODO: test for the error
	handler := new(AuthHandler)
	instances, _ := makeInstances(t, sortableInstances)
	err := handler.Init(nil)
	require.Nil(t, err)
	handler.AddInstances(instances...)
	require.Equal(t, len(instances), len(handler.componentsList[""].funcList))
}

// TestAuthHandlerRemoveInstances will make sure you can remove instances
func TestAuthHandlerRemoveInstances(t *testing.T) {
	handler := new(AuthHandler)
	instances, _ := makeInstances(t, sortableInstances)
	err := handler.Init(nil)
	require.Nil(t, err)
	handler.AddInstances(instances...)
	handler.RemoveInstances(sortableInstances[0].name, sortableInstances[2].name)
	_, ok := handler.componentsList[""].funcMap[sortableInstances[0].name]
	require.False(t, ok)
	_, ok = handler.componentsList[""].funcMap[sortableInstances[2].name]
	require.False(t, ok)
	require.Equal(t, len(sortableInstances)-2, len(handler.componentsList[""].funcList))
	handler.RemoveInstances(sortableInstances[5].name)
	require.Equal(t, len(sortableInstances)-3, len(handler.componentsList[""].funcList))
	_, ok = handler.componentsList[""].funcMap[sortableInstances[5].name]
	require.False(t, ok)
}

// TestAuthHandlerAddLists will test the AddLists() method of AuthHandler
func TestAuthHandlerAddLists(t *testing.T) {
	handler := new(AuthHandler)
	instances, _ := makeInstances(t, sortableInstances)
	err := handler.Init(nil)
	require.Nil(t, err)
	template := new(AuthFuncListTemplate)
	template.Init("testTemplate", instances...)
	err = handler.AddLists(template)
	require.Nil(t, err)
	require.Equal(t, 2, len(handler.componentsList))
	err = handler.AddLists(template)
	require.Equal(t, ErrNameTaken, errors.Cause(err))
	require.Equal(t, 2, len(handler.componentsList))
}

// TestAuthHandlerRemoveLists will test the RemoveLists() method of AuthHandler
func TestAuthHandlerRemoveLists(t *testing.T) {
	handler := new(AuthHandler)
	instances, _ := makeInstances(t, sortableInstances)
	renamedSortables := make([]testInstancesRaw, len(sortableInstances))
	renamedSortables2 := make([]testInstancesRaw, len(sortableInstances))
	copy(renamedSortables, sortableInstances)
	copy(renamedSortables2, sortableInstances)
	for i := range sortableInstances {
		renamedSortables[i].name = renamedSortables[i].name + "2"
		renamedSortables2[i].name = renamedSortables2[i].name + "3"
	}
	instances2, _ := makeInstances(t, renamedSortables)
	instances3, _ := makeInstances(t, renamedSortables2)
	err := handler.Init(nil)
	require.Nil(t, err)

	template := new(AuthFuncListTemplate)
	err = template.Init("testTemplate", instances...)
	require.NoError(t, err)

	template2 := new(AuthFuncListTemplate)
	err = template2.Init("testTemplate2", instances...)
	require.NoError(t, err)

	template3 := new(AuthFuncListTemplate)
	err = template3.Init("testTemplate3", instances...)
	require.NoError(t, err)

	err = handler.AddLists(template)

	err = handler.AddLists(template2, template3)

	template4 := new(AuthFuncListTemplate)
	err = template4.Init("testTemplate2", instances2...)
	require.NoError(t, err)

	template5 := new(AuthFuncListTemplate)
	err = template5.Init("testTemplate5", instances3...)
	require.NoError(t, err)

	// TODO: what if we add an united template?
	err = handler.AddLists(template4, template5)
	require.Equal(t, errors.Wrap(ErrNameTaken, "testTemplate2").Error(), err.Error())
	err = handler.AddLists(template5)
	require.NoError(t, err)
	require.Equal(t, 5, len(handler.componentsList))
	handler.RemoveLists("testTemplate2")
	require.Equal(t, 4, len(handler.componentsList))
	//require.Equal(t, 45, len(handler.activeLists[handler.currentList].funcList))

}

// I'm not happy with how unInited structs are handled (ErrNameTaken conflict with "")
// I'm not witht he lack of transactionality
// I feel like this could be more concise and readable
// There needs to be better logging

// TestAuthHandlerUpdateHandler makes sure that UpdateHandler works and it's notifier works. This test is pretty weak.
func TestAuthHandlerUpdateHandler(t *testing.T) {
	notifier := make(chan int, 1) // if not a buffer of one, handler.UpdateHandler must be a goroutine
	handler := new(AuthHandler)
	handler.Init(nil)
	require.Nil(t, handler.activeLists[handler.currentList])
	require.Equal(t, 0, handler.currentList)
	handler.UpdateHandler(notifier)
	require.Equal(t, 1, <-notifier)
	require.NotNil(t, handler.activeLists[handler.currentList])
	require.Equal(t, 1, handler.currentList)
}

// TestAuthHandlerServeHTTP makes sure that the ServeHTTP function works. It has no requires but would throw an error if it didn't work. This test is pretty weak.
func TestAuthHandlerServeHTTP(t *testing.T) {
	handler := new(AuthHandler)
	handler.Init(nil)
	handler.ServeHTTP(nil, nil) // cool
}

// TestIntegration tests as many functions of the library as possible.
func TestIntegration(t *testing.T) {
	var err error
	renamedSortables := make([]testInstancesRaw, len(sortableInstances))
	copy(renamedSortables, sortableInstances)
	for i := range sortableInstances {
		renamedSortables[i].expRet.Auth = AuthFailed
		renamedSortables[i].expRet.Resp = Ignored
	}
	renamedSortables2 := make([]testInstancesRaw, len(sortableInstances))
	renamedSortables3 := make([]testInstancesRaw, len(sortableInstances))
	copy(renamedSortables2, renamedSortables)
	copy(renamedSortables3, renamedSortables)
	for i := range sortableInstances {
		renamedSortables2[i].name = renamedSortables2[i].name + "2"
		renamedSortables3[i].name = renamedSortables3[i].name + "3"
	}
	instances, mocks := makeInstances(t, renamedSortables)
	instances2, mocks2 := makeInstances(t, renamedSortables2)
	instances3, mocks3 := makeInstances(t, renamedSortables3)

	template := new(AuthFuncListTemplate)
	err = template.Init("template", instances2...)
	require.Nil(t, err)
	template2 := new(AuthFuncListTemplate)
	err = template2.Init("template2", instances3...)
	require.Nil(t, err)

	handler := new(AuthHandler)
	handler2 := new(AuthHandler)

	err = handler.Init(nil)
	require.Nil(t, err)
	err = handler2.Init(nil)
	require.Nil(t, err)
	// TODO: Init the handler to check if it can be called
	// TODO: Change one of the items on the list so that it _is_ called
	// TODO: take the every permutation list and loop through this test whith each item. We will know whether or not the func should be called.

	handler.AddLists(template)
	handler2.AddLists(template2)
	handler.AddInstances(instances...)
	t.Logf("Number of instances: %v size of template: %v", len(instances), len(template.funcList))
	handler.UpdateHandler(nil)
	handler2.UpdateHandler(nil)
	currentItems := handler.activeLists[handler.currentList].ListInstances()
	t.Logf("%+v\n", currentItems)
	handler.ServeHTTP(nil, nil)
	handler2.ServeHTTP(nil, nil)
	// TODO Can we change lists?
	// TODO Can we remove lists?
	for i := range mocks {
		mocks[i].RequireCalled(t)

		mocks2[i].RequireCalled(t)
		mocks3[i].RequireCalled(t)
	}
	_ = mocks
	_ = mocks2
	_ = mocks3
}

// BenchmarkAuthHandlerInit will test the Init function to determine how long it takes
func BenchmarkAuthHandlerInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handler := new(AuthHandler)
		emptyHandler := new(emptyHandler)
		handler.Init(emptyHandler)
	}
}

// BenchmarkAuthHandlerAddRemoveInstances will benchmark adding and removing instances to a handler
func BenchmarkAuthHandlerAddRemoveInstances(b *testing.B) {
	handler := new(AuthHandler)
	instances, _ := makeInstances(b, sortableInstances)
	handler.Init(nil)
	for i := 0; i < b.N; i++ {
		handler.AddInstances(instances[0])
		handler.RemoveInstances(sortableInstances[0].name)
	}
}

// BenchmarkAuthHandlerAddRemoveLists will benchmaring adding and removing of lists to a handler
func BenchmarkAuthHandlerAddRemoveLists(b *testing.B) {
	handler := new(AuthHandler)
	instances, _ := makeInstances(b, sortableInstances)
	handler.Init(nil)
	template := new(AuthFuncListTemplate)
	template.Init("testTemplate", instances...)
	for i := 0; i < b.N; i++ {
		handler.AddLists(template)
		handler.RemoveLists("testTemplate")
	}

}

// BenchmarkAuthHandlerUpdateHandler will benchmark update handler in two methods- with buffered channels or with goroutines.
func BenchmarkAuthHandlerUpdateHandler(b *testing.B) {
	handler := new(AuthHandler)
	handler.Init(nil)
	b.Run("buffered", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			notifier := make(chan int, 1) // if not a buffer of one, handler.UpdateHandler must be a goroutine
			handler.UpdateHandler(notifier)
			<-notifier
		}
	})
	b.Run("goroutine", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			notifier := make(chan int) // if not a buffer of one, handler.UpdateHandler must be a goroutine
			go handler.UpdateHandler(notifier)
			<-notifier
		}
	})
}

// BenchmarkAuthHandlerServeHTTP just benchmarks the ServeHTTP function with no handlers to call.
func BenchmarkAuthHandlerServeHTTP(b *testing.B) {
	handler := new(AuthHandler)
	handler.Init(nil)
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(nil, nil) // cool
	}
}
