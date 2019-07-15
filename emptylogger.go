package authdoor

import (
	"os"
	"syscall"

	"go.uber.org/zap"
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
	stderr *os.File
}

func (l *simpleLogger) Init() error {
	l.stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/stderr")
	return nil
}

func (l *simpleLogger) Info(output string) error {
	_, err := l.stderr.WriteString(output + "\n")
	return err
}

func (l *simpleLogger) Error(output string) error {
	_, err := l.stderr.WriteString(output + "\n")
	return err
}

// TODO: times and errors

// zapWrap produces a uber-zap logging connection
type zapWrap struct {
	sugar       bool
	ZapLogger   *zap.Logger
	SugarLogger *zap.SugaredLogger
	InfoFunc    func(output string) error
	ErrorFunc   func(output string) error
}

// Init starts a production level zap logger, which we use since we don't use all the same logging levels as Zap.
func (z *zapWrap) Init() error {
	z.ZapLogger, _ = zap.NewProduction()
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
