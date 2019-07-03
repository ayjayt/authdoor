package authdoor

import (
	"error"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	// ErrNameTaken is returned when someone tries to register an auth method on a handler that already exists
	ErrNameTaken = error.New("tried to create an auth function with the same name as an existing function")
)

// AuthStatus contains information from an AuthFunc about authorization status.
type AuthStatus uint8

// Responded is true if we wrote to the ResponseWriter- it is returned by an Authfunc.
type Responded bool

const (
	// AuthFailed is returned by an AuthFunc if it couldn't determine the users identity.
	AuthFailed AuthStatus = iota
	// AuthGranted is returned by an AuthFunc it was determined the user is authorized
	AuthGranted
	// AuthDenied is returned by an AuthFunc essentially if we know the user is unauthorized.
	AuthDenied
)

// AuthFunc is any function that takes a response writer and request and returns two state variables, AuthStatus and Responded. TODO: Probably need to return some user data.
type AuthFunc func(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded)

// This interface represents anything that has a call function- so a list or an instance.
type authFuncCallable interface {
	call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded)
}

// authFuncInstance is the structure actually used by a handler, it includes some meta data around the function. DESIGN NOTE: We don't use receiver because it makes it harder to create more concurrency issues by having multiple goroutines access the same structure.
type authFuncInstance struct {
	name     string
	authFunc AuthFunc
	priority int
}

// NewAuthFuncInstance takes some AuthFunc and lets you build an instance out of it
func NewAuthFuncInstance(name string, authFunc AuthFunc, priority int) authFuncInstance {
	return authFuncInstance{
		name:     name,
		authFunc: AuthFunc,
		priority: priority,
	}
}

// call does the work of calling the auth function. DESIGN NOTE: Originally implemented to call the function along with it's containing structures race-preventing types, I'm not sure it's the best choice now.
func (i *authFuncInstance) call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded) {
	return i.authFunc(w, r)
}

// authFuncListCore is the basic idea of a list of iterable AuthFuncs.
type authFuncListCore struct {
	name     string
	funcList []authFuncInstance // these are copied, and this needs to be reordered
	funcMap  map[string]int     // cornelk/hashmap would be faster
	wg       sync.WaitGroup     // for tracking readers
}

func (c *authFuncListCore) Len() int {
	return len(c.funcList)
}
func (c *authFuncListCore) Swap(i, j int) {
	c.funcList[i], c.funcList[j] = c.funcList[j], c.funcList[i]
}
func (c *authFuncListCore) Less(i, j int) bool {
	c.funcList[i].priority < c.funcList[j].priority
}
func (c *authFuncListCore) WriteMap() {
	for i, _ := range c.funcList {
		c.funcMap[funcList[i].name] = i
	}
}

// authFuncList provides concurency support to authFuncList.
type authFuncList struct {
	authFuncListCore
	handlers []*authHandler // No way around this pointer
	mutex    *sync.RWMutex  // This pointer helps us avoiding copying a mutex
}

// newAuthFuncListCore will take all instances- so the values of authFuncList.funcList too- and merge everything into a new sorted authFuncList with it's own WaitGroup
func newAuthFuncListCore(name string, instances ...authFuncCallable) (authFuncListCore, error) {
	ret := authFuncListCore{
		name: name,
	}
	err := ret.AddCallables(instances)
	return ret, err
}

// NewAuthFuncList creates a new list that can be used as a component of a handler's list.
func NewAuthFuncList(name string, instances ...authFuncCallable) *authFuncList {
	ret := authFuncList{
		authFuncListCore: newAuthFuncListCore(name, instances),
		handlers:         new([]*authHandler),
		mutex:            new(sync.RWMutex),
	}
}

// call will iterate through the authFuncListCore and return when AuthStatus is Denied or Responded is true, or when it completes without finding anything.
func (l *authFuncListCore) call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded) {
	// If we had a hint about which to call we could
	for i, _ := range l.funcList {
		status, responded := l.funcList[i](w, r)
		if (status == AuthDenied) || (Responded) {
			return status, responded
		}
	}
}

// AddCallables will add any AuthFuncList/Instance to it's own authFuncListCore, sorted properly.
func (l *authFuncListCore) AddCallables(callables ...authFuncCallable) error {
	// build tree out of all and build slice
	for i, _ := range callables {
		switch callables[i].(type) { // There needs to be a better way...
		case authFuncListCore:
			fallthrough
		case authFuncList:
			for j, _ = range callables[i].funcList {
				if _, ok := l.funcMap[callables[i].funcList[j].name]; ok {
					return errors.Wrap(ErrNameTaken, callables[i].funcList[j].name)
				}
			}
			authFuncListCore.funcList = append(authFuncListCore.funcList, callables[i].funcList...)
		case authFuncInstance:
			authFuncListCore.funcList = append(authFuncListCore.funcList, callables[i])
		}
	}
	sort.Sort(l)
	l.WriteMap()
	return nil
}

// RemoveCallables can remove a AuthFuncList/Instance from it's core
func (l *authFuncListCore) RemoveCallables(names ...string) {
	for i, _ := range names {
		l.funcList[l.funcMap[names[i]]].authFunc = nil
	}
	zombieCounter := 0
	newSize := 0
	for i, _ := range l.funcList {
		if l.funcList[i].authFunc == nil {
			zombieCounter++
		} else {
			newSize++
			if zombieCounter > 0 {
				l.funcList[i-zombieCounter] = l.funcList[i]
			}
		}
	}
	l.funcList = l.funcList[:newSize] // could be an error
}

// addHandler will have the handler points to the list
func (l *authFuncList) addHandler(handler *authHandler) {
	// TODO
}

// removeHandler will have the handler NOT point to the list
func (l *authFuncList) removeHandler(handler *authHandler) {
	// TODO
}

// UpdateHandlers will actually have the handler reorganize and rewrite the list that it is implementing
func (l *authFuncList) UpdateHandlers() {
	// TODO: go through all the registered handlers and reorganize
}

// authHandler is an http.Handler wrapper that manages its authorization options
type authHandler struct { // how many time will this be reused
	base http.Handler
	// This struct wraps the unique concurrency requirements of authHandlers. Concept is explained below the parent structures
	authFuncs struct {
		activeLists [2]authFuncListCore // the lists actually being used

		currentList    int                      // for directing readers
		mutex          sync.Mutex               // for writing
		componentsList map[string]*authFuncList // for default and external lists
	}
}

// CONCEPT: this lets us build up a new authFuncListCore based off a modified componentsList. Doing anything requires you to hold the Mutex, then you make the switch. Hold the Mutex and wait for the WaitGroup on the buffer you want to be 0.

// NewAuthHandler sets the base http.Handler
func NewAuthHandler(handler http.Handler, instances ...authFuncCallable) *authHandler {
	h := &authHandler{base: handler, authFuncs.mutex: new(sync.Mutex)}
	// TODO: propogate authFuncList if not null
} // Pointer here is important since we only want one, both for the Mutex and also to deal with the Lists pointing to it.

// GetBase returns the underlying http.Handler
func (h *authHandler) GetBase() http.Handler {
	return h.base
}

// GetBase sets the underlying http.Handler
func (h *authHandler) SetBase(handler http.Handler) {
	h.base = handler
}

func (h *authHandler) AddInstances(instances ...authFuncInstance) {
	// TODO, add an instance to your basic list
}

func (h *authHandler) RemoveInstances(instanceNames ...string) {
	// TODO, remove an instance from your basic list
}

func (h *authHandler) AddLists(lists ...authFuncList) {
	// TODO, add a list to your list of lists
}

func (h *authHandler) RemoveLists(listNames ...string) {
	// TODO, remove a list from your list of lists
}

func (h *authHandler) UpdateActiveList() {
	// TODO now reset lists
}
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: range through Auth, generally
	// TODO: CORS- see authHandler todo0 we should have like a preflight function we can assign
	h.base.ServeHTTP()
}
