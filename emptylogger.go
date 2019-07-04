package authdoor

import (
	"github.com/go-logr/logr"
)

type emptyLogger struct{}

// Info is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) Info(msg string, keysAndValues ...interface{}) { return }

// Enabled is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) Enabled() bool { return false }

// Error is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) Error(err error, msg string, keysAndValues ...interface{}) { return }

// V is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) V(level int) logr.InfoLogger { return e }

// WithValues is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) WithValues(keysAndValues ...interface{}) logr.Logger { return e }

// WithName is a shim to create an empty logging function to be used if none is supplied
func (e *emptyLogger) WithName(name string) logr.Logger { return e }
