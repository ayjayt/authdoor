package authdoor

import (
	"github.com/go-logr/logr"
)

// A global default logger makes things a lot easier
var defaultLogger logr.Logger

func init() {
	if defaultLogger == nil {
		defaultLogger = new(emptyLogger)
	}
}

// SetDefaultLogger allows you set a logger like github.com/go-logr/zapr
func SetDefaultLogger(newLogger logr.Logger) {
	defaultLogger = newLogger.WithName("authdoor")
	defaultLogger.Info("Default logger set")
}

// logger.Info("msg")
// logger.Error(err, "msg")
