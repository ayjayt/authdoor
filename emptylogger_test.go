package authdoor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestEmptyLogger tests the empty logger
func TestEmptyLogger(t *testing.T) {
	fakeLogger := &emptyLogger{}
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
	fakeLogger := &simpleLogger{}
	fakeLogger.Init()
	err := fakeLogger.Info("simpleLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("simpleLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &simpleLogger{path: "/dev/null"}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// TestZapLogger tests the zap logger
func TestZapLogger(t *testing.T) {
	fakeLogger := &zapWrap{}
	fakeLogger.Init()
	err := fakeLogger.Info("zapLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("zapLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &zapWrap{paths: []string{"/dev/null"}}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// TestZapSugarLogger tests the zap sugared logger
func TestZapSugarLogger(t *testing.T) {
	fakeLogger := &zapWrap{sugar: true}
	fakeLogger.Init()
	err := fakeLogger.Info("zapSugaredLogger Info Test")
	require.Nil(t, err)
	err = fakeLogger.Error("zapSugaredLogger Error Test")
	require.Nil(t, err)

	fakeLogger2 := &zapWrap{paths: []string{"/dev/null"}, sugar: true}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	err = fakeLogger2.Info("simpleLogger devnull test")
	require.Nil(t, err)
	err = fakeLogger2.Error("simpleLogger devnull test")
	require.Nil(t, err)
}

// BUG: These benchmarks only work if declared globally like this. This bug stopped occuring after I removed
// the re-opening of stderr. However, this function si here to show that it worked compared to the function below it,
// "Benchmark Logger"
/*
var GemptyLogger *emptyLogger
var GsimpleLogger *simpleLogger
var GzapLogger *zapWrap
var GsugaredLogger *zapWrap

func init() {
	GemptyLogger = &emptyLogger{}
	GsimpleLogger = &simpleLogger{}
	GsimpleLogger.Init()
	GsimpleLogger.Info("SimpleLogger test")
	GzapLogger = &zapWrap{}
	GzapLogger.Init()
	GzapLogger.Info("ZapLogger test")
	GsugaredLogger = &zapWrap{sugar: true}
	GsugaredLogger.Init()
	GsugaredLogger.Info("SugarLogger test")

}

// BenchmarkLogger will test empty logger, simple logger, and two types of zap loggers
func BenchmarkLoggerWorks(b *testing.B) {

	b.Run("Benchmark empty logger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GemptyLogger.Info("emptyLogger.Info()")
		}
	})
	b.Run("Benchmark simple logger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GsimpleLogger.Info("simpleLogger.Info()")
		}
	})
	b.Run("Benchmark zap production logger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GzapLogger.Info("zapLogger.Info()")
		}
	})
	b.Run("Benchmark zap sugared logger", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GsugaredLogger.Info("sugaredLogger.Info()")
		}
	})
	// TODO: maybe open fake terminal and output there? github.com/creack/pty
}
*/

// BenchmarkLogger works to output testing to /dev/null
func BenchmarkLogger(b *testing.B) {

	b.Run("Benchmark empty logger", func(b *testing.B) {
		emptyLogger := &emptyLogger{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			emptyLogger.Info("emptyLogger.Info()")
		}
	})
	b.Run("Benchmark simple logger", func(b *testing.B) {
		simpleLogger := &simpleLogger{path: "/dev/null"}
		simpleLogger.Init()
		simpleLogger.Info("SimpleLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			simpleLogger.Info("simpleLogger.Info()")
		}
	})
	b.Run("Benchmark zap production logger", func(b *testing.B) {
		zapLogger := &zapWrap{paths: []string{"/dev/null"}}
		zapLogger.Init()
		zapLogger.Info("ZapLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			zapLogger.Info("zapLogger.Info()")
		}
	})
	b.Run("Benchmark zap sugared logger", func(b *testing.B) {
		sugaredLogger := &zapWrap{sugar: true, paths: []string{"/dev/null"}}
		sugaredLogger.Init()
		sugaredLogger.Info("SugarLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sugaredLogger.Info("sugaredLogger.Info()")
		}
	})
	// TODO: maybe open fake terminal and output there? github.com/creack/pty
}
