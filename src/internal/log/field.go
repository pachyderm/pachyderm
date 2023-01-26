package log

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/protoextensions"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Proto is a Field containing a protocol buffer message.
func Proto(name string, msg proto.Message) Field {
	switch x := msg.(type) {
	case zapcore.ObjectMarshaler:
		return zap.Object(name, x)
	case *types.Empty:
		return zap.Skip()
	case *types.Timestamp:
		t, err := types.TimestampFromProto(x)
		if err != nil {
			return zap.Any(name, x)
		}
		return zap.Time(name, t)
	case *types.Duration:
		d, err := types.DurationFromProto(x)
		if err != nil {
			return zap.Any(name, x)
		}
		return zap.Duration(name, d)
	case *types.BytesValue:
		return zap.Object(name, protoextensions.ConciseBytes(x.GetValue()))
	case *types.Int64Value:
		return zap.Int64(name, x.GetValue())
	case *types.Any:
		var any types.DynamicAny
		if err := types.UnmarshalAny(x, &any); err != nil {
			return zap.Any(name, x)
		}
		return Proto(name, any.Message)
	}
	return zap.Any(name, msg)
}

type attempt struct{ i, max int }

// MarshalLogObject implements zapcore.ObjectMarshaler.
func (a attempt) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("attempt", a.i)
	enc.AddInt("totalAttempts", a.max)
	return nil
}

// RetryAttempt is a Field that encodes the current retry (0-indexed) and the total number of
// retries.  It's intended for a for loop where "i" is the loop iterator and "max" is the upper
// bound "i < max".
func RetryAttempt(i int, max int) Field {
	return zap.Inline(&attempt{i: i, max: max})
}
