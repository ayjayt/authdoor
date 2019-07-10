package authdoor

import (
	"github.com/go-logr/logr"
)

// we can inject a logger like github.com/go-logr/zapr if we want
var logger logr.Logger

func init() {
	if logger == nil {
		logger = &emptyLogger{}
	}
}

// SetLogger allows you set a logger like github.com/go-logr/zapr
func SetLogger(logger logr.Logger) {
	logger = logger.WithName("authdoor.go")
}

// logger.Info("msg")
// logger.Error(err, "msg")
