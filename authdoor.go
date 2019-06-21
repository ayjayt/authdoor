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

// AuthFunc is any function that takes a response writer and request and returns a bool, true if auth succeeds
// If false, don't write- we've already written
type AuthStatus int

const (
	AuthGranted AuthStatus = iota
	Pending
	AuthDenied
)

type AuthFunc func(w http.ResponseWriter, r *http.Request) AuthStatus

// NewAuthFuncInstance takes some AuthFunc and lets you build an instance out of it
func NewAuthFuncInstance(name string, authFunc AuthFunc, priority int) AuthFuncInstance {
	return AuthFuncInstance{
		name:     name,
		authFunc: AuthFunc,
		priority: priority,
		active:   true,
	}
}

// AuthFuncInstance is the structure actually used by a handler, it includes some help for race conditions and ident
type AuthFuncInstance struct {
	name     string
	authFunc AuthFunc
	priority int
	wg       sync.WaitGroup
	active   bool
}

// Call does the work of calling the auth function
func (i *AuthFuncInstance) Call() AuthStatus {
	if !active {
		return AuthDenied
	}
	i.wg.Add()
	if !active {
		i.wg.Done()
		return AuthDenied
	}
	res := authFunc()
	i.wg.Done()
	return res
}

// Deactivate is used to turn off an AuthFuncInstance and will block until all flights are cleared
func (i *AuthFuncInstance) Deactivate() error { // TODO: add timeout
	i.active = false
	i.wg.Wait()
	// TODO: goroutine to wait on waitgroup w/ TimeOut
	return nil
}

// AuthHandler is an http.Handler wrapper that manages its authorization options
type AuthHandler struct {
	Base  http.Handler
	auths [2]struct {
		size     int
		funcList [MaxAuths]AuthFuncInstance
		funcMap  map[string]int
		wg       sync.WaitGroup // TODO: I guess be able to set instant or slow timeout- add hooks to authentication
		active   bool
	}
	currentList      int
	currentListMutex sync.RWMutex
}

// AttachHandler sets the base http.Handler
func (h *AuthHandler) AttachHandler(handler http.Handler) {
	h.Base = handler
}

// AddAuth can be used to add authentication to a handler
func (h *AuthHandler) AddAuths([]AuthFuncInstance) error {
	_, ok := h.Auth.LoadOrStore(name, authfunc)
	if !ok {
		return ErrNameTaken
	}
	// order the auths by priority with lowest first
	// wait for mutex, and then lock, so lock
	// then wait for the buffer (any clients still scrolling through the list?)
	// I think these need context's with timeouts
	// copy over the func list and add the ones you want
	// switch over currentList and unlock the mutex
	return nil
}

// DeleteAuth removed authentication from a handler
func (h *AuthHandler) DeleteAuth(name string) {
	_, ok := h.Auth.Delete(name)
}

// okay, you may want to somehow edit the list
// you may want to add a function of a certain priority
// or remove a function of a certain priority
// but in this case, you'll need to lock currentList

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: range through off, generally
	H.Base.ServeHTTP()
}
