package basicpass

import (
	"fmt"
	"testing"

	"github.com/ayjayt/ilog"
	"github.com/stretchr/testify/require"
)

// TsstMain runs first just to see if we should turn on verbose logging during testing
func TestMain(t *testing.T) {
	if testing.Verbose() {
		fmt.Printf("Verbose...\n")
		newLogger := new(ilog.ZapWrap)
		err := newLogger.Init()
		if err != nil {
			panic(err)
		}
		SetDefaultLogger(newLogger)
		defaultLogger.Info("authfuncs/basicpass/main_test.go set logger")
	}
}

// TestSetDefaultLogger ensures set logger works by verifying init() ran correctly
func TestSetDefaultLogger(t *testing.T) {
	if testing.Verbose() {
		require.IsType(t, &ilog.ZapWrap{}, defaultLogger)
	} else {
		require.IsType(t, &ilog.EmptyLogger{}, defaultLogger)
	}
}
