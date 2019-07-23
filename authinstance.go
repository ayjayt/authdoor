package authdoor

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/ayjayt/ilog"
)

var (
	// ErrNameTaken is returned when someone tries to register an instance on a handler twice
	ErrNameTaken = errors.New("tried to create an auth function with the same name as an existing function")
)

// AuthFunc is any function that takes a response writer and request and returns information about auth (status and user info) as well as an error
type AuthFunc func(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error)

// AuthFuncReturn wraps all the relevenat return data from an AuthFunc
type AuthFuncReturn struct {
	// Auth represents whether or not access was granted etc
	Auth AuthStatus
	// Resp lets us know whether or not we've made a reply via HTTP
	Resp RespStatus
	// Info supplies any info about the user and auth method we want
	Info InstanceReturnInfo
}

// AuthStatus contains information from an AuthFunc about authorization status.
type AuthStatus uint8

const (
	// AuthFailed is returned by an AuthFunc if it couldn't determine the users identity.
	AuthFailed AuthStatus = iota
	// AuthGranted is returned by an AuthFunc it was determined the user is authorized
	AuthGranted
	// AuthDenied is returned by an AuthFunc essentially if we know the user is unauthorized.
	AuthDenied
)

// String provides aw ay to convert an AuthStatus to descriptive text- affects logs and errors
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

// RespStatus is returned "true" from an AuthFunc (see consts below) if we wrote to the ResponseWriter.
type RespStatus bool

const (
	// Answered is the value of ResponseStatus when we wrote to the ResponseWriter
	Answered RespStatus = true
	// Ignored is the value of ResponseStatus when we did not write to the ResponseWriter
	Ignored RespStatus = false
)

// String provides aw ay to convert an AuthStatus to descriptive text- affects logs and errors
func (r RespStatus) String() string {
	switch r {
	case Answered:
		return "Answered"
	case Ignored:
		return "Ignored"
	}
	return "Unknown"
}

// InstanceReturnInfo represents data from some AuthFunc
type InstanceReturnInfo struct {
	// name is unexported because we don't want people to change it- it comes right from the instance
	name string
	// Info would be arbitrary data supplied by the auth method
	Info json.RawMessage
}

// IsAnswered is a simple helper to check if the ResponseWriter was written to
func (r *AuthFuncReturn) IsAnswered() bool {
	return r.Resp == Answered
}

// IsDone is a helper method to let us know if we need to keep looping through auth functions
func (r *AuthFuncReturn) IsDone() bool {
	if r.Auth == AuthGranted || r.Auth == AuthDenied || r.IsAnswered() {
		return true
	}
	return false
}

// AuthFuncInstance is the structure actually used by a handler, it includes some meta data around the function.
type AuthFuncInstance struct {
	name     string
	authFunc AuthFunc
	priority int
	logger   ilog.LoggerInterface
}

// NewAuthFuncInstance takes some AuthFunc and lets you build an instance out of it.
func (i *AuthFuncInstance) Init(name string, authFunc AuthFunc, priority int, logger ilog.LoggerInterface) {
	if logger == nil {
		i.logger = defaultLogger
	} else {
		i.logger = logger
	}
	i.logger.Info("Creating instance called \"" + name + "\" with priority " + strconv.Itoa(priority))
	i.name = name
	i.authFunc = authFunc
	i.priority = priority
}

// call does the work of calling the auth function. It's a simple wrapper.
func (i *AuthFuncInstance) call(w http.ResponseWriter, r *http.Request) (AuthFuncReturn, error) {
	// Avoid logging in a hotpath?
	ret, err := i.authFunc(w, r)
	ret.Info.name = i.name
	return ret, err
}
