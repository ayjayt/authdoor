package authdoor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	newLogger := new(ZapWrap)
	newLogger.Init()
	SetDefaultLogger(newLogger)
	defaultLogger.Info("authdoor_test.go set logger")
}

// TestSetDefaultLogger ensures set logger works by verifying init() ran correctly
func TestSetDefaultLogger(t *testing.T) {
	require.IsType(t, &ZapWrap{}, defaultLogger)
}
