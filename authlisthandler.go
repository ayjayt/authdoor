package authdoor

import (
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

// AuthHandler is an http.Handler wrapper that manages its authorization options, and provides a double-buffered RW-race-safe structure heavily biased towards reads.
type AuthHandler struct {
	base           http.Handler
	activeLists    [2]*AuthFuncListSafe // the lists actually being used
	activeMutex    *sync.RWMutex
	toUpdate       bool
	currentList    int                              // for directing readers
	componentMutex *sync.Mutex                      // for writing
	componentsList map[string]*AuthFuncListTemplate // for default and external lists
	logger         LoggerInterface
}

// SetLogger sets a custom logger for this handler
func (h *AuthHandler) SetLogger(newLogger LoggerInterface) {
	h.logger = newLogger
}

// Init sets the base http.Handler and initializes all members that need to be- maps, slices, and pointers to sync primitives.
func (h *AuthHandler) Init(handler http.Handler) error {
	if h.logger == nil {
		h.logger = defaultLogger
	}
	h.logger.Info("Initializing a new handler")
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

// SetBase sets the underlying http.Handler
func (h *AuthHandler) SetBase(handler http.Handler) {
	h.base = handler
}

// AddInstances adds instances to the default ListTemplate, I don't think this needs a mutex because the one underneath has one
func (h *AuthHandler) AddInstances(instances ...AuthFuncInstance) error {
	return h.componentsList[""].AddInstances(instances...)
}

// RemoveInstances removes instances from the underlying default ListTemplate
func (h *AuthHandler) RemoveInstances(instanceNames ...string) { // TODO writerMutex?
	h.componentsList[""].RemoveInstances(instanceNames...)
}

// AddLists add ListTemplates to handler's list of temoplates. You must call UpdateHandler manually after this.
func (h *AuthHandler) AddLists(lists ...*AuthFuncListTemplate) error {
	h.componentMutex.Lock()
	defer h.componentMutex.Unlock()
	for i, _ := range lists {
		if _, ok := h.componentsList[lists[i].name]; ok {
			return errors.Wrap(ErrNameTaken, lists[i].name)
		}
		lists[i].AddHandler(h)
		h.componentsList[lists[i].name] = lists[i]
	}
	return nil
}

// RemoveLists removes added to the component list by name. You must call UpdateHnalder manually after this
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
	// NOTE: corner condition race case where we could have false updates
	// i'd like to explore it more
	return true
}

// endLock is used to unlock the mutex
func (h *AuthHandler) endLock() {
	h.activeMutex.Unlock()
}

// UpdateHandler is what builds the components into your lists. It needs to be called when a list is updated or added.
func (h *AuthHandler) UpdateHandler(completionNotifier chan int) error {
	if completionNotifier != nil {
		defer func() {
			completionNotifier <- 1
		}()
	}
	h.componentMutex.Lock()
	// Not defered unlock because we unlock it sooner
	componentsListSlice := make([]AuthFuncInstance, 0, len(h.componentsList)*3)
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

// ServeHTTP is the handler function that wraps the base ServeHTTP, while calling the authorization functions.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Set CORS here or force it elsewhere?
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
			// Set contex here.
			if h.base != nil {
				h.base.ServeHTTP(w, r)
			}
		}
		return
	}
	return
}
