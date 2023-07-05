// Package protoutil contains some utilities for interacting with protocol buffer objects.
package protoutil

import (
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MustTime converts a google.protobuf.Timestamp to a time.Time.
func MustTime(tspb *timestamppb.Timestamp) time.Time {
	if tspb == nil {
		return time.Time{}
	}
	return tspb.AsTime()
}

// MustTimestamp converts a time.Time to a google.protobuf.Timestamp, or panics.
func MustTimestamp(ts time.Time) *timestamppb.Timestamp {
	if ts.IsZero() {
		return nil
	}
	return timestamppb.New(ts)
}

// MustTimestampFromPointer converts a *time.Time to a google.protobuf.Timestamp, or panics.
func MustTimestampFromPointer(ts *time.Time) *timestamppb.Timestamp {
	if ts == nil {
		return nil
	}
	if ts.IsZero() {
		return nil
	}
	return timestamppb.New(*ts)
}

// Clone clones a proto in a type-safe manner.
func Clone[T proto.Message](x T) T {
	return proto.Clone(x).(T)
}
