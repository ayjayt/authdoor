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
func (l *AuthFuncListSafe) Init(instances ...AuthFuncInstance) error {
	l.listMutex = new(sync.RWMutex)
	return l.AuthFuncList.Init(instances...)
}

// AddInstances will add any AuthFuncInstance to it's own authFuncList, sorted properly.
func (l *AuthFuncListSafe) AddInstances(instances ...AuthFuncInstance) error {
	l.listMutex.RLock()
	ret := l.AuthFuncList.AddInstances(instances...)
	l.listMutex.RUnlock()
	return ret
}

// RemoveInstances can remove a AuthFuncList/Instance from it's list
func (l *AuthFuncListSafe) RemoveInstances(names ...string) {
	l.listMutex.RLock()
	l.AuthFuncList.RemoveInstances(names...)
	l.listMutex.RUnlock()
}

// Call
func (l *AuthFuncListSafe) Call(w http.ResponseWriter, r *http.Request, name string) (ret AuthFuncReturn, err error) {
	l.listMutex.RLock()
	ret, err = l.AuthFuncList.Call(w, r, name)
	l.listMutex.RUnlock()
	return ret, err
}

// CallAll will iterate through the list and call each function
func (l *AuthFuncListSafe) CallAll(w http.ResponseWriter, r *http.Request) (ret AuthFuncReturn, err error) {
	l.listMutex.RLock()
	ret, err = l.AuthFuncList.CallAll(w, r)
	l.listMutex.RUnlock()
	return ret, err
}

// GetFuncs from a the Safe returns a copy of the funclist. This is because we don't know how long the caller will take, and we want things to be deterministic. It also forces read-only.
func (l *AuthFuncListSafe) GetFuncs() []AuthFuncInstance {
	ret := make([]AuthFuncInstance, len(l.funcList))
	l.listMutex.Lock()
	copy(ret, l.funcList)
	l.listMutex.Unlock()
	return ret
} // Do we need this?
