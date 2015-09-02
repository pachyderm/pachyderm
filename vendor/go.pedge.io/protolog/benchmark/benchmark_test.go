package benchmark

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	stdlogrus "github.com/Sirupsen/logrus"
	stdglog "github.com/golang/glog"
	"github.com/peter-edge/go-google-protobuf"
	"github.com/stretchr/testify/require"
	"go.pedge.io/protolog"
	"go.pedge.io/protolog/glog"
	"go.pedge.io/protolog/logrus"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

const (
	s = "hello"
	d = 1234
)

var (
	arg1 = "foo"
	arg2 = &google_protobuf.Timestamp{Seconds: 1000, Nanos: 1000}

	foo = &Foo{
		One: "one",
		Two: 2,
		Bar: &Foo_Bar{
			One: "1",
			Two: "two",
		},
	}

	fooString = foo.String()
)

func BenchmarkProto(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Print(foo) }, false)
}

func BenchmarkProtoText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Print(foo) }, false)
}

func BenchmarkProtoStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(foo) }, false)
}

func BenchmarkProtoStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(foo) }, false)
}

func BenchmarkProtoLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Println(foo) }, true)
}

func BenchmarkProtoGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Println(foo) }, true)
}

func BenchmarkProtoStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(foo) }, true)
}

func BenchmarkThreadProto(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Print(foo) }, true)
}

func BenchmarkThreadProtoText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Print(foo) }, true)
}

func BenchmarkThreadProtoStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(foo) }, true)
}

func BenchmarkThreadProtoStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(foo) }, true)
}

func BenchmarkThreadProtoLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Println(foo) }, true)
}

func BenchmarkThreadProtoGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Println(foo) }, true)
}

func BenchmarkThreadProtoStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(foo) }, true)
}

func BenchmarkFieldProto(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.WithField("key", "value").Print(foo) }, false)
}

func BenchmarkFieldProtoText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.WithField("key", "value").Print(foo) }, false)
}

func BenchmarkFieldProtoStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.(*stdlogrus.Logger).WithField("key", "value").Println(foo) }, false)
}

func BenchmarkString(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Println(fooString) }, false)
}

func BenchmarkStringText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Println(fooString) }, false)
}

func BenchmarkStringStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(fooString) }, false)
}

func BenchmarkStringStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(fooString) }, false)
}

func BenchmarkStringLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Println(fooString) }, false)
}

func BenchmarkStringGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Println(fooString) }, false)
}

func BenchmarkStringStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(fooString) }, false)
}

func BenchmarkThreadString(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Println(fooString) }, true)
}

func BenchmarkThreadStringText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Println(fooString) }, true)
}

func BenchmarkThreadStringStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(fooString) }, true)
}

func BenchmarkThreadStringStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(fooString) }, true)
}

func BenchmarkThreadStringLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Infoln(fooString) }, true)
}

func BenchmarkThreadStringGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Infoln(fooString) }, true)
}

func BenchmarkThreadStringStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(fooString) }, true)
}

func BenchmarkFreeformf(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Printf("%s %d\n", s, d) }, false)
}

func BenchmarkFreeformfStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infof("%s %d\n", s, d) }, false)
}

func BenchmarkThreadFreeformf(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Printf("%s %d\n", s, d) }, true)
}

func BenchmarkThreadFreeformfStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infof("%s %d\n", s, d) }, true)
}

func BenchmarkFreeformln(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Println(arg1, arg2) }, false)
}

func BenchmarkFreeformlnText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Println(arg1, arg2) }, false)
}

func BenchmarkFreeformlnStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(arg1, arg2) }, false)
}

func BenchmarkFreeformlnStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(arg1, arg2) }, false)
}

func BenchmarkFreeformlnLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Println(arg1, arg2) }, true)
}

func BenchmarkFreeformlnGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Println(arg1, arg2) }, true)
}

func BenchmarkFreeformlnStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformln(b *testing.B) {
	runBenchmark(b, setupLogger, func(logger protolog.Logger) { logger.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnText(b *testing.B) {
	runBenchmark(b, setupLoggerText, func(logger protolog.Logger) { logger.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnStd(b *testing.B) {
	runBenchmarkStd(b, setupStdLogger, func(logger benchLogger) { logger.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnStdLogrus(b *testing.B) {
	runBenchmarkStd(b, setupStdLogrusLogger, func(logger benchLogger) { logger.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnLogrus(b *testing.B) {
	runBenchmarkLogrus(b, func() { protolog.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnGLog(b *testing.B) {
	runBenchmarkGLog(b, func() { protolog.Println(arg1, arg2) }, true)
}

func BenchmarkThreadFreeformlnStdGLog(b *testing.B) {
	runBenchmarkStdGLog(b, func() { stdglog.Infoln(arg1, arg2) }, true)
}

func setupLogger(b *testing.B) (string, *os.File, protolog.Logger) {
	tempDir, err := ioutil.TempDir("", "protolog")
	require.NoError(b, err)
	file, err := os.Create(filepath.Join(tempDir, "log.out"))
	require.NoError(b, err)
	logger := protolog.NewLogger(
		protolog.NewWritePusher(
			protolog.NewFileFlusher(file),
			protolog.WritePusherOptions{},
		),
		protolog.LoggerOptions{},
	).AtLevel(protolog.Level_LEVEL_INFO)
	return tempDir, file, logger
}

func setupLoggerText(b *testing.B) (string, *os.File, protolog.Logger) {
	tempDir, err := ioutil.TempDir("", "protolog")
	require.NoError(b, err)
	file, err := os.Create(filepath.Join(tempDir, "log.out"))
	require.NoError(b, err)
	logger := protolog.NewLogger(
		protolog.NewWritePusher(
			protolog.NewFileFlusher(file),
			protolog.WritePusherOptions{
				Marshaller: protolog.NewTextMarshaller(protolog.MarshallerOptions{}),
			},
		),
		protolog.LoggerOptions{},
	).AtLevel(protolog.Level_LEVEL_INFO)
	return tempDir, file, logger
}

func setupStdLogger(b *testing.B) (string, *os.File, benchLogger) {
	tempDir, err := ioutil.TempDir("", "protolog")
	require.NoError(b, err)
	file, err := os.Create(filepath.Join(tempDir, "log.out"))
	require.NoError(b, err)
	logger := log.New(file, "", log.LstdFlags)
	return tempDir, file, logger
}

func setupStdLogrusLogger(b *testing.B) (string, *os.File, benchLogger) {
	tempDir, err := ioutil.TempDir("", "protolog")
	require.NoError(b, err)
	file, err := os.Create(filepath.Join(tempDir, "log.out"))
	require.NoError(b, err)
	logger := stdlogrus.New()
	logger.Out = file
	return tempDir, file, logger
}

func runBenchmark(b *testing.B, setup func(*testing.B) (string, *os.File, protolog.Logger), run func(protolog.Logger), thread bool) {
	b.StopTimer()
	tempDir, _, logger := setup(b)
	b.StartTimer()
	if thread {
		var wg sync.WaitGroup
		wg.Add(b.N)
		for i := 0; i < b.N; i++ {
			go func() {
				run(logger)
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < b.N; i++ {
			run(logger)
		}
	}
	_ = logger.Flush()
	b.StopTimer()
	_ = os.RemoveAll(tempDir)
	b.StartTimer()
}

func runBenchmarkStd(b *testing.B, setup func(*testing.B) (string, *os.File, benchLogger), run func(benchLogger), thread bool) {
	b.StopTimer()
	tempDir, file, logger := setup(b)
	b.StartTimer()
	if thread {
		var wg sync.WaitGroup
		wg.Add(b.N)
		for i := 0; i < b.N; i++ {
			go func() {
				run(logger)
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < b.N; i++ {
			run(logger)
		}
	}
	_ = file.Sync()
	b.StopTimer()
	_ = os.RemoveAll(tempDir)
	b.StartTimer()
}

func runBenchmarkLogrus(b *testing.B, run func(), thread bool) {
	b.StopTimer()
	tempDir, err := ioutil.TempDir("", "protolog")
	require.NoError(b, err)
	file, err := os.Create(filepath.Join(tempDir, "log.out"))
	require.NoError(b, err)
	logrus.SetPusherOptions(
		logrus.PusherOptions{
			Out: file,
			Formatter: &stdlogrus.TextFormatter{
				ForceColors: true,
			},
		},
	)
	b.StartTimer()
	if thread {
		var wg sync.WaitGroup
		wg.Add(b.N)
		for i := 0; i < b.N; i++ {
			go func() {
				run()
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < b.N; i++ {
			run()
		}
	}
	_ = protolog.Flush()
	b.StopTimer()
	_ = os.RemoveAll(tempDir)
	b.StartTimer()
}

func runBenchmarkGLog(b *testing.B, run func(), thread bool) {
	b.StopTimer()
	glog.Register()
	b.StartTimer()
	if thread {
		var wg sync.WaitGroup
		wg.Add(b.N)
		for i := 0; i < b.N; i++ {
			go func() {
				run()
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < b.N; i++ {
			run()
		}
	}
	_ = protolog.Flush()
}

func runBenchmarkStdGLog(b *testing.B, run func(), thread bool) {
	if thread {
		var wg sync.WaitGroup
		wg.Add(b.N)
		for i := 0; i < b.N; i++ {
			go func() {
				run()
				wg.Done()
			}()
		}
		wg.Wait()
	} else {
		for i := 0; i < b.N; i++ {
			run()
		}
	}
	stdglog.Flush()
}

type benchLogger interface {
	Printf(format string, args ...interface{})
	Println(args ...interface{})
}
