package authdoor

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

var (
	// ErrNameTaken is returned when someone tries to register an instance on a handler twice
	ErrNameTaken = errors.New("tried to create an auth function with the same name as an existing function")
)

// AuthStatus contains information from an AuthFunc about authorization status.
type AuthStatus uint8

// RespStatus is returned "true" from an AuthFunc (see consts below) if we wrote to the ResponseWriter.
type RespStatus bool

const (
	// AuthFailed is returned by an AuthFunc if it couldn't determine the users identity.
	AuthFailed AuthStatus = iota
	// AuthGranted is returned by an AuthFunc it was determined the user is authorized
	AuthGranted
	// AuthDenied is returned by an AuthFunc essentially if we know the user is unauthorized.
	AuthDenied
)

func (a AuthStatus) String() string {
	switch a {
	case AuthFailed:
		return "AuthFailed"
	case AuthGranted:
		return "AuthGranted"
	case AuthDenied:
		return "AuthDenied"
	}
	return "Unknown"
}

const (
	// Answered is the value of ResponseStatus when we wrote to the ResponseWriter
	Answered RespStatus = true
	// Ignored is the value of ResponseStatus when we did not write to the ResponseWriter
	Ignored RespStatus = false
)

func (r RespStatus) String() string {
	switch r {
	case Answered:
		return "Answered"
	case Ignored:
		return "Ignored"
	}
	return "Unknown"
}

type InstanceReturnInfo struct {
	name string
	Info json.RawMessage
}

type AuthFuncReturn struct {
	Auth AuthStatus
	Resp RespStatus
	Info InstanceReturnInfo
}

func (r *AuthFuncReturn) IsDone() bool {
	if r.Auth == AuthGranted || r.Auth == AuthDenied || r.Resp == Answered {
		return true
	}
	return false
}

// AuthFunc is any function that takes a response writer and request and returns two state variables, AuthStatus and ResponseStatus. TODO: Probably need to return some user data.
type AuthFunc func(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error)

// AuthFuncInstance is the structure actually used by a handler, it includes some meta data around the function.
type AuthFuncInstance struct {
	name     string
	authFunc AuthFunc
	priority int
	logger   LoggerInterface
}

// call does the work of calling the auth function. It's a simple wrapper.
func (i *AuthFuncInstance) call(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error) {
	// i.logger.Info("Calling an AuthFunc", "name", i.name, "priority", i.priority) - this logger is causing allocs // TODO
	ret, err := i.authFunc(w, r) // this will also need to return some data wrapped
	ret.Info.name = i.name
	return ret, err
}

// NewAuthFuncInstance takes some AuthFunc and lets you build an instance out of it.
func (i *AuthFuncInstance) Init(name string, authFunc AuthFunc, priority int, logger LoggerInterface) {
	if logger == nil {
		logger = defaultLogger
	}
	i.logger = logger
	i.logger.Info("Creating instance called \"" + name + "\" with priority " + strconv.Itoa(priority))
	i.name = name
	i.authFunc = authFunc
	i.priority = priority
}
