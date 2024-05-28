package log

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
	ctx := TestParallel(context.Background(), t)
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

	goodAny, err := anypb.New(&versionpb.Version{Major: 42})
	if err != nil {
		t.Fatal(err)
	}
	Info(ctx, "good any", Proto("any", goodAny))
	Info(ctx, "nil any", Proto("any", (*anypb.Any)(nil)))

	Info(ctx, "empty", Proto("empty", &emptypb.Empty{}))
	Info(ctx, "nil empty", Proto("empty", (*emptypb.Empty)(nil)))

	Info(ctx, "duration", Proto("duration", durationpb.New(24*time.Hour)))
	Info(ctx, "nil duration", Proto("duration", (*durationpb.Duration)(nil)))

	Info(ctx, "time", Proto("time", timestamppb.Now()))
	Info(ctx, "nil time", Proto("time", (*timestamppb.Timestamp)(nil)))

	v := &versionpb.Version{}
	dynamic := dynamicpb.NewMessage(v.ProtoReflect().Descriptor())
	dynamic.Set(v.ProtoReflect().Descriptor().Fields().ByName("major"), protoreflect.ValueOf(uint32(3)))
	raw, err := protoToJSONMap(dynamic)
	if err != nil {
		t.Errorf("protoToJSONMap: %v", err)
	}
	if diff := cmp.Diff(map[string]any{"major": float64(3)}, raw); diff != "" {
		t.Errorf("protoToJSONMap (-want +got):\n%s", diff)
	}
	Info(ctx, "dynamic message", Proto("dynamic", dynamic))
	Info(ctx, "nil dynamic message", Proto("dynamic", &dynamicpb.Message{}))
}

func TestRedactedString(t *testing.T) {
	testData := []struct {
		name     string
		original string
		redact   string
		want     string
	}{
		{
			name:     "all empty",
			original: "",
			redact:   "",
			want:     "",
		},
		{
			name:     "empty original",
			original: "",
			redact:   "secret",
			want:     "",
		},
		{
			name:     "empty redact",
			original: "foo bar",
			redact:   "",
			want:     "foo bar",
		},
		{
			name:     "redact postgres dsn",
			original: "sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=supersecret connect_timeout=30 statement_cache_mode=describe",
			redact:   "supersecret",
			want:     "sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=[11 masked bytes] connect_timeout=30 statement_cache_mode=describe",
		},
		{
			name:     "redact postgres dsn twice",
			original: "sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=supersecret connect_timeout=30 statement_cache_mode=describe sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=supersecret connect_timeout=30 statement_cache_mode=describe",
			redact:   "supersecret",
			want:     "sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=[11 masked bytes] connect_timeout=30 statement_cache_mode=describe sslmode=disable host=pg-bouncer port=5432 dbname=pachyderm user=pachyderm password=[11 masked bytes] connect_timeout=30 statement_cache_mode=describe",
		},
		{
			name:     "nothing to redact",
			original: "foobar",
			redact:   "barbaz",
			want:     "foobar",
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			field := RedactedString("x", test.original, test.redact)
			if got, want := field.String, test.want; got != want {
				t.Errorf("redact %q in %q:\n  got: %v\n want: %v", test.redact, test.original, got, want)
			}
		})
	}
}
