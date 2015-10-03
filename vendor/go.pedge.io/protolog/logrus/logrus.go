/*
Package logrus defines functionality for integration with Logrus.
*/
package logrus

import (
	"sync"

	"github.com/Sirupsen/logrus"

	"go.pedge.io/protolog"
)

var (
	globalPusherOptions = PusherOptions{}
	globalLoggerOptions = protolog.LoggerOptions{}
	globalLock          = &sync.Mutex{}
)

// PusherOptions defines options for constructing a new Logrus protolog.Pusher.
type PusherOptions struct {
	Out             protolog.WriteFlusher
	Hooks           []logrus.Hook
	Formatter       logrus.Formatter
	EnableID        bool
	DisableContexts bool
}

// NewPusher creates a new protolog.Pusher that logs using Logrus.
func NewPusher(options PusherOptions) protolog.Pusher {
	return newPusher(options)
}

// SetPusherOptions sets the global PusherOptions.
func SetPusherOptions(options PusherOptions) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalPusherOptions = options
	register()
}

// SetLoggerOptions set theglobal protolog.LoggerOptions.
func SetLoggerOptions(options protolog.LoggerOptions) {
	globalLock.Lock()
	defer globalLock.Unlock()
	globalLoggerOptions = options
	register()
}

// Register registers the logrus global Logger as the protolog Logger.
func Register() {
	globalLock.Lock()
	defer globalLock.Unlock()
	register()
}

func register() {
	protolog.SetLogger(protolog.NewLogger(NewPusher(globalPusherOptions), globalLoggerOptions))
}
