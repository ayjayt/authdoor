package authdoor

import (
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// We can't use http's handlerfuncs as handlers to test because require can't compare funcs
type emptyHandler struct{}

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
	renamedSortables := sortableInstances
	renamedSortables2 := sortableInstances
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

// I'm not happy with how unInited structs are handled (ErrNameTaken conflict with ""
// I'm not witht he lack of transactionality
// I feel like this could be more concise and readable
// There needs to be better logging

// Test UpdateHandler
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

// Test ServeHTTP
func TestAuthHandlerServeHTTP(t *testing.T) {
	handler := new(AuthHandler)
	handler.Init(nil)
	handler.ServeHTTP(nil, nil) // cool
}

// This is an integration test
// we should build it up and provide a mock
// creating an instance
// giving that instance a mock
// calling it
// integrationtest

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

// Update
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

func BenchmarkAuthHandlerServeHTTP(b *testing.B) {
	handler := new(AuthHandler)
	handler.Init(nil)
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(nil, nil) // cool
	}
}
