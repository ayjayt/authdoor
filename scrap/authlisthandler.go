package authdoor

import (
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

type AuthFuncListTemplate struct {
	name string
	AuthFuncListSafe
	handlerMutex *sync.RWMutex
	handlers     []*AuthHandler
}

// Init will take all instances- so the values of AuthFuncList.funcList too- and merge everything into a new sorted AuthFuncList with it's own WaitGroup
func (l *AuthFuncListTemplate) Init(instances ...authFuncInstance) error {
	l.handlerMutex = new(sync.RWMutex)
	l.handlers = make([]*AuthHandler, 1)
	return l.AuthFuncListSafe.Init(instances...)
}

// AddHandler will have the handler points to the list, intended to be called by AuthHandler.AddLists().
func (l *AuthFuncListTemplate) AddHandler(handler *AuthHandler) {
	l.handlerMutex.Lock()
	defer l.handlerMutex.Unlock()
	l.handlers = append(l.handlers, handler)
}

// RemoveHandler will remove the handlers from the list, intended to be called by AuthHandler.RemoveLists()
func (l *AuthFuncListTemplate) RemoveHandler(handler *AuthHandler) {
	l.handlerMutex.Lock()
	defer l.handlerMutex.Unlock()
	for i, _ := range l.handlers {
		if l.handlers[i] == handler {
			l.handlers[len(l.handlers)-1], l.handlers[i] = nil, l.handlers[len(l.handlers)-1]
			l.handlers = l.handlers[:len(l.handlers)-1]
			return
		}
	}
}

// UpdateHandlers will actually have the handler reorganize and rewrite the list that it is implementing
func (l *AuthFuncListTemplate) UpdateHandlers() (chan int, int) {
	l.handlerMutex.RLock()
	defer l.handlerMutex.RUnlock()
	totalHandlers := len(l.handlers)
	completionNotifier := make(chan int, totalHandlers)
	for i, _ := range l.handlers {
		go l.handlers[i].UpdateHandler(completionNotifier)
	}
	return completionNotifier, totalHandlers
}

// BlockForUpdate will wait for all the handlers to update- may be unnecessary
func (l *AuthFuncListTemplate) BlockForUpdate(completionNotifier chan int, totalHandlers int) {
	var handlersComplete int
	for i := range completionNotifier {
		handlersComplete += i
		if handlersComplete == totalHandlers {
			return
		}
	}
}

// handlerMutex is when you're RW's a template's handler list or using it
// listMutex is when you're updating the actual list
// writerMutex is when you're actually changing a handler's activeList or componentList (these would basically happen at the same time, unless you're updating the activeList after updating an actual list)

// AuthHandler is an http.Handler wrapper that manages its authorization options
type AuthHandler struct {
	base           http.Handler
	activeLists    [2]*AuthFuncListSafe // the lists actually being used
	activeMutex    *sync.RWMutex
	toUpdate       bool
	currentList    int                          // for directing readers
	componentMutex *sync.Mutex                  // for writing
	componentsList map[string]*AuthFuncListSafe // for default and external lists
}

// Init sets the base http.Handler
func (h *AuthHandler) Init(handler http.Handler) {
	h.base = handler
	h.componentMutex = new(sync.Mutex)
	h.componentsList = make(map[string]*AuthFuncList, 1)
	list := new(AuthFuncListSafe)
	err := list.Init()
	if err != nil {
		return nil
	}
	h.componentsList[""] = list
	// TODO activeLists
	// who is updating the lists
	// what do we do if there is nothing
	// TODO activeMutex
}

// GetBase returns the underlying http.Handler
func (h *AuthHandler) GetBase() http.Handler {
	return h.base
}

// GetBase sets the underlying http.Handler
func (h *AuthHandler) SetBase(handler http.Handler) {
	h.base = handler
}

// TODO writerMutex?- possibly conflict if something removes or deletes componentsList?- what if something is recompiling the active list
func (h *AuthHandler) AddInstances(instances ...authFuncInstance) error {
	return h.componentsList[""].AddInstances(instances...)
}

// RemoveInstances ... TODO
func (h *AuthHandler) RemoveInstances(instanceNames ...string) { // TODO writerMutex?
	h.componentsList[""].RemoveInstances(instanceNames...)
}

// UpdateHandler is UpdateHandlers for one handler
func (h *AuthHandler) UpdateHandler(completionNotifier chan int) {
	lockNeeded := h.startLock() // todo this is for the writerMutex- maybe it should be another TODO
	if lockNeeded {
		h.UpdateActiveList()
		h.endLock()
	}
	completionNotifier <- 1
}

func (h *AuthHandler) AddLists(lists ...AuthFuncList) error {
	h.componentMutex.Lock()
	defer h.componentMutex.Unlock()
	for i, _ := range lists {
		if _, ok := h.componentsList[lists[i].name]; ok {
			return errors.Wrap(ErrNameTaken, lists[i].name)
		}
		lists[i].AddHandler(h)
		h.componentsList[lists[i].name] = &lists[i]
		// We need to update our composite now
	}
	return nil
}

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

// UpdateActiveLists is what builds the components into your lists. It needs to be called when a list is updated or added.
func (h *AuthHandler) UpdateActiveList() error {
	h.componentMutex.Lock()
	// Not defered unlock because we unlock it sooner
	componentsListSlice := make([]authFuncInstance, len(h.componentsList), len(h.componentsList)*3) // Arbitrary constant to extend capacity
	for _, i := range h.componentsList {
		componentsListSlice = append(componentsListSlice, h.componentsList[i]...)
	}
	h.componentMutex.Unlock()
	h.activeMutex.Lock()
	defer h.activeMutex.Unlock()
	h.activeLists[1-h.currentList] = new(AuthFuncListSafe)
	var err error
	h.activeLists[1-h.currentList], err = h.activeLists[1-h.currentList].Init(componentsListSlice...) // why can't i use := here
	if err != nil {
		return err
	}
	h.currentList = 1 - h.currentList
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
		ret, ans, err := h.activeList[currentList].CallAll(w, r)
		if err != nil {
			return
		}
		if (ret == AuthGranted) || (ans == Ignored) {
			h.base.ServeHTTP(w, r)
		}
		return
	}
	return // should never get here eh
}
