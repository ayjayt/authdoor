package authdoor

import (
	"net/http"
	"sync"
)

// AuthFuncListSafe provides concurency support to AuthFuncList.
type AuthFuncListSafe struct {
	AuthFuncList
	listMutex *sync.RWMutex // This pointer helps us avoiding copying a mutex if the strucutre is copied
}

// Init will create a new AuthFuncListSafe by calling AuthFuncList.Init()
func (l *AuthFuncListSafe) Init(instances ...AuthFuncInstance) error {
	l.listMutex = new(sync.RWMutex)
	return l.AuthFuncList.Init(instances...)
}

// AddInstances will add any AuthFuncInstance to it's own AuthFuncList, sorted properly.
func (l *AuthFuncListSafe) AddInstances(instances ...AuthFuncInstance) error {
	l.listMutex.RLock()
	ret := l.AuthFuncList.AddInstances(instances...)
	l.listMutex.RUnlock()
	return ret
}

// RemoveInstances can remove a AuthFuncInstance from the receiver AuthFuncList(Safe)
func (l *AuthFuncListSafe) RemoveInstances(names ...string) {
	l.listMutex.RLock()
	l.AuthFuncList.RemoveInstances(names...)
	l.listMutex.RUnlock()
}

// Call is a wrapper for AuthFuncList.Call with it's concurrency protection
func (l *AuthFuncListSafe) Call(w http.ResponseWriter, r *http.Request, name string) (ret AuthFuncReturn, err error) {
	l.listMutex.RLock()
	ret, err = l.AuthFuncList.Call(w, r, name)
	l.listMutex.RUnlock()
	return ret, err
}

// CallAll will iterate through the list and call each function, using mutexes to protect against cocurrent writes.
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
} // Do we need this? Is there a better way. Is this the better way, given how long Call takes.
