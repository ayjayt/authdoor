package authdoor

import (
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

// handlerMutex is when you're RW's a template's handler list or using it
// listMutex is when you're updating the actual list
// writerMutex is when you're actually changing a handler's activeList or componentList (these would basically happen at the same time, unless you're updating the activeList after updating an actual list)

// AuthHandler is an http.Handler wrapper that manages its authorization options
type AuthHandler struct {
	base           http.Handler
	activeLists    [2]*AuthFuncListSafe // the lists actually being used
	activeMutex    *sync.RWMutex
	toUpdate       bool
	currentList    int                              // for directing readers
	componentMutex *sync.Mutex                      // for writing
	componentsList map[string]*AuthFuncListTemplate // for default and external lists
}

// Init sets the base http.Handler and initializes some variables
func (h *AuthHandler) Init(handler http.Handler) error {
	h.base = handler
	h.componentMutex = new(sync.Mutex)
	h.componentsList = make(map[string]*AuthFuncListTemplate, 1)
	list := new(AuthFuncListTemplate)
	err := list.Init("")
	if err != nil {
		return nil
	}
	h.componentsList[""] = list
	h.activeMutex = new(sync.RWMutex)
	// Will it be okay if neither activeLists slices are initialized?
	return nil
}

// GetBase returns the underlying http.Handler
func (h *AuthHandler) GetBase() http.Handler {
	return h.base
}

// GetBase sets the underlying http.Handler
func (h *AuthHandler) SetBase(handler http.Handler) {
	h.base = handler
}

// AddInstances adds instances to the default ListTemplate, I don't think this needs a mutex because the one underneath has one
func (h *AuthHandler) AddInstances(instances ...AuthFuncInstance) error {
	return h.componentsList[""].AddInstances(instances...)
}

// RemoveInstances ... TODO
func (h *AuthHandler) RemoveInstances(instanceNames ...string) { // TODO writerMutex?
	h.componentsList[""].RemoveInstances(instanceNames...)
}

// AddLists add lists
func (h *AuthHandler) AddLists(lists ...*AuthFuncListTemplate) error {
	h.componentMutex.Lock()
	defer h.componentMutex.Unlock()
	for i, _ := range lists {
		if _, ok := h.componentsList[lists[i].name]; ok {
			return errors.Wrap(ErrNameTaken, lists[i].name)
		}
		lists[i].AddHandler(h)
		h.componentsList[lists[i].name] = lists[i]
		// We need to update our composite now
	}
	return nil
}

// RemoveLists remove lists
func (h *AuthHandler) RemoveLists(listNames ...string) {
	h.componentMutex.Lock()
	defer h.componentMutex.Unlock()
	for _, v := range listNames {
		h.componentsList[v].RemoveHandler(h) // modifying the list, calls its own lox
		delete(h.componentsList, v)
	}
}

// startLock is used to set updating as a priority and lock it. It returns false if the lock wasn't required as update was deemed redundant.
func (h *AuthHandler) startLock() bool {
	h.toUpdate = true
	h.activeMutex.Lock()
	if !h.toUpdate {
		h.activeMutex.Unlock()
		return false
	}
	h.toUpdate = false
	// corner condition race case
	// i'd like to explore it more
	return true
}

// endLock is used to unlock the mutex
func (h *AuthHandler) endLock() {
	h.activeMutex.Unlock()
}

// UpdateHandler is what builds the components into your lists. It needs to be called when a list is updated or added.
func (h *AuthHandler) UpdateHandler(completionNotifier chan int) error {
	defer func() {
		completionNotifier <- 1
	}()
	h.componentMutex.Lock()
	// Not defered unlock because we unlock it sooner
	componentsListSlice := make([]AuthFuncInstance, len(h.componentsList), len(h.componentsList)*3)
	for _, i := range h.componentsList {
		componentsListSlice = append(componentsListSlice, h.componentsList[i.name].AuthFuncListSafe.GetFuncs()...)
	}
	h.componentMutex.Unlock()
	if h.startLock() {
		defer h.endLock()
		h.activeLists[1-h.currentList] = new(AuthFuncListSafe)
		err := h.activeLists[1-h.currentList].Init(componentsListSlice...) // why can't i use := here
		if err != nil {
			return err
		}
		h.currentList = 1 - h.currentList
	}
	return nil
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: CORS- see AuthHandler todo0 we should have like a preflight function we can assign
	// TODO: awful structure
	currentList := -1
	for currentList != h.currentList {
		currentList := h.currentList
		h.activeMutex.RLock()
		if h.currentList != currentList {
			h.activeMutex.RUnlock()
			currentList = -1
			continue
		}
		defer h.activeMutex.RUnlock()
		if h.activeLists[currentList] == nil {
			return
		}
		ret, err := h.activeLists[currentList].CallAll(w, r)
		if err != nil {
			return
		}
		if (ret.Auth == AuthGranted) || (ret.Resp == Ignored) {
			h.base.ServeHTTP(w, r)
		}
		return
	}
	return // should never get here eh
}
