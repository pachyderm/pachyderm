package store

import (
	"time"

	"github.com/peter-edge/go-google-protobuf"
)

var (
	defaultTimer = &systemTimer{}
)

func timeToTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

type timer interface {
	Now() time.Time
}

type systemTimer struct{}

func (t *systemTimer) Now() time.Time {
	return time.Now().UTC()
}
