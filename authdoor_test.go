package authdoor

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

// TEstMain runs first just to see if we should turn on verbose logging during testing
func TestMain(t *testing.T) {
	if testing.Verbose() {
		fmt.Printf("Verbose...\n")
		newLogger := new(ZapWrap)
		err := newLogger.Init()
		if err != nil {
			panic(err)
		}
		SetDefaultLogger(newLogger)
		defaultLogger.Info("authdoor_test.go set logger")
	}
}

// TestSetDefaultLogger ensures set logger works by verifying init() ran correctly
func TestSetDefaultLogger(t *testing.T) {
	if testing.Verbose() {
		require.IsType(t, &ZapWrap{}, defaultLogger)
	} else {
		require.IsType(t, &EmptyLogger{}, defaultLogger)
	}
}
