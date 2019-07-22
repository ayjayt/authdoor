/*
Package authdoor provides a race-safe ordered index-map of functions called by an http.Handler
implementation to be used as authorization functions.
*/
package authdoor

import ()

// defaultLogger is a global default logger makes things a lot easier
var defaultLogger LoggerInterface

func init() {
	if defaultLogger == nil {
		defaultLogger = new(EmptyLogger)
	}
}

// SetDefaultLogger allows you set a logger like github.com/go-logr/zapr
func SetDefaultLogger(newLogger LoggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("Default logger set")
}
