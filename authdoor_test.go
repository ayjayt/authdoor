package authdoor

import (
	"testing"

	"github.com/stretchr/testify/require"
	//"go.uber.org/zap"
)

func init() {
	newLogger := &simpleLogger{}
	newLogger.Init()
	SetDefaultLogger(newLogger)
	defaultLogger.Info("authdoor_test.go set logger")
}

// TestSetDefaultLogger ensures set logger works by verifying init() ran correctly
func TestSetDefaultLogger(t *testing.T) {
	require.IsType(t, &simpleLogger{}, defaultLogger)
}
