package log

import (
	"context"
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/constants"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/protoextensions"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc/metadata"
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

// Metadata is a Field that logs the provided metadata (canonicalizing keys, collapsing
// single-element values to strings, and removing the Pachyderm auth token).
func Metadata(name string, md metadata.MD) Field {
	return zap.Object(name, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		keys := maps.Keys(md)
		sort.Strings(keys)
		for _, k := range keys {
			v := md[k]
			if k == constants.ContextTokenKey {
				for i := range v {
					v[i] = "[MASKED]"
				}
			}
			switch len(v) {
			case 0:
				continue
			case 1:
				enc.AddString(strings.ToLower(k), v[0])
			default:
				if err := enc.AddArray(strings.ToLower(k), zapcore.ArrayMarshalerFunc(
					func(enc zapcore.ArrayEncoder) error {
						for _, x := range v {
							enc.AppendString(x)
						}
						return nil
					},
				)); err != nil {
					return errors.Wrap(err, "add metadata value array")
				}
			}
		}
		return nil
	}))
}

// OutgoingMetadata is a Field that logs the outgoing metadata associated with the provided context.
func OutgoingMetadata(ctx context.Context) Field {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return zap.Skip()
	}
	return Metadata("metadata", md)
}
