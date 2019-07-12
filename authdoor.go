package authdoor

import (
	"github.com/go-logr/logr"
)

// A global default logger makes things a lot easier IMO
var logger logr.Logger

func init() {
	if logger == nil {
		logger = &emptyLogger{}
	}
}

// SetLogger allows you set a logger like github.com/go-logr/zapr
func SetLogger(newLogger logr.Logger) {
	logger = newLogger.WithName("authdoor.go")
	logger.Info("Logger set")
}

// logger.Info("msg")
// logger.Error(err, "msg")
