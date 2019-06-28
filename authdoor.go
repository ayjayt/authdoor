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

// authFuncContainer could be a authFuncListCore or something that inherited it. DESIGN NOTE: Different authFuncListCore wrappers would be used to implement different concurrency or data management architectures on top of the core type.
type authFuncContainer interface {
	AddCallables(callables ...authFuncCallable)
	RemoveCallables(name ...string)
}

// authFuncListCore is the basic idea of a list of iterable AuthFuncs.
type authFuncListCore struct {
	name     string
	funcList []authFuncInstance
	funcMap  map[string]int // cornelk/hashmap would be faster
	wg       sync.WaitGroup // for tracking readers
}

// authFuncList provides concurency support to authFuncList.
type authFuncList struct {
	authFuncListCore
	handlers []*authHandler // No way around this pointer
	mutex    *sync.RWMutex  // This pointer helps us avoiding copying a mutex
}

// newAuthFuncListCore will take all instances- so the values of authFuncList.funcList too- and merge everything into a new sorted authFuncList with it's own WaitGroup
func newAuthFuncListCore(name string, instances ...authFuncCallable) authFuncListCore {
	ret := authFuncListCore{
		name: name,
	}
	ret.AddCallables(instances)
	return ret
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

// btreeNode is a simple datastructure used to expedite sorting.
type btreeNode struct {
	item                *authFuncInstance
	parent, left, right *list
	size                int // only the root node would contain the size, will never be 0 if authFuncCallable != nil
}

// insert will place an authFuncInstance in the btree by order of priority
func (n *btreeNode) insert(*authFuncInstance) {
	if (n.size != 0) || parent == nil {
		size++
	}
	if n.item == nil {
		n.item = authFuncCallable
		return
	}
	if n.item.priority > authFuncInstance.priority {
		if n.left == nil {
			n.left = &btreeNode{parent: n}
		}
		n.left.insert(authFuncInstance)
		return
	}
	if n.right == nil {
		n.right = &btreeNode{parent: n}
	}
	n.right.insert(authFuncInstance)
}

func (n *btreeNode) min() *breeNode {
	if n.left == nil {
		return n
	}
	return n.left.min()
}

func (n *btreeNode) nextHighest() *btreeNode {
	// if left
	// if right
	// gosh which way do we go
	if n.right != nil {
		return n.right.min()
	}
	n.parent.nextHighest
}

func (n *btreeNode) writeSlice(slice []authFuncInstance) {
	return nil
}

// AddCallables will add any AuthFuncList/Instance to it's own authFuncListCore, sorted properly.
func (l *authFuncListCore) AddCallables(callables ...authFuncCallable) {
	// build tree out of all and build slice
}

// RemoveCallables can remove a AuthFuncList/Instance from it's core
func (l *authFuncListCore) RemoveCallables(names ...string) {
	// TODO, remove callables from list
	// first, make them nil by name
	// then loop through moving it over to the left
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
