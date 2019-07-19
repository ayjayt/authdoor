package authdoor

import (
	"errors"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

// testLogger is an implmentation of a LoggerInterface but, like the "testing" package, requires one to set either the benchmark or testing variable.
type testLogger struct {
	t testing.TB
}

// NewTestLogger is a wrapper to create a logger for use with those tests
func NewTestLogger(t testing.TB) testLogger {
	ret := testLogger{t: t}
	ret.Init()
	return ret
}

// Info supplies the less severe testing method of testLogger
func (t *testLogger) Info(output string) {
	t.t.Log(output)
}

// Error supplies the most sever testing method of testLogger
func (t *testLogger) Error(output string) {
	t.t.Error(output)
}

// Init is the required Init function of a LoggerInterface for testLogger
func (t *testLogger) Init() error {
	if t.t == nil {
		return errors.New("You must set testLogger.t")
	}
	return nil
}

// This just ensures we're fufilling the interface we intend to at compile time
var _ LoggerInterface = &testLogger{}

// TestEmptyLogger tests the empty logger
func TestEmptyLogger(t *testing.T) {
	fakeLogger := new(EmptyLogger)
	err := fakeLogger.Init()
	require.Nil(t, err)
	fakeLogger.Info("nothing")
	fakeLogger.Error("nothing")
	// Output:
}

// TestSimpleLogger tests the simple logger
func TestSimpleLogger(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-output")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	fakeLogger := &SimpleLogger{Path: file.Name()}
	err = fakeLogger.Init()
	require.Nil(t, err)
	fakeLogger.Info("simpleLogger Info Test")
	fakeLogger.Error("simpleLogger Error Test")

	fakeLogger2 := &SimpleLogger{Path: "/dev/null"}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	fakeLogger2.Info("simpleLogger devnull test")
	fakeLogger2.Error("simpleLogger devnull test")
}

// TestZapLogger tests the zap logger
func TestZapLogger(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-output")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	fakeLogger := &ZapWrap{Paths: []string{file.Name()}}
	err = fakeLogger.Init()
	require.Nil(t, err)
	fakeLogger.Info("zapLogger Info Test")
	fakeLogger.Error("zapLogger Error Test")

	fakeLogger2 := &ZapWrap{Paths: []string{"/dev/null"}}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	fakeLogger2.Info("simpleLogger devnull test")
	fakeLogger2.Error("simpleLogger devnull test")
}

// TestZapSugarLogger tests the zap sugared logger
func TestZapSugarLogger(t *testing.T) {
	file, err := ioutil.TempFile("/tmp", "test-output")
	if err != nil {
		panic(err)
	}
	defer os.Remove(file.Name())
	fakeLogger := &ZapWrap{Sugar: true, Paths: []string{file.Name()}}
	err = fakeLogger.Init()
	require.Nil(t, err)
	fakeLogger.Info("zapSugaredLogger Info Test")
	fakeLogger.Error("zapSugaredLogger Error Test")

	fakeLogger2 := &ZapWrap{Paths: []string{"/dev/null"}, Sugar: true}
	err = fakeLogger2.Init()
	require.Nil(t, err)
	fakeLogger2.Info("simpleLogger devnull test")
	fakeLogger2.Error("simpleLogger devnull test")
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
		err := simpleLogger.Init()
		if err != nil {
			b.Error(err)
		}
		simpleLogger.Info("SimpleLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			simpleLogger.Info("simpleLogger.Info()")
		}
	})
	b.Run("Benchmark zap production logger", func(b *testing.B) {
		zapLogger := &ZapWrap{Paths: []string{"/dev/null"}}
		err := zapLogger.Init()
		if err != nil {
			b.Error(err)
		}
		zapLogger.Info("ZapLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			zapLogger.Info("zapLogger.Info()")
		}
	})
	b.Run("Benchmark zap sugared logger", func(b *testing.B) {
		sugaredLogger := &ZapWrap{Sugar: true, Paths: []string{"/dev/null"}}
		err := sugaredLogger.Init()
		if err != nil {
			b.Error(err)
		}
		sugaredLogger.Info("SugarLogger test")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sugaredLogger.Info("sugaredLogger.Info()")
		}
	})
}
