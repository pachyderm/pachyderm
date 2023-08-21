package log

import (
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestProto(t *testing.T) {
	// This is just a test for not panicking.  Eventually Proto() will be replaced by a protobuf
	// add-on that generates MarshalLogObject methods for each proto, which is much cleaner and
	// faster than this.
	ctx := TestParallel(t)
	Info(ctx, "some proto", Proto("version", &versionpb.Version{Major: 42}))
	Info(ctx, "some proto", Proto("version", (*versionpb.Version)(nil)))
	Info(ctx, "some proto without MarshalLogObject", Proto("int", wrapperspb.Int32(42)))
	Info(ctx, "some proto without MarshalLogObject", Proto("int", (*wrapperspb.Int32Value)(nil)))

	Info(ctx, "int64", Proto("int", wrapperspb.Int64(42)))
	Info(ctx, "int64", Proto("int", (*wrapperspb.Int64Value)(nil)))

	b := [4096]byte{}
	Info(ctx, "lots of bytes", Proto("bytes", wrapperspb.Bytes(b[:])))
	Info(ctx, "empty bytes", Proto("bytes", &wrapperspb.BytesValue{}))
	Info(ctx, "nil bytes", Proto("bytes", (*wrapperspb.BytesValue)(nil)))
	Info(ctx, "some bytes", Proto("bytes", wrapperspb.Bytes(b[:31])))
	Info(ctx, "some bytes", Proto("bytes", wrapperspb.Bytes(b[:32])))
	Info(ctx, "some bytes", Proto("bytes", wrapperspb.Bytes(b[:33])))

	badAny := &anypb.Any{
		TypeUrl: "totally invalid",
		Value:   []byte("many bytes are here"),
	}
	Info(ctx, "bad any", Proto("any", badAny))

	any, err := anypb.New(&versionpb.Version{Major: 42})
	if err != nil {
		t.Fatal(err)
	}
	Info(ctx, "good any", Proto("any", any))
	Info(ctx, "nil any", Proto("any", (*anypb.Any)(nil)))

	Info(ctx, "empty", Proto("empty", &emptypb.Empty{}))
	Info(ctx, "nil empty", Proto("empty", (*emptypb.Empty)(nil)))

	Info(ctx, "duration", Proto("duration", durationpb.New(24*time.Hour)))
	Info(ctx, "nil duration", Proto("duration", (*durationpb.Duration)(nil)))

	Info(ctx, "time", Proto("time", timestamppb.Now()))
	Info(ctx, "nil time", Proto("time", (*timestamppb.Timestamp)(nil)))
}
