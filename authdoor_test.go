package authdoor

import (
	"testing"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	SetDefaultLogger(zapr.NewLogger(zapLog))
	defaultLogger.Info("authdoor_test.go set logger")
}

// TestSetDefaultLogger ensures set logger works by verifying init() ran correctly
func TestSetDefaultLogger(t *testing.T) {
	zapLog, err := zap.NewDevelopment()
	require.Nil(t, err)
	realLogger := zapr.NewLogger(zapLog)
	require.IsType(t, realLogger, defaultLogger)
}

func BenchmarkLogger(b *testing.B) {
	b.Log("This benchmark tests various zap loggers to give an idea to the overhead added by logging. However, it logs to /dev/null instead of stdout which is two orders of magnitude faster than stdout")
	devConfig := zap.NewDevelopmentConfig()
	devConfig.OutputPaths = []string{"/dev/null"}
	devLogger, err := devConfig.Build()
	if err != nil {
		panic(err)
	}
	prodConfig := zap.NewProductionConfig()
	prodConfig.OutputPaths = []string{"/dev/null"}
	prodLogger, err := prodConfig.Build()
	if err != nil {
		panic(err)
	}
	b.Run("Benchmark zap development logger", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			devLogger.Info("This is to benchmark logging, but output to stdout makes it way too verbose")
		}
	})
	b.Run("Benchmark zap production logger", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			prodLogger.Info("This is to benchmark logging, but output to stdout makes it way too verbose")
		}
	})
	// TODO: maybe open fake terminal and output there? github.com/creack/pty
}
