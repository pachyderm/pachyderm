package prototime // import "go.pedge.io/proto/time"

// the functionality in here is moving directly to go.pedge.io/pb

import (
	"time"

	"go.pedge.io/pb/go/google/protobuf"
)

// TimeToTimestamp converts a go Time to a protobuf Timestamp.
func TimeToTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

// TimestampToTime converts a protobuf Timestamp to a go Time.
func TimestampToTime(timestamp *google_protobuf.Timestamp) time.Time {
	if timestamp == nil {
		return time.Unix(0, 0).UTC()
	}
	return time.Unix(
		timestamp.Seconds,
		int64(timestamp.Nanos),
	).UTC()
}

// TimestampLess returns true if i is before j.
func TimestampLess(i *google_protobuf.Timestamp, j *google_protobuf.Timestamp) bool {
	if j == nil {
		return false
	}
	if i == nil {
		return true
	}
	if i.Seconds < j.Seconds {
		return true
	}
	if i.Seconds > j.Seconds {
		return false
	}
	return i.Nanos < j.Nanos
}

// Now returns the current time as a protobuf Timestamp.
func Now() *google_protobuf.Timestamp {
	return TimeToTimestamp(time.Now().UTC())
}

// DurationToProto converts a go Duration to a protobuf Duration.
func DurationToProto(d time.Duration) *google_protobuf.Duration {
	return &google_protobuf.Duration{
		Seconds: int64(d) / int64(time.Second),
		Nanos:   int32(int64(d) % int64(time.Second)),
	}
}

// DurationFromProto converts a protobuf Duration to a go Duration.
func DurationFromProto(duration *google_protobuf.Duration) time.Duration {
	if duration == nil {
		return 0
	}
	return time.Duration((duration.Seconds * int64(time.Second)) + int64(duration.Nanos))
}
