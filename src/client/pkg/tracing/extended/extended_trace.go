package extended

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

const (
	// traceCtxKey is the grpc metadata key whose value is a ExtendedTrace
	// identifying the current RPC/commit
	traceCtxKey = "commit-trace"

	// TracesCollectionPrefix is the prefix associated with the 'traces'
	// collection in etcd (which maps pipelines and commits to extended traces)
	TracesCollectionPrefix = "commit_traces"

	// ExtendedTraceEnvVar determines how long extended traces are updated until
	// they're deleted from the cluster
	ExtendedTraceEnvVar = "PACH_EXTENDED_TRACE"
)

var (
	// CommitIDIndex is a secondary index for extended traces by the set of
	// commit IDs watched by the trace
	CommitIDIndex = &col.Index{
		Field: "CommitIDs",
		Multi: true,
	}

	// PipelineIndex is a secondary index for extended traces by the pipelint
	// watched by the trace
	PipelineIndex = &col.Index{
		Field: "Pipeline",
	}

	// TraceGetOpts are the default options for retrieving a trace from
	// 'TracesCol'
	TraceGetOpts = &col.Options{Target: etcd.SortByKey, Order: etcd.SortNone}
)

// TracesCol returns the etcd collection of extended traces
func TracesCol(c *etcd.Client) col.Collection {
	return col.NewEtcdCollection(c,
		TracesCollectionPrefix,
		[]*col.Index{CommitIDIndex, PipelineIndex},
		&TraceProto{},
		nil,
		nil)
}

// GetTraceFromCtx extracts any extended trace embeded in 'ctx's metadata
func GetTraceFromCtx(ctx context.Context) (*TraceProto, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil // no trace
	}
	vals := md.Get(traceCtxKey)
	if len(vals) < 1 {
		return nil, nil // no trace
	}
	if len(vals) > 1 {
		return nil, errors.Errorf("ctx should have at most one extended trace, but found %d", len(vals))
	}
	marshalledProto, err := base64.URLEncoding.DecodeString(vals[0])
	if err != nil {
		return nil, errors.Wrapf(err, "error base64-decoding marshalled ExtendedTrace proto")
	}
	extendedTrace := &TraceProto{}
	if err := extendedTrace.Unmarshal(marshalledProto); err != nil {
		return nil, errors.Wrapf(err, "error unmarshalling extended trace from ctx")

	}
	return extendedTrace, nil
}

// TraceIn2Out copies any extended traces from the incoming RPC context in 'ctx'
// into the outgoing RPC context in 'ctx'. Currently, this is only called by
// CreatePipeline, when it forwards extended contexts to the PutFile RPC with
// the new commit info.
func TraceIn2Out(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx // no trace
	}

	// Expected len('vals') is 1, but don't bother verifying
	vals := md.Get(traceCtxKey)
	pairs := make([]string, 2*len(vals))
	for i := 0; i < len(pairs); i += 2 {
		pairs[i], pairs[i+1] = traceCtxKey, vals[i/2]
	}
	return metadata.AppendToOutgoingContext(ctx, pairs...)
}

func (t *TraceProto) isValid() bool {
	return len(t.SerializedTrace) > 0
}

// AddPipelineSpanToAnyTrace finds any extended traces associated with
// 'pipeline', and if any such trace exists, it creates a new span associated
// with that trace and returns it
func AddPipelineSpanToAnyTrace(ctx context.Context, c *etcd.Client,
	pipeline, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	var tracesFound int
	var extendedTrace TraceProto
	tracesCol := TracesCol(c).ReadOnly(ctx)
	if err := tracesCol.GetByIndex(PipelineIndex, pipeline, &extendedTrace, TraceGetOpts,
		func(key string) error {
			if tracesFound++; tracesFound > 1 {
				log.Errorf("found second, unexpected span with key %q (using new key)", key)
			}
			return nil
		}); err != nil {
		log.Errorf("error getting trace via pipeline %q: %v", pipeline, err)
		return nil, ctx
	}
	if !extendedTrace.isValid() {
		return nil, ctx // no trace found
	}

	// Deserialize opentracing span from 'extendedTrace'
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
	if err != nil {
		log.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
		return nil, ctx
	}

	// return new span
	span, ctx := opentracing.StartSpanFromContext(ctx,
		operation, opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{"pipeline", pipeline})
	tracing.TagAnySpan(span, kvs...)
	return span, ctx
}

// StartAnyExtendedTrace adds a new trace to 'ctx' (and returns an augmented
// context) based on whether the environment variable in 'targetRepoEnvVar' is
// set.
// Returns a context that may have the new span attached, and 'true' if an an
// extended trace was created, or 'false' otherwise
func StartAnyExtendedTrace(ctx context.Context, operation string, pipeline string) (newCtx context.Context, ok bool) {
	if _, ok := os.LookupEnv(ExtendedTraceEnvVar); !ok || !tracing.IsActive() {
		return ctx, false
	}

	// Create trace
	clientSpan, ctx := opentracing.StartSpanFromContext(
		ctx, operation, ext.SpanKindRPCClient,
		opentracing.Tag{string(ext.Component), "gRPC"},
		opentracing.Tag{"pipeline", pipeline})
	defer clientSpan.Finish()

	// embed extended trace proto in RPC context
	extendedTrace := TraceProto{
		SerializedTrace: map[string]string{}, // init map
		Pipeline:        pipeline,
	}
	opentracing.GlobalTracer().Inject(
		clientSpan.Context(),
		opentracing.TextMap,
		opentracing.TextMapCarrier(extendedTrace.SerializedTrace),
	)

	marshalledTrace, err := extendedTrace.Marshal()
	if err != nil {
		fmt.Printf("Warning: could not marshal commit trace proto (can only get intra-RPC trace): %v", err)
		return ctx, false
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		traceCtxKey, base64.URLEncoding.EncodeToString(marshalledTrace),
	))
	return ctx, true
}
