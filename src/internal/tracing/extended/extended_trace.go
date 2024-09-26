// Package extended needs to be documented.
//
// TODO: document
package extended

import (
	"context"
	"os"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
)

const (
	// traceMDKey is the grpc metadata key whose value is a serialized
	// ExtendedTrace proto tied to the current CreatePipeline request. In a grpc
	// quirk, this key must end in '-bin' so that the value (a serialized
	// timestamp) is treated as arbitrary bytes and base-64 encoded before being
	// transmitted (see
	// https://github.com/grpc/grpc-go/blob/b2c5f4a808fd5de543c4e987cd85d356140ed681/Documentation/grpc-metadata.md)
	traceMDKey = "pipeline-trace-duration-bin"

	// tracesCollectionPrefix is the prefix associated with the 'traces'
	// collection in etcd (which maps pipelines and commits to extended traces)
	tracesCollectionPrefix = "commit_traces"

	// TraceDurationEnvVar determines whether a traced 'CreatePipeline' RPC is
	// propagated to the PPS master, and whether worker creation and such is
	// traced in addition to the original RPC. This value should be set to a
	// go duration to create an extended trace
	TraceDurationEnvVar = "PACH_TRACE_DURATION"

	// The default duration over which to conduct an extended trace (used if the
	// RPC's duration can't be parsed)
	defaultDuration = 5 * time.Minute
)

// TracesCol returns the etcd collection of extended traces
func TracesCol(c *etcd.Client) col.EtcdCollection {
	return col.NewEtcdCollection(c,
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
func PersistAny(ctx context.Context, c *etcd.Client, pipeline *pps.Pipeline) {
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
		log.Info(ctx, "multiple durations attached to extended trace", zap.String("pipeline", pipeline.GetName()), zap.String("usingDuration", vals[0]))
	}

	// Extended trace found, now create a span & persist it to etcd
	duration, err := time.ParseDuration(vals[0])
	if err != nil {
		log.Error(ctx, "could not parse extended span duration", zap.String("duration", vals[0]), zap.Error(err))
		return // Ignore extended trace attached to RPC
	}

	// serialize extended trace & write to etcd
	traceProto := &TraceProto{
		SerializedTrace: map[string]string{}, // init map
		Project:         pipeline.GetProject().GetName(),
		Pipeline:        pipeline.GetName(),
	}
	if err := opentracing.GlobalTracer().Inject(
		span.Context(), opentracing.TextMap,
		opentracing.TextMapCarrier(traceProto.SerializedTrace),
	); err != nil {
		log.Info(ctx, "could not inject context into GlobalTracer", zap.Error(err))
	}
	if _, err := col.NewSTM(ctx, c, func(stm col.STM) error {
		tracesCol := TracesCol(c).ReadWrite(stm)
		return errors.EnsureStack(tracesCol.PutTTL(ctx, pipeline.String(), traceProto, int64(duration.Seconds())))
	}); err != nil {
		log.Error(ctx, "could not persist extended trace for pipeline to etcd", zap.String("pipeline", pipeline.GetName()), zap.Error(err))
	}
}

func (t *TraceProto) isValid() bool {
	return len(t.SerializedTrace) > 0
}

// AddSpanToAnyPipelineTrace finds any extended traces associated with
// 'pipeline', and if any such trace exists, it creates a new span associated
// with that trace and returns it
func AddSpanToAnyPipelineTrace(ctx context.Context, c *etcd.Client,
	pipeline *pps.Pipeline, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	traceProto := &TraceProto{}
	tracesCol := TracesCol(c).ReadOnly()
	if err := tracesCol.Get(ctx, pipeline, traceProto); err != nil {
		if !col.IsErrNotFound(err) {
			log.Error(ctx, "error getting trace for pipeline", zap.String("pipeline", pipeline.GetName()), zap.Error(err))
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
		log.Error(ctx, "could not extract span context from ExtendedTrace proto", zap.Error(err))
		return nil, ctx
	}

	// return new span
	span, ctx := opentracing.StartSpanFromContext(ctx,
		operation, opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{Key: "project", Value: pipeline.Project.GetName()},
		opentracing.Tag{Key: "pipeline", Value: pipeline.Name})
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
