package authdoor

import (
	"sync"
)

type AuthFuncListTemplate struct {
	name string
	AuthFuncListSafe
	handlerMutex *sync.RWMutex
	handlers     []*AuthHandler
}

// Init will take all instances- so the values of AuthFuncList.funcList too- and merge everything into a new sorted AuthFuncList with it's own WaitGroup
func (l *AuthFuncListTemplate) Init(name string, instances ...AuthFuncInstance) error {
	l.name = name
	l.handlerMutex = new(sync.RWMutex)
	l.handlers = make([]*AuthHandler, 0)
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
