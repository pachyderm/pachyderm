package protolog_testing

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"testing"
	"time"

	stdlogrus "github.com/Sirupsen/logrus"
	"go.pedge.io/protolog"
	"go.pedge.io/protolog/glog"
	"go.pedge.io/protolog/logrus"

	"github.com/stretchr/testify/require"
)

func TestRoundtripAndTextMarshaller(t *testing.T) {
	buffer := bytes.NewBuffer(nil)
	fakeTimer := newFakeTimer(0)
	logger := protolog.NewLogger(
		protolog.NewWritePusher(
			protolog.NewWriterFlusher(buffer),
			protolog.WritePusherOptions{},
		),
		protolog.LoggerOptions{
			IDAllocator: newFakeIDAllocator(),
			Timer:       fakeTimer,
		},
	)
	logger.Debug(
		&Foo{
			StringField: "one",
			Int32Field:  2,
		},
	)
	logger.Info(
		&Baz{
			Bat: &Baz_Bat{
				Ban: &Baz_Bat_Ban{
					StringField: "one",
					Int32Field:  2,
				},
			},
		},
	)
	writer := logger.InfoWriter()
	for _, s := range []string{
		"hello",
		"world",
		"writing",
		"strings",
		"is",
		"fun",
	} {
		_, _ = writer.Write([]byte(s))
	}
	logger.Infoln("a normal line")
	logger.WithField("someKey", "someValue").Warnln("a warning line")

	puller := protolog.NewReadPuller(
		buffer,
		protolog.ReadPullerOptions{},
	)
	writeBuffer := bytes.NewBuffer(nil)
	writePusher := protolog.NewWritePusher(
		protolog.NewWriterFlusher(writeBuffer),
		protolog.WritePusherOptions{
			Marshaller: protolog.NewTextMarshaller(
				protolog.MarshallerOptions{
					DisableTimestamp: true,
				},
			),
			Newline: true,
		},
	)
	for entry, pullErr := puller.Pull(); pullErr != io.EOF; entry, pullErr = puller.Pull() {
		require.NoError(t, pullErr)
		require.NoError(t, writePusher.Push(entry))
	}
	require.Equal(
		t,
		`DEBUG protolog.testing.Foo {"string_field":"one","int32_field":2}
INFO  protolog.testing.Baz {"bat":{"ban":{"string_field":"one","int32_field":2}}}
INFO  hello
INFO  world
INFO  writing
INFO  strings
INFO  is
INFO  fun
INFO  a normal line
WARN  a warning line contexts=[{"someKey":"someValue"}]
`,
		writeBuffer.String(),
	)
}

func TestPrintSomeStuff(t *testing.T) {
	testPrintSomeStuff(t, protolog.NewStandardLogger(protolog.NewFileFlusher(os.Stderr)))
}

func TestPrintSomeStuffLogrus(t *testing.T) {
	logrus.SetPusherOptions(logrus.PusherOptions{})
	testPrintSomeStuff(t, protolog.GlobalLogger())
}

func TestPrintSomeStuffLogrusForceColors(t *testing.T) {
	logrus.SetPusherOptions(
		logrus.PusherOptions{
			Formatter: &stdlogrus.TextFormatter{
				ForceColors: true,
			},
		},
	)
	testPrintSomeStuff(t, protolog.GlobalLogger())
}

func TestPrintSomeStuffGLog(t *testing.T) {
	require.NoError(t, flag.CommandLine.Set("logtostderr", "true"))
	glog.Register()
	testPrintSomeStuff(t, protolog.GlobalLogger())
}

func testPrintSomeStuff(t *testing.T, logger protolog.Logger) {
	logger.Debug(
		&Foo{
			StringField: "one",
			Int32Field:  2,
		},
	)
	logger.Info(
		&Baz{
			Bat: &Baz_Bat{
				Ban: &Baz_Bat_Ban{
					StringField: "one",
					Int32Field:  2,
				},
			},
		},
	)
	writer := logger.InfoWriter()
	for _, s := range []string{
		"hello",
		"world",
		"writing",
		"strings",
		"is",
		"fun",
	} {
		_, _ = writer.Write([]byte(s))
	}
	logger.Infoln("a normal line")
	logger.WithField("someKey", "someValue").WithField("someOtherKey", 1).Warnln("a warning line")
}

type fakeIDAllocator struct {
	value int32
}

func newFakeIDAllocator() *fakeIDAllocator {
	return &fakeIDAllocator{-1}
}

func (f *fakeIDAllocator) Allocate() string {
	return fmt.Sprintf("%d", atomic.AddInt32(&f.value, 1))
}

type fakeTimer struct {
	unixTimeUsec int64
}

func newFakeTimer(initialUnixTimeUsec int64) *fakeTimer {
	return &fakeTimer{initialUnixTimeUsec}
}

func (f *fakeTimer) Now() time.Time {
	return time.Unix(f.unixTimeUsec/int64(time.Second), f.unixTimeUsec%int64(time.Second)).UTC()
}

func (f *fakeTimer) Add(secondDelta int64, nanosecondDelta int64) {
	atomic.AddInt64(&f.unixTimeUsec, (secondDelta*int64(time.Second))+nanosecondDelta)
}
