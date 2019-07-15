package authdoor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestEmptyLogger tests the empty logger
func TestEmptyLogger(t *testing.T) {
	fakeLogger := new(EmptyLogger)
	err := fakeLogger.Init()
	require.Nil(t, err)
	err = fakeLogger.Info("nothing")
	require.Nil(t, err)
	err = fakeLogger.Error("nothing")
	require.Nil(t, err)
	// Output:
}

// TestSimpleLogger tests the simple logger
func TestSimpleLogger(t *testing.T) {
	fakeLogger := new(SimpleLogger)
	fakeLogger.Init()
	err := fakeLogger.Info("simpleLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("simpleLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &SimpleLogger{Path: "/dev/null"}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// TestZapLogger tests the zap logger
func TestZapLogger(t *testing.T) {
	fakeLogger := new(ZapWrap)
	fakeLogger.Init()
	err := fakeLogger.Info("zapLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("zapLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &ZapWrap{Paths: []string{"/dev/null"}}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// TestZapSugarLogger tests the zap sugared logger
func TestZapSugarLogger(t *testing.T) {
	fakeLogger := &ZapWrap{Sugar: true}
	fakeLogger.Init()
	err := fakeLogger.Info("zapSugaredLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("zapSugaredLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &ZapWrap{Paths: []string{"/dev/null"}, Sugar: true}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// BenchmarkLogger works to output testing to /dev/null
func BenchmarkLogger(b *testing.B) {

	b.Run("Benchmark empty logger", func(b *testing.B) {
		EmptyLogger := new(EmptyLogger)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			EmptyLogger.Info("emptyLogger.Info()")
		}
	})
	b.Run("Benchmark simple logger", func(b *testing.B) {
		simpleLogger := &SimpleLogger{Path: "/dev/null"}
		simpleLogger.Init()
		simpleLogger.Info("SimpleLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			simpleLogger.Info("simpleLogger.Info()")
		}
	})
	b.Run("Benchmark zap production logger", func(b *testing.B) {
		zapLogger := &ZapWrap{Paths: []string{"/dev/null"}}
		zapLogger.Init()
		zapLogger.Info("ZapLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			zapLogger.Info("zapLogger.Info()")
		}
	})
	b.Run("Benchmark zap sugared logger", func(b *testing.B) {
		sugaredLogger := &ZapWrap{Sugar: true, Paths: []string{"/dev/null"}}
		sugaredLogger.Init()
		sugaredLogger.Info("SugarLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sugaredLogger.Info("sugaredLogger.Info()")
		}
	})
}
