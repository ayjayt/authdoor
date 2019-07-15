package authdoor

import (
	"go.uber.org/zap"
	"os"
)

type loggerInterface interface {
	Init() error
	Info(string) error
	Error(string) error
}

type emptyLogger struct{}

func (l *emptyLogger) Init() error {
	return nil
}

func (l *emptyLogger) Info(output string) error {
	return nil
}

func (l *emptyLogger) Error(output string) error {
	return nil
}

type simpleLogger struct {
	path string
	file *os.File
}

func (l *simpleLogger) Init() error {
	var err error
	if len(l.path) == 0 {
		l.file = os.Stderr
	} else {
		l.file, err = os.OpenFile(l.path, os.O_APPEND|os.O_WRONLY, 0644)
	}
	return err
}

func (l *simpleLogger) Info(output string) error {
	_, err := l.file.WriteString(output + "\n")
	return err
}

func (l *simpleLogger) Error(output string) error {
	_, err := l.file.WriteString(output + "\n")
	return err
}

// TODO: times and errors

// zapWrap produces a uber-zap logging connection
type zapWrap struct {
	sugar       bool
	paths       []string
	ZapLogger   *zap.Logger
	SugarLogger *zap.SugaredLogger
	InfoFunc    func(output string) error
	ErrorFunc   func(output string) error
}

// Init starts a production level zap logger, which we use since we don't use all the same logging levels as Zap.
func (z *zapWrap) Init() error {
	config := zap.NewProductionConfig()
	if len(z.paths) > 0 {
		config.OutputPaths = z.paths
	}
	z.ZapLogger, _ = config.Build()
	// ZapLogger, change file?
	if z.sugar {
		z.SugarLogger = z.ZapLogger.Sugar()
		z.InfoFunc = func(output string) error {
			z.SugarLogger.Info(output)
			return nil
		}

		z.ErrorFunc = func(output string) error {
			z.SugarLogger.Error(output)
			return nil
		}
	} else {
		z.InfoFunc = func(output string) error {
			z.ZapLogger.Info(output)
			return nil
		}

		z.ErrorFunc = func(output string) error {
			z.ZapLogger.Error(output)
			return nil
		}
	}
	return nil
}

func (z *zapWrap) Info(output string) error {
	z.InfoFunc(output)
	return nil
}

func (z *zapWrap) Error(output string) error {
	z.ErrorFunc(output)
	return nil
}
