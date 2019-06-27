package authdoor

import (
	"error"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	// ErrNameTaken is returned when someone tries to register an auth method on a handler that already exists
	ErrNameTaken = error.New("tried to create an auth function with the same name as an existing function")
)

// TODO: decide if you're going to have the registry here

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

// AuthFunc is any function that takes a response writer and request and returns two state variables, AuthStatus and Responded
type AuthFunc func(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded)

type authFuncCallable interface {
	call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded)
}

// authFuncInstance is the structure actually used by a handler, it includes some meta data around the function
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

// Call does the work of calling the auth function
func (i *authFuncInstance) call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded) {
	return i.authFunc(w, r)
}

// authFuncListCore is the basic idea of a list of iterable AuthFuncs.
type authFuncListCore struct {
	name     string
	funcList []authFuncInstance
	funcMap  map[string]int // cornelk/hashmap would be faster
}

// authFuncList provides concurency support to authFuncList.
type authFuncList struct {
	authFuncListCore
	handlers []*authHandler // No way around this pointer
	mutex    *sync.RWMutex  // This pointer helps us avoiding copying a mutex
}

func NewAuthFuncList() AuthFuncList {
	// TODO: create an extenralAuthFuncList, and  return it after initializing it's mutex
}

func (l *authFuncListCore) call(w http.ResponseWriter, r *http.Request) (AuthStatus, Responded) {
	// TODO: iterate through the list if active and do the return or w/e
}

func (l *authFuncListCore) AddCallables(callables ...authFuncCallable) {
	// TODO, add calables to list
}

func (l *authFuncListCore) RemoveCallables(names ...string) {
	// TODO, remove callables from list
}

func (l *authFuncList) Cycle() {
	// TODO: go through all the registered handlers and reorganize
	// TODO: remove it from the handler if it's no longer active
}

// TODO: does this lock inside or outside, part of transactions

// newAuthFuncListCore will take all instances- so the values of authFuncList.funcList too- and merge everything into a new sorted authFuncList with it's own WaitGroup
func newAuthFuncListCore(name string, instances ...authFuncCallable) authFuncListCore {
	// TODO: just create it and call add (make map)
}

// authHandler is an http.Handler wrapper that manages its authorization options
type authHandler struct { // how many time will this be reused
	base http.Handler
	// This struct wraps the unique concurrency requirements of authHandlers. Concept is explained below the parent structures
	authFuncs struct {
		activeLists    [2]authFuncListCore         // the lists actually being used
		currentListWg  [2]sync.WaitGroup           // for tracking readers
		currentList    int                         // for directing readers
		mutex          sync.Mutex                  // for writing
		componentsList map[string]authFuncListCore // for default and external lists
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
