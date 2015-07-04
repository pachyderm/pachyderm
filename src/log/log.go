/*
Package log wraps all logging for pachyderm. This allows us
to use other logging implementations in the future easily.
*/
package log

import (
	"log"
	"os"
	"sync"
)

var (
	// DefaultLogger is the default logger used.
	DefaultLogger Logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

	logger = DefaultLogger
	lock   = &sync.Mutex{}
)

// Logger is a super simple logger.
type Logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})
}

// SetLogger sets the logger used by pachyderm.
func SetLogger(l Logger) {
	lock.Lock()
	defer lock.Unlock()
	logger = l
}

// Print prints a log message with fmt.Print semantics using the pachyderm Logger.
func Print(args ...interface{}) {
	logger.Print(args...)
}

// Printf prints a log message with fmt.Printf semantics using the pachyderm Logger.
func Printf(format string, args ...interface{}) {
	logger.Printf(format, args...)
}
