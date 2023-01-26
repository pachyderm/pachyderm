package log

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

func TestProto(t *testing.T) {
	// This is just a test for not panicking.  Eventually Proto() will be replaced by a protobuf
	// add-on that generates MarshalLogObject methods for each proto, which is much cleaner and
	// faster than this.
	ctx := TestParallel(t)
	Info(ctx, "some proto", Proto("version", &versionpb.Version{Major: 42}))
	Info(ctx, "some proto", Proto("version", (*versionpb.Version)(nil)))
	Info(ctx, "some proto without MarshalLogObject", Proto("int", &types.Int32Value{Value: 42}))
	Info(ctx, "some proto without MarshalLogObject", Proto("int", (*types.Int32Value)(nil)))

	Info(ctx, "int64", Proto("int", &types.Int64Value{Value: 42}))
	Info(ctx, "int64", Proto("int", (*types.Int64Value)(nil)))

	b := [4096]byte{}
	Info(ctx, "lots of bytes", Proto("bytes", &types.BytesValue{Value: b[:]}))
	Info(ctx, "empty bytes", Proto("bytes", &types.BytesValue{}))
	Info(ctx, "nil bytes", Proto("bytes", (*types.BytesValue)(nil)))
	Info(ctx, "some bytes", Proto("bytes", &types.BytesValue{Value: b[:31]}))
	Info(ctx, "some bytes", Proto("bytes", &types.BytesValue{Value: b[:32]}))
	Info(ctx, "some bytes", Proto("bytes", &types.BytesValue{Value: b[:33]}))

	badAny := &types.Any{
		TypeUrl: "totally invalid",
		Value:   []byte("many bytes are here"),
	}
	Info(ctx, "bad any", Proto("any", badAny))

	any, err := types.MarshalAny(&versionpb.Version{Major: 42})
	if err != nil {
		t.Fatal(err)
	}
	Info(ctx, "good any", Proto("any", any))
	Info(ctx, "nil any", Proto("any", (*types.Any)(nil)))

	Info(ctx, "empty", Proto("empty", &types.Empty{}))
	Info(ctx, "nil empty", Proto("empty", (*types.Empty)(nil)))

	Info(ctx, "duration", Proto("duration", types.DurationProto(24*time.Hour)))
	Info(ctx, "nil duration", Proto("duration", (*types.Duration)(nil)))

	Info(ctx, "time", Proto("time", types.TimestampNow()))
	Info(ctx, "nil time", Proto("time", (*types.Timestamp)(nil)))
}
