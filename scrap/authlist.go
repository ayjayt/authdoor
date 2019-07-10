package authdoor

import (
	"net/http"
	"sort"

	"github.com/pkg/errors"
)

var (
	// ErrNotFound is returned by "call" when we ask for an AuthFuncInstance that doesn't exist
	ErrNotFound = errors.New("Name wasn't found")
)

// AuthFuncList is the basic idea of a list of iterable AuthFuncs.
type AuthFuncList struct {
	funcList []authFuncInstance // these are copied, and this needs to be reordered
	funcMap  map[string]int     // cornelk/hashmap would be faster
}

// Init will take all instances- so the values of AuthFuncList.funcList too- and merge everything into a new sorted AuthFuncList with it's own WaitGroup
func (l *AuthFuncList) Init(instances ...authFuncInstance) error {
	l.funcMap = make(map[string]int)
	err := l.AddInstances(instances...) // no lock because it's the first call and nothing has access to it yet
	return err
}

// GetFuncs returns the AuthFunc instances
func (l *AuthFuncList) GetFuncs() []authFuncInstance {
	return l.funcList // now we're accessing the underlying array. This is our read. Maybe it's to call them, maybe its to copy them.
	// If it's to Call: A handler will call it
	// If it's to Copy: Another list will call it.
}

// Len returns the length of the object to be sorted (used by sort.Sort)
func (l *AuthFuncList) Len() int {
	return len(l.funcList)
}

// Swap swaps the two objects by index (used by sort.Sort)
func (l *AuthFuncList) Swap(i, j int) {
	l.funcList[i], l.funcList[j] = l.funcList[j], l.funcList[i]
}

// Less is a comparison operator used by sort.Sort
func (l *AuthFuncList) Less(i, j int) bool {
	return l.funcList[i].priority < l.funcList[j].priority
}

// writeMap constructs the funcMap (to be used after sorting) which is used to quickly map a name of a AuthFunc to its index.
func (l *AuthFuncList) writeMap() {
	l.funcMap = make(map[string]int, len(l.funcList))
	for i, _ := range l.funcList {
		l.funcMap[l.funcList[i].name] = i
	}
}

// Sort implements stdlib's sort and then writes the map
func (l *AuthFuncList) sort() {
	sort.Sort(l)
	l.writeMap()
}

// Call is used to find an AuthFuncInstance by name and then call it.
func (l *AuthFuncList) Call(w http.ResponseWriter, r *http.Request, name string) (status AuthStatus, response ResponseStatus, err error) {
	instance, ok := l.funcMap[name]
	if !ok {
		return AuthFailed, Ignored, ErrNotFound
	}
	status, response = l.funcList[instance].call(w, r)
	return status, response, nil
}

// CallAll will iterate through the list and call each function
func (l *AuthFuncList) CallAll(w http.ResponseWriter, r *http.Request) (status AuthStatus, response ResponseStatus, err error) {
	for i, _ := range l.funcList {
		status, response = l.funcList[i].call(w, r)
		if (status == AuthFailed) && (response == Ignored) {
			continue
		}
		return status, response, nil
	}
	return AuthFailed, Ignored, nil
}

// AddInstances will add any AuthFuncInstance to it's own AuthFuncList, sorted properly.
func (l *AuthFuncList) AddInstances(instances ...authFuncInstance) error {
	for i, _ := range instances {
		if _, ok := l.funcMap[instances[i].name]; ok {
			return errors.Wrap(ErrNameTaken, instances[i].name)
		}
		l.funcMap[instances[i].name] = i
	}
	l.funcList = append(l.funcList, instances...) // are instances copied?
	l.sort()
	return nil
}

// RemoveInstances can remove a AuthFuncList/Instance from it's list
func (l *AuthFuncList) RemoveInstances(names ...string) {
	for i, _ := range names {
		l.funcList[l.funcMap[names[i]]].authFunc = nil // the function is set to nil, but not the two values
	}
	zombieCounter := 0
	newSize := 0
	for i, _ := range l.funcList {
		if l.funcList[i].authFunc == nil {
			l.funcList[i] = authFuncInstance{} // preventing a leak int he underlying array?
			zombieCounter++
		} else {
			newSize++
			if zombieCounter > 0 {
				l.funcList[i-zombieCounter] = l.funcList[i] // everything is shifted back
			}
		}
	}
	l.funcList = l.funcList[:newSize]
	l.writeMap()
}
