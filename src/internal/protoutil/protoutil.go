// Package protoutil contains some utilities for interacting with protocol buffer objects.
package protoutil

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
)

// MustTime converts a google.protobuf.Timestamp to a time.Time, or panics.
func MustTime(tspb *types.Timestamp) time.Time {
	if tspb == nil {
		return time.Time{}
	}
	converted, err := types.TimestampFromProto(tspb)
	if err != nil {
		panic(fmt.Sprintf("convert google.protobuf.Timestamp to time.Time: %v", err))
	}
	return converted
}

// MustTimestamp converts a time.Time to a google.protobuf.Timestamp, or panics.
func MustTimestamp(ts time.Time) *types.Timestamp {
	if ts.IsZero() {
		return nil
	}
	converted, err := types.TimestampProto(ts)
	if err != nil {
		panic(fmt.Sprintf("convert time.Time to google.protobuf.Timestamp: %v", err))
	}
	return converted
}

// MustTimestampFromPointer converts a *time.Time to a google.protobuf.Timestamp, or panics.
func MustTimestampFromPointer(ts *time.Time) *types.Timestamp {
	if ts == nil {
		return nil
	}
	if ts.IsZero() {
		return nil
	}
	converted, err := types.TimestampProto(*ts)
	if err != nil {
		panic(fmt.Sprintf("convert time.Time to google.protobuf.Timestamp: %v", err))
	}
	return converted
}
