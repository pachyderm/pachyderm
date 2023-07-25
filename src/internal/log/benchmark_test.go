package log

import (
	"errors"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func BenchmarkFields(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "debug", zap.Int("i", i))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkFieldsSampled(b *testing.B) {
	ctx, w := NewBenchLogger(true)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "debug", zap.Int("i", i))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkSpan(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	errEven := errors.New("even")
	for i := 0; i < b.N; i++ {
		func() (retErr error) {
			defer Span(ctx, "bench", zap.Int("i", i))()
			if i%2 == 0 {
				return errEven //nolint:wrapcheck
			}
			return nil
		}() //nolint:errcheck
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkSpanWithError(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	errEven := errors.New("even")
	for i := 0; i < b.N; i++ {
		func() (retErr error) {
			defer Span(ctx, "bench", zap.Int("i", i))(Errorp(&retErr))
			if i%2 == 0 {
				return errEven //nolint:wrapcheck
			}
			return nil
		}() //nolint:errcheck
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusFields(b *testing.B) {
	l := logrus.New()
	w := new(byteCounter)
	l.Formatter = &logrus.JSONFormatter{}
	l.Out = w
	l.Level = logrus.DebugLevel
	for i := 0; i < b.N; i++ {
		l.WithField("i", i).Debug("debug")
	}
	if w.c.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusSugar(b *testing.B) {
	l := logrus.New()
	w := new(byteCounter)
	l.Formatter = &logrus.JSONFormatter{}
	l.Out = w
	l.Level = logrus.DebugLevel
	for i := 0; i < b.N; i++ {
		l.Debugf("debug: %d", i)
	}
	if w.c.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkLogrusWrapper(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	l := NewLogrus(ctx)
	for i := 0; i < b.N; i++ {
		l.Debugf("debug: %d", i)
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

var bigProto = &pps.CreatePipelineRequest{
	Pipeline: &pps.Pipeline{
		Project: &pfs.Project{
			Name: "project",
		},
		Name: "pipeline",
	},
	Transform: &pps.Transform{
		Image: "debian",
		Cmd:   []string{"set +x", "cp /pfs/in /pfs/out -avr"},
		Secrets: []*pps.SecretMount{
			{
				Name:      "name",
				Key:       "key",
				MountPath: "path",
			},
		},
		MemoryVolume: true,
		Env:          map[string]string{"PATH": "/", "HOME": "/pfs"},
	},
	ResourceRequests: &pps.ResourceSpec{
		Cpu: 2,
	},
	DatumTimeout: durationpb.New(time.Hour),
	JobTimeout:   durationpb.New(24 * time.Hour),
	ParallelismSpec: &pps.ParallelismSpec{
		Constant: 10,
	},
	Tolerations: []*pps.Toleration{
		{
			Key:      "dedicated",
			Operator: pps.TolerationOperator_EXISTS,
			Effect:   pps.TaintEffect_NO_SCHEDULE,
		},
		{
			Key:               "NotReady",
			Operator:          pps.TolerationOperator_EXISTS,
			Effect:            pps.TaintEffect_NO_EXECUTE,
			TolerationSeconds: wrapperspb.Int64(60),
		},
	},
	PodPatch: "{}",
	Input: &pps.Input{
		Cross: []*pps.Input{
			{
				Pfs: &pps.PFSInput{
					Project: "project",
					Name:    "a",
					Glob:    "/*",
				}},
			{
				Pfs: &pps.PFSInput{
					Project: "project",
					Name:    "b",
					Glob:    "/*",
				}},
		},
	},
	DatumTries: 3,
}

func BenchmarkProtoReflect(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "proto", zap.Reflect("pipeline", bigProto))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkProtoObject(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "proto", zap.Object("pipeline", bigProto))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkProtoJSONEncode(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		j, err := protojson.Marshal(bigProto)
		if err != nil {
			panic(err)
		}
		Debug(ctx, "proto", zap.ByteString("json", j))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkProtoTextEncode(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		Debug(ctx, "proto", zap.Stringer("text", bigProto))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}

func BenchmarkProtoBinaryEncode(b *testing.B) {
	ctx, w := NewBenchLogger(false)
	for i := 0; i < b.N; i++ {
		m, err := proto.Marshal(bigProto)
		if err != nil {
			panic(err)
		}
		Debug(ctx, "proto", zap.ByteString("binary", m))
	}
	if w.Load() == 0 {
		b.Fatal("no bytes added to logger")
	}
}
