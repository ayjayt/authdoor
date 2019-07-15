package authdoor

import ()

// A global default logger makes things a lot easier
var defaultLogger loggerInterface

func init() {
	if defaultLogger == nil {
		defaultLogger = new(emptyLogger)
	}
}

// SetDefaultLogger allows you set a logger like github.com/go-logr/zapr
func SetDefaultLogger(newLogger loggerInterface) {
	defaultLogger = newLogger
	defaultLogger.Info("Default logger set")
}

// logger.Info("msg")
// logger.Error(err, "msg")
