package authdoor

import (
	"sync"
)

// AuthFuncListTemplate wraps an AuthFuncListSafe in meta data and a list of handlers using it as a reference.
type AuthFuncListTemplate struct {
	name string
	AuthFuncListSafe
	handlerMutex *sync.RWMutex
	handlers     []*AuthHandler
}

// Init will initialize the AuthFuncListTemplate by calling it's underlying type's inits as well as initializing the wrapper-specific datatypes.
func (l *AuthFuncListTemplate) Init(name string, instances ...AuthFuncInstance) error {
	l.name = name
	l.handlerMutex = new(sync.RWMutex)
	l.handlers = make([]*AuthHandler, 0)
	return l.AuthFuncListSafe.Init(instances...)
}

// AddHandler will add a pointer to the list of handlers.
func (l *AuthFuncListTemplate) AddHandler(handler *AuthHandler) {
	l.handlerMutex.Lock()
	l.AuthFuncListSafe.AuthFuncList.logger.Info("Adding a handler to " + l.name)
	defer l.handlerMutex.Unlock()
	l.handlers = append(l.handlers, handler)
}

// RemoveHandler will remove the handlers from the handler list.
func (l *AuthFuncListTemplate) RemoveHandler(handler *AuthHandler) {
	l.handlerMutex.Lock()
	l.AuthFuncListSafe.AuthFuncList.logger.Info("Removing a handler to " + l.name)
	defer l.handlerMutex.Unlock()
	for i, _ := range l.handlers {
		if l.handlers[i] == handler {
			l.handlers[len(l.handlers)-1], l.handlers[i] = nil, l.handlers[len(l.handlers)-1]
			l.handlers = l.handlers[:len(l.handlers)-1]
			return
		}
	}
}

// UpdateHandlers will go through each handler and tell it that it needs to be updated. It is asynchronous and returns a channel that will receive an integer value of a "1" for each handler that's updated- to a total of the second return value.
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

// BlockForUpdate will wait for all the handlers to update- and can use the two retrun values of UpdateHandlers. It does not support timeout concurrently, but absolutely should.
func (l *AuthFuncListTemplate) BlockForUpdate(completionNotifier chan int, totalHandlers int) {
	var handlersComplete int
	for i := range completionNotifier {
		handlersComplete += i
		if handlersComplete == totalHandlers {
			return
		}
	}
}
