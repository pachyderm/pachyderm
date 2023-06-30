// Package protoextensions is the runtime support code for protoc-gen-zap (in ../etc/proto).
// log.Proto also uses this code, so that embedded messages are marshaled the same way as raw
// messages.
package protoextensions

import (
	fmt "fmt"

	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// AddTimestamp encodes a google.protobuf.Timestamp.
func AddTimestamp(enc zapcore.ObjectEncoder, key string, ts *timestamppb.Timestamp) {
	if ts == nil {
		return
	}
	enc.AddTime(key, ts.AsTime())
}

// AddDuration encodes a google.protobuf.Duration.
func AddDuration(enc zapcore.ObjectEncoder, key string, dpb *durationpb.Duration) {
	if dpb == nil {
		return
	}
	enc.AddDuration(key, dpb.AsDuration())
}

// AddBytesValue encodes an abridged google.protobuf.BytesValue.
func AddBytesValue(enc zapcore.ObjectEncoder, key string, b *wrapperspb.BytesValue) {
	if b == nil {
		return
	}
	enc.AddObject(key, ConciseBytes(b.GetValue())) //nolint:errcheck
}

// AddBytes encodes an abridged []byte.
func AddBytes(enc zapcore.ObjectEncoder, key string, b []byte) {
	if len(b) > 32 {
		enc.AddObject(key, ConciseBytes(b)) //nolint:errcheck
		return
	}
	enc.AddBinary(key, b)
}

// AddAny encodes a google.protobuf.Any.
func AddAny(enc zapcore.ObjectEncoder, key string, a *anypb.Any) {
	if a == nil {
		return
	}
	msg, err := a.UnmarshalNew()
	if err != nil {
		enc.AddReflected(key, a) //nolint:errcheck
		return
	}
	if m, ok := msg.(zapcore.ObjectMarshaler); ok {
		enc.AddObject(key, m) //nolint:errcheck
	} else {
		enc.AddReflected(key, msg) //nolint:errcheck
	}
}

// AddInt64Value encodes a google.protobuf.Int64Value.
func AddInt64Value(enc zapcore.ObjectEncoder, key string, i *wrapperspb.Int64Value) {
	if i == nil {
		return
	}
	enc.AddInt64(key, i.GetValue())
}

// ConciseBytes is []byte that implements zap.ObjectMarshaler in a way that only prints the first 32
// of the provided bytes.
type ConciseBytes []byte

// MarshalLogObject implements zap.ObjectMarshaler.
func (b ConciseBytes) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if len(b) > 32 {
		enc.AddInt("len", len(b))
		enc.AddBinary("firstBytes", b[:32])
	} else {
		enc.AddBinary("bytes", b)
	}
	return nil
}

// AddHalfString adds the first half of a string, and a message saying how many bytes were omitted.
func AddHalfString(enc zapcore.ObjectEncoder, key, value string) {
	if value == "" {
		enc.AddString(key, "")
		return
	}
	enc.AddString(key, fmt.Sprintf("%s.../%d", value[:len(value)/2], len(value)))
}
