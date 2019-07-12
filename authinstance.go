package authdoor

import (
	"errors"
	"net/http"
)

var (
	// ErrNameTaken is returned when someone tries to register an instance on a handler twice
	ErrNameTaken = errors.New("tried to create an auth function with the same name as an existing function")
)

// AuthStatus contains information from an AuthFunc about authorization status.
type AuthStatus uint8

// ResponseStatus is returned "true" from an AuthFunc (see consts below) if we wrote to the ResponseWriter.
type ResponseStatus bool

const (
	// AuthFailed is returned by an AuthFunc if it couldn't determine the users identity.
	AuthFailed AuthStatus = iota
	// AuthGranted is returned by an AuthFunc it was determined the user is authorized
	AuthGranted
	// AuthDenied is returned by an AuthFunc essentially if we know the user is unauthorized.
	AuthDenied
)

const (
	// Answered is the value of ResponseStatus when we wrote to the ResponseWriter
	Answered ResponseStatus = true
	// Ignored is the value of ResponseStatus when we did not write to the ResponseWriter
	Ignored ResponseStatus = false
)

// AuthFunc is any function that takes a response writer and request and returns two state variables, AuthStatus and ResponseStatus. TODO: Probably need to return some user data.
type AuthFunc func(w http.ResponseWriter, r *http.Request) (AuthStatus, ResponseStatus)

// authFuncInstance is the structure actually used by a handler, it includes some meta data around the function.
type authFuncInstance struct {
	name     string
	authFunc AuthFunc
	priority int
}

// call does the work of calling the auth function. It's a simple wrapper.
func (i *authFuncInstance) call(w http.ResponseWriter, r *http.Request) (AuthStatus, ResponseStatus) {
	logger.Info("Calling an AuthFunc", "name", i.name, "priority", i.priority)
	return i.authFunc(w, r)
}

// NewAuthFuncInstance takes some AuthFunc and lets you build an instance out of it.
func NewAuthFuncInstance(name string, authFunc AuthFunc, priority int) authFuncInstance {
	logger.Info("Creating new AuthFuncInstance", "name", name, "priority", priority)
	return authFuncInstance{
		name:     name,
		authFunc: authFunc,
		priority: priority,
	}
}
