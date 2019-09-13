/*
Package authdoor provides a race-safe ordered index-map of functions called by an http.Handler
implementation to be used as authorization functions.
*/
package authdoor

import (
	"github.com/ayjayt/ilog"
)

// This test comment is just to provoke gomod into fetching the latest nested module version

// defaultLogger is a global default logger makes things a lot easier
var defaultLogger ilog.LoggerInterface

func init() {
	if defaultLogger == nil {
		defaultLogger = new(ilog.EmptyLogger)
	}
}

// SetDefaultLogger allows you set a logger like github.com/go-logr/zapr
func SetDefaultLogger(newLogger ilog.LoggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("Default logger set")
}
