package authdoor

import (
	"go.uber.org/zap"
	"os"
)

// LogerInterface defines a simple interface to be used for logging. NOTE: Originally logr was used, but more features lead to less efficiency.
type LoggerInterface interface {
	Init() error
	Info(string) error
	Error(string) error
}

// EmptyLogger is a logger that can be used to turn off logging entirely.
type EmptyLogger struct{}

// Init is EmptyLogger's blank Init method
func (l *EmptyLogger) Init() error {
	return nil
}

// Info is EmptyLogger's blank Info method
func (l *EmptyLogger) Info(output string) error {
	return nil
}

// Error is EmptyLogger's blank Error method
func (l *EmptyLogger) Error(output string) error {
	return nil
}

// SimpleLogger is a simple logger that writes to stderr or a path it's given. It is NOT safe for concurrent use.
type SimpleLogger struct {
	Path string
	file *os.File
}

// Init attaches SimpleLogger to some path
func (l *SimpleLogger) Init() error {
	var err error
	if len(l.Path) == 0 {
		l.file = os.Stderr
	} else {
		l.file, err = os.OpenFile(l.Path, os.O_APPEND|os.O_WRONLY, 0644)
	}
	return err
}

// Info writes the info string to the output for SimpleLogger
func (l *SimpleLogger) Info(output string) error {
	_, err := l.file.WriteString(output + "\n")
	return err
}

// Error writes the error string to the output for SimpleLogger
func (l *SimpleLogger) Error(output string) error {
	_, err := l.file.WriteString(output + "\n")
	return err
}

// ZapWrap produces a uber-zap logging connection
type ZapWrap struct {
	// Sugar is a flag to indicate whether we shoudl use a Sugared logger
	Sugar bool
	// Paths lets us set the logging paths, otherwise we use stderr
	Paths []string
	// ZapLogger is the underlying ZapLogger
	ZapLogger *zap.Logger
	// SugarLogger is the underlying SugaredLogger
	SugarLogger *zap.SugaredLogger
	// infoFunc is the function called by Info() method
	infoFunc func(output string) error
	// errorFunc is the function called by the Error() method
	errorFunc func(output string) error
}

// Init starts a production level zap logger, which we use since we don't use all the same logging levels as Zap. It will switch the info or error func depending on whether or not its a sugared logger.
func (z *ZapWrap) Init() error {
	config := zap.NewProductionConfig()
	if len(z.Paths) > 0 {
		config.OutputPaths = z.Paths
	}
	z.ZapLogger, _ = config.Build()
	if z.Sugar {
		z.SugarLogger = z.ZapLogger.Sugar()
		z.infoFunc = func(output string) error {
			z.SugarLogger.Info(output)
			return nil
		}

		z.errorFunc = func(output string) error {
			z.SugarLogger.Error(output)
			return nil
		}
	} else {
		z.infoFunc = func(output string) error {
			z.ZapLogger.Info(output)
			return nil
		}

		z.errorFunc = func(output string) error {
			z.ZapLogger.Error(output)
			return nil
		}
	}
	return nil
}

// Info is ZapWraps Info method, but just a wrapper for z.infoFunc
func (z *ZapWrap) Info(output string) error {
	return z.infoFunc(output)
}

// Error is ZapWraps Error method, just a wrapper for z.errorFunc
func (z *ZapWrap) Error(output string) error {
	return z.errorFunc(output)
}
