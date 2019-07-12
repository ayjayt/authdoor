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
	SetLogger(zapr.NewLogger(zapLog))
	logger.Info("authdoor_test.go set logger")
}

// TestSetLogger ensures set logger works by verifying init() ran correctly
func TestSetLogger(t *testing.T) {
	zapLog, err := zap.NewDevelopment()
	require.Nil(t, err)
	realLogger := zapr.NewLogger(zapLog)
	require.IsType(t, realLogger, logger)
}
