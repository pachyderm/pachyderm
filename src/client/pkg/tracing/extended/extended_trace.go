package extended

import (
	"context"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

const (
	// traceMDKey is the grpc metadata key whose value is a serialized
	// ExtendedTrace proto tied to the current CreatePipeline request. In a grpc
	// quirk, this key must end in '-bin' so that the value (a serialized
	// timestamp) is treated as arbitrary bytes and base-64 encoded before being
	// transmitted (see
	// https://github.com/grpc/grpc-go/blob/b2c5f4a808fd5de543c4e987cd85d356140ed681/Documentation/grpc-metadata.md)
	traceMDKey = "pipeline-trace-duration-bin"

	// TracesCollectionPrefix is the prefix associated with the 'traces'
	// collection in etcd (which maps pipelines and commits to extended traces)
	tracesCollectionPrefix = "commit_traces"

	// ExtendedTraceEnvVar determines whether a traced 'CreatePipeline' RPC is
	// propagated to the PPS master, and whether worker creation and such is
	// traced in addition to the original RPC. This value should be set to a
	// go duration to create an extended trace
	TraceDurationEnvVar = "PACH_TRACE_DURATION"

	// The default duration over which to conduct an extended trace (used if the
	// RPC's duration can't be parsed)
	defaultDuration = 5 * time.Minute
)

// TracesCol returns the etcd collection of extended traces
func TracesCol(c *etcd.Client) col.Collection {
	return col.NewCollection(c,
		tracesCollectionPrefix,
		nil, // no indexes
		&TraceProto{},
		nil, // no key check (keys are pipeline names)
		nil) // no val check
}

// PersistAny copies any extended traces from the incoming RPC context in 'ctx'
// into etcd. Currently, this is only called by CreatePipeline, when it stores a
// trace for future updates by the PPS master and workers.  This function is
// best-effort, and therefore doesn't currently return an error. Any errors are
// logged, and then the given context is returned.
func PersistAny(ctx context.Context, c *etcd.Client, pipeline string) {
	if !tracing.IsActive() {
		return
	}
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		// No incoming trace, so nothing to propagate
		return
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return // no extended trace attached to RPC
	}

	// Expected len('vals') is 0 or 1
	vals := md.Get(traceMDKey)
	if len(vals) == 0 {
		return // no extended trace attached to RPC
	}
	if len(vals) > 1 {
		log.Warnf("Multiple durations attached to extended trace for %q, using %s", pipeline, vals[0])
	}

	// Extended trace found, now create a span & persist it to etcd
	duration, err := time.ParseDuration(vals[0])
	if err != nil {
		log.Errorf("could not parse extended span duration %q for pipeline %q: %v",
			vals[0], pipeline, err)
		return // Ignore extended trace attached to RPC
	}

	// serialize extended trace & write to etcd
	traceProto := &TraceProto{
		SerializedTrace: map[string]string{}, // init map
		Pipeline:        pipeline,
	}
	opentracing.GlobalTracer().Inject(
		span.Context(), opentracing.TextMap,
		opentracing.TextMapCarrier(traceProto.SerializedTrace),
	)
	if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
		tracesCol := TracesCol(c).ReadWrite(stm)
		return tracesCol.PutTTL(pipeline, traceProto, int64(duration.Seconds()))
	}); err != nil {
		log.Errorf("could not persist extended trace for pipeline %q to etcd: %v. ",
			vals[0], pipeline, err, defaultDuration)
	}
	return
}

func (t *TraceProto) isValid() bool {
	return len(t.SerializedTrace) > 0
}

// AddSpanToAnyPipelineTrace finds any extended traces associated with
// 'pipeline', and if any such trace exists, it creates a new span associated
// with that trace and returns it
func AddSpanToAnyPipelineTrace(ctx context.Context, c *etcd.Client,
	pipeline, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	traceProto := &TraceProto{}
	tracesCol := TracesCol(c).ReadOnly(ctx)
	if err := tracesCol.Get(pipeline, traceProto); err != nil {
		if !col.IsErrNotFound(err) {
			log.Errorf("error getting trace for pipeline %q: %v", pipeline, err)
		}
		return nil, ctx
	}
	if !traceProto.isValid() {
		return nil, ctx // no trace found
	}

	// Deserialize opentracing span from 'traceProto'
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap,
		opentracing.TextMapCarrier(traceProto.SerializedTrace))
	if err != nil {
		log.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
		return nil, ctx
	}

	// return new span
	span, ctx := opentracing.StartSpanFromContext(ctx,
		operation, opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{Key: "pipeline", Value: pipeline})
	tracing.TagAnySpan(span, kvs...)
	return span, ctx
}

// EmbedAnyDuration augments 'ctx' (and returns a new ctx) based on whether
// the environment variable in 'ExtendedTraceEnvVar' is set.  Returns a context
// that may have the new span attached, and 'true' if an an extended trace was
// created, or 'false' otherwise. Currently only called by the CreatePipeline
// cobra command
func EmbedAnyDuration(ctx context.Context) (newCtx context.Context, err error) {
	duration, ok := os.LookupEnv(TraceDurationEnvVar)
	if !ok {
		return ctx, nil // PACH_TRACE_DURATION is not set
	}
	if _, ok := os.LookupEnv(tracing.ShortTraceEnvVar); !ok {
		return ctx, errors.Errorf("cannot set %s without setting %s",
			TraceDurationEnvVar, tracing.ShortTraceEnvVar)
	}
	if _, err := time.ParseDuration(duration); err != nil {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceMDKey,
			defaultDuration.String()))
		return ctx, errors.Wrapf(err,
			"could not parse duration %q (using default duration %q)", duration, defaultDuration)
	}
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(traceMDKey, duration))
	return ctx, nil
}
