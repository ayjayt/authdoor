package authdoor

import (
	"net/http"
	"sort"
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrNameTaken is returned when someone tries to register an auth method on a handler that already exists
	ErrNameTaken = errors.New("tried to create an auth function with the same name as an existing function")
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
		authFunc: authFunc,
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
}

// newAuthFuncListCore will take all instances- so the values of authFuncList.funcList too- and merge everything into a new sorted authFuncList with it's own WaitGroup
func newAuthFuncListCore(name string, callables ...authFuncCallable) (authFuncListCore, error) {
	ret := authFuncListCore{
		name: name,
	}
	err := ret.AddCallables(callables...) // no lock because it's the first call and nothing has access to it yet
	return ret, err                       // Should this be a pointer? We don't want ot ocpy cores, copy sync.WaitGroups
}

// Len returns the lenfth of the object to be sorted (used by sort.Sort)
func (c *authFuncListCore) Len() int {
	return len(c.funcList)
}

// Swap swaps the two objects by index (used by sort.Sort)
func (c *authFuncListCore) Swap(i, j int) {
	c.funcList[i], c.funcList[j] = c.funcList[j], c.funcList[i]
}

// Less is a comparison operator used by sort.Sort
func (c *authFuncListCore) Less(i, j int) bool {
	return c.funcList[i].priority < c.funcList[j].priority
}

// WriteMap constructs the funcMap (to be used after sorting) which is used to quickly map a name of a AuthFunc to its index.
func (c *authFuncListCore) WriteMap() {
	for i, _ := range c.funcList {
		c.funcMap[c.funcList[i].name] = i
	}
}

// call will iterate through the authFuncListCore and return when AuthStatus is Denied or Responded is true, or when it completes without finding anything.

func (l *authFuncListCore) call(w http.ResponseWriter, r *http.Request) (status AuthStatus, responded Responded) {
	// If we had a hint about which to call we could
	for i, _ := range l.funcList {
		status, responded = l.funcList[i].call(w, r)
		if (status == AuthDenied) || (responded) {
			return status, responded
		}
	}
	return status, responded
}

// AddCallables will add any AuthFuncList/Instance to it's own authFuncListCore, sorted properly.
func (l *authFuncListCore) AddCallables(callables ...authFuncCallable) error {
	for i, _ := range callables {
		switch callables[i].(type) { // There needs to be a better way...
		case *authFuncList, *authFuncListCore:
			for j, _ := range callables[i].(*authFuncListCore).funcList {
				if _, ok := l.funcMap[callables[i].(*authFuncListCore).funcList[j].name]; ok {
					return errors.Wrap(ErrNameTaken, callables[i].(*authFuncListCore).funcList[j].name)
				}
			}
			l.funcList = append(l.funcList, callables[i].(*authFuncListCore).funcList...)
		case *authFuncInstance:
			l.funcList = append(l.funcList, *callables[i].(*authFuncInstance))
		}
	}
	sort.Sort(l)
	l.WriteMap()
	return nil
}

// RemoveCallables can remove a AuthFuncList/Instance from it's core. It maintains order.
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
	l.WriteMap()
}

// authFuncList provides concurency support to authFuncList.
type authFuncList struct {
	authFuncListCore
	handlers []*authHandler // No way around this pointer
	mutex    *sync.RWMutex  // This pointer helps us avoiding copying a mutex
}

// NewAuthFuncList creates a new list that can be used as a component of a handler's list.
func NewAuthFuncList(name string, callables ...authFuncCallable) (*authFuncList, error) {
	core, err := newAuthFuncListCore(name, callables...)
	if err != nil {
		return nil, err
	}
	ret := &authFuncList{
		authFuncListCore: core,
		mutex:            new(sync.RWMutex),
	}
	return ret, nil
}

// addHandler will have the handler points to the list, intended to be called by authHandler.AddLists().
func (l *authFuncList) addHandler(handler *authHandler) {
	l.handlers = append(l.handlers, handler)
}

// removeHandler will remove the handlers from the list, intended to be called by authHandler.RemoveLists()
func (l *authFuncList) removeHandler(handler *authHandler) {
	for i, _ := range l.handlers {
		if l.handlers[i] == handler {
			l.handlers[len(l.handlers)-1], l.handlers[i] = nil, l.handlers[len(l.handlers)-1]
			l.handlers = l.handlers[:len(l.handlers)-1]
			return
		}
	}
}

// UpdateHandlers will actually have the handler reorganize and rewrite the list that it is implementing
func (l *authFuncList) UpdateHandlers() (chan int, int) {
	totalHandlers := len(l.handlers)
	completionNotifier := make(chan int, totalHandlers)
	for i, _ := range l.handlers {
		go func() {
			lockNeeded := l.handlers[i].startLock()
			if lockNeeded {
				l.handlers[i].UpdateActiveList()
				l.handlers[i].endLock()
			}
			completionNotifier <- 1
		}()
	}
	return completionNotifier, totalHandlers
}

// BlockForUpdate will wait for all the handlers to update- may be unnecessary
func (l *authFuncList) BlockForUpdate(completionNotifier chan int, totalHandlers int) {
	var handlersComplete int
	for i := range completionNotifier {
		handlersComplete += i
		if handlersComplete == totalHandlers {
			return
		}
	}
}

type authFuncLock struct {
	activeLists    [2]authFuncListCore // the lists actually being used
	wg             [2]sync.WaitGroup
	toUpdate       bool
	currentList    int                      // for directing readers
	mutex          *sync.Mutex              // for writing
	componentsList map[string]*authFuncList // for default and external lists
}

// authHandler is an http.Handler wrapper that manages its authorization options
type authHandler struct {
	base http.Handler
	// This struct wraps the unique concurrency requirements of authHandlers. Concept is explained below the parent structures
	authFuncs authFuncLock
}

// NewAuthHandler sets the base http.Handler
func NewAuthHandler(handler http.Handler, callables ...authFuncCallable) *authHandler {
	h := &authHandler{base: handler, authFuncs: authFuncLock{mutex: new(sync.Mutex)}}
	h.authFuncs.componentsList = make(map[string]*authFuncList)
	h.authFuncs.componentsList[""], _ = NewAuthFuncList("")
	for i, _ := range callables {
		switch callables[i].(type) {
		case *authFuncListCore:
			// TODO: silent error woops- add it to instances?
		case *authFuncList:
			h.AddLists(*callables[i].(*authFuncList))
		case *authFuncInstance:
			h.AddInstances(callables[i].(*authFuncInstance))
		}
	}
	h.UpdateActiveList()
	return h
}

// GetBase returns the underlying http.Handler
func (h *authHandler) GetBase() http.Handler {
	return h.base
}

// GetBase sets the underlying http.Handler
func (h *authHandler) SetBase(handler http.Handler) {
	h.base = handler
}

func (h *authHandler) AddInstances(instances ...authFuncCallable) error {
	for i, _ := range instances {
		if _, ok := h.authFuncs.componentsList[""].funcMap[instances[i].(*authFuncInstance).name]; ok {
			return errors.Wrap(ErrNameTaken, instances[i].(*authFuncInstance).name)
		}
		h.authFuncs.componentsList[""].AddCallables(instances...)
		return nil
	}
	return nil
}

func (h *authHandler) RemoveInstances(instanceNames ...string) {
	h.authFuncs.componentsList[""].RemoveCallables(instanceNames...)
}

// UpdateHandler is UpdateHandlers for one handler
func (h *authHandler) UpdateHandler() (chan int, int) {
	completionNotifier := make(chan int, 1)
	go func() {
		lockNeeded := h.startLock()
		if lockNeeded {
			h.UpdateActiveList()
			h.endLock()
		}
		completionNotifier <- 1
	}()
	return completionNotifier, 1
}

func (h *authHandler) AddLists(lists ...authFuncList) error {
	for i, _ := range lists {
		if _, ok := h.authFuncs.componentsList[lists[i].name]; ok {
			return errors.Wrap(ErrNameTaken, lists[i].name)
		}
		lists[i].mutex.Lock()
		h.authFuncs.componentsList[lists[i].name] = &lists[i]
		lists[i].addHandler(h)
		lists[i].mutex.Unlock()
	}
	return nil
}

func (h *authHandler) RemoveLists(listNames ...string) {
	for _, v := range listNames {
		// lock lists
		h.authFuncs.componentsList[v].mutex.Lock()
		h.authFuncs.componentsList[v].removeHandler(h)
		h.authFuncs.componentsList[v].mutex.Unlock()
		delete(h.authFuncs.componentsList, v)
	}

}

// startLock is used to set updating as a priority and lock it. It returns false if the lock wasn't required as update was deemed redundant.
func (h *authHandler) startLock() bool {
	h.authFuncs.toUpdate = true
	h.authFuncs.mutex.Lock()
	if !h.authFuncs.toUpdate {
		h.authFuncs.mutex.Unlock()
		return false
	}
	h.authFuncs.toUpdate = false // corner case race condition with no negative impact. Could set to true right after this before we update, and then we don't really need to update but its neglibile issue.
	return true
}

// endLock is used to unlcok the mutex
func (h *authHandler) endLock() {
	h.authFuncs.mutex.Unlock()
}

func (h *authHandler) UpdateActiveList() error {
	h.authFuncs.wg[1-h.authFuncs.currentList].Wait()
	componentsListSlice := make([]authFuncCallable, len(h.authFuncs.componentsList))
	i := 0
	for k, _ := range h.authFuncs.componentsList {
		componentsListSlice[i] = h.authFuncs.componentsList[k]
		i++
	}
	var err error
	h.authFuncs.activeLists[1-h.authFuncs.currentList], err = newAuthFuncListCore("", componentsListSlice...)
	if err != nil {
		return err
	}
	h.authFuncs.currentList = 1 - h.authFuncs.currentList
	return nil
}
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: CORS- see authHandler todo0 we should have like a preflight function we can assign
	// TODO: awful structure
	currentList := -1
	for currentList != h.authFuncs.currentList {
		currentList := h.authFuncs.currentList
		h.authFuncs.wg[currentList].Add(1)
		if h.authFuncs.currentList != currentList {
			h.authFuncs.wg[currentList].Done()
			currentList = -1
			continue
		}
		for i, _ := range h.authFuncs.activeLists[currentList].funcList {
			ret, ans := h.authFuncs.activeLists[currentList].funcList[i].call(w, r)
			if (ret == AuthGranted) && (!ans) {
				h.base.ServeHTTP(w, r)
				h.authFuncs.wg[currentList].Done()
				return
			}
		}
		h.authFuncs.wg[currentList].Done()
		return
	}
	return // should never get here eh
}
