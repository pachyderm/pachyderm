/*
Package syslog defines functionality for integration with syslog.
*/
package syslog
import (
	"log/syslog"

	"go.pedge.io/protolog"
)

var (
	globalMarshaller = protolog.NewTextMarshaller(
		protolog.MarshallerOptions{
			DisableTimestamp: true,
			DisableLevel:     true,
		},
	)
)

// PusherOptions defines options for constructing a new syslog protolog.Pusher.
type PusherOptions struct {
	Marshaller protolog.Marshaller
}

// NewPusher creates a new protolog.Pusher that logs using syslog.
func NewPusher(writer *syslog.Writer, options PusherOptions) protolog.Pusher {
	return newPusher(writer, options)
}
