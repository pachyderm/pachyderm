/*
Package log wraps all logging for pachyderm. This allows us
to use other logging implementations in the future easily.
*/
package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	// DefaultLogger is the default logger used.
	DefaultLogger Logger = newLogger(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))

	globalLogger = DefaultLogger
	lock         = &sync.Mutex{}
)

// Logger is a super simple logger.
type Logger interface {
	io.Writer
	Print(args ...interface{})
	Printf(format string, args ...interface{})
}

// SetLogger sets the logger used by pachyderm.
func SetLogger(l Logger) {
	lock.Lock()
	defer lock.Unlock()
	globalLogger = l
}

// Writer returns an io.Writer for logging.
func Writer() io.Writer {
	return globalLogger
}

// Print prints a log message with fmt.Print semantics using the pachyderm Logger.
func Print(args ...interface{}) {
	globalLogger.Print(args...)
}

// Printf prints a log message with fmt.Printf semantics using the pachyderm Logger.
func Printf(format string, args ...interface{}) {
	globalLogger.Printf(format, args...)
}

type logger struct {
	*log.Logger
}

func newLogger(l *log.Logger) *logger {
	return &logger{l}
}

func (l *logger) Write(p []byte) (int, error) {
	l.Print(string(p))
	return len(p), nil
}
