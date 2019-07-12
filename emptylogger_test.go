package authdoor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestEmptyLogger tests the logger global and then sets it for our use
func TestEmptyLogger(t *testing.T) {
	fakeLogger := &emptyLogger{}
	fakeLogger.Info("nothing", "key", "value")
	isEnabled := fakeLogger.Enabled()
	require.Equal(t, false, isEnabled)
	fakeLogger.Error(nil, "nothing", "key", "value")
	newLogger := fakeLogger.V(10)
	require.Equal(t, fakeLogger, newLogger)
	newLogger = fakeLogger.WithValues("key", "value")
	require.Equal(t, fakeLogger, newLogger)
	newLogger = fakeLogger.WithName("nothing")
	require.Equal(t, fakeLogger, newLogger)
}

// TODO, maybe make a wrapper for T.Logf so we can pass the T.Logf in as the logger during test
