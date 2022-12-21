package log

import (
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type conciseBytes []byte

func (b conciseBytes) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("len", len(b))
	if len(b) > 30 {
		enc.AddBinary("firstBytes", b[:30])
	} else {
		enc.AddBinary("bytes", b)
	}
	return nil
}

// Proto is a Field containing a protocol buffer message.
func Proto(name string, msg proto.Message) Field {
	if bv, ok := msg.(*types.BytesValue); ok {
		return zap.Object(name, conciseBytes(bv.GetValue()))
	}
	if _, ok := msg.(*types.Empty); ok {
		return zap.Skip()
	}
	if a, ok := msg.(*types.Any); ok {
		var msg types.DynamicAny
		if err := types.UnmarshalAny(a, &msg); err == nil {
			return zap.Any(name, msg)
		}
	}

	// The plan is to use some sort of proto compiler that adds zap.ObjectMarshaler methods, and
	// then switch this to zap.Object.  The various marshalers available also have tools for
	// redacting sensitive data, so that will be nice.
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
