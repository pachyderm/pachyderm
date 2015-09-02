/*
Package glog defines functionality for integration with glog.
*/
package glog

import (
	"sync"

	"go.pedge.io/protolog"
)

var (
	globalLogDebug   = false
	globalMarshaller = protolog.NewTextMarshaller(
		protolog.MarshallerOptions{
			DisableTimestamp: true,
			DisableLevel:     true,
		},
	)
	globalLoggerOptions = protolog.LoggerOptions{}
	globalLock          = &sync.Mutex{}
)

// SetDebug sets whether to log events at the debug level.
func SetDebug(debug bool) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLogDebug = debug
	register()
}

// SetMarshaller sets the global protolog.Marshaller.
func SetMarshaller(marshaller protolog.Marshaller) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalMarshaller = marshaller
	register()
}

// SetLoggerOptions sets the global protolog.LoggerOptions.
func SetLoggerOptions(options protolog.LoggerOptions) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLoggerOptions = options
	register()
}

// Register registers the glog global logger as the protolog Logger.
func Register() {
	globalLock.Lock()
	defer globalLock.Unlock()
	register()
}

func register() {
	protolog.SetLogger(protolog.NewLogger(newPusher(globalMarshaller, globalLogDebug), globalLoggerOptions))
}
