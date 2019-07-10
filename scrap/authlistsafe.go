package authdoor

import (
	"net/http"
	"sync"
)

// AuthFuncListSafe provides concurency support to AuthFuncList.
type AuthFuncListSafe struct {
	AuthFuncList
	listMutex *sync.RWMutex // This pointer helps us avoiding copying a mutex
}

// Init will take all instances- so the values of authFuncList.funcList too- and merge everything into a new sorted authFuncList with it's own WaitGroup
func (l *AuthFuncListSafe) Init(instances ...authFuncInstance) error {
	l.listMutex = new(sync.RWMutex)
	return l.AuthFuncList.Init(instances...)
}

// AddInstances will add any AuthFuncInstance to it's own authFuncList, sorted properly.
func (l *AuthFuncListSafe) AddInstances(instances ...authFuncInstance) error {
	l.listMutex.RLock()
	ret := l.AuthFuncList.AddInstances(instances...)
	l.listMutex.RUnlock()
	return ret
}

// RemoveInstances can remove a AuthFuncList/Instance from it's list
func (l *AuthFuncListSafe) RemoveInstances(names ...string) {
	listMutex.RLock()
	l.AuthFuncList.RemoveInstances(names...)
	l.listMutex.RUnlock()
}

// Call
func (l *AuthFuncListSafe) Call(w http.ResponseWriter, r *http.Request, name string) (status AuthStatus, response ResponseStatus, err error) {
	l.RWMutex.RLock()
	status, response, err = l.AuthFuncList.Call(w, r)
	l.RWMutex.RUnlock()
	return status, response, err
}

// CallAll will iterate through the list and call each function
func (l *AuthFuncListSafe) CallAll(w http.ResponseWriter, r *http.Request) (status AuthStatus, response ResponseStatus, err error) {
	l.RWMutex.RLock()
	status, response, err = l.AuthFuncList.CallAll(w, r)
	l.RWMutex.RUnlock()
	return status, response, err
}

// GetFuncs from a the Safe returns a copy of the funclist. This is because we don't know how long the caller will take, and we want things to be deterministic. It also forces read-only.
func (l *AuthFuncListSafe) GetFuncs() []authFuncInstance {
	ret := make([]authFuncInstance, len(l.funcList))
	l.RWMutex.Lock()
	copy(ret, l.funcList)
	l.RWMutex.Unlock()
	return ret
}
