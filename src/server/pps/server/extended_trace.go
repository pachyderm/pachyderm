package server

import (
	"context"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	opentracing "github.com/opentracing/opentracing-go"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing/extended"
)

func addPipelineSpanToAnyExtendedTrace(ctx context.Context, c *etcd.Client, pipeline, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if !tracing.IsActive() {
		return nil, ctx // no Jaeger instance to send trace info to
	}

	// use key from 'ctx' to read full ExtendedTrace reference from etcd
	tracesCol := extended.TracesCol(c).ReadOnly(ctx)
	var extendedTrace extended.TraceProto
	var spanCtx opentracing.SpanContext
	// options can't be nil, but there should only be at most one trace attached
	// to 'pipeline' so the options shouldn't matter
	if err := tracesCol.GetByIndex(extended.PipelineIndex, pipeline, &extendedTrace, extended.TraceGetOpts,
		func(key string) error {
			if spanCtx != nil {
				return fmt.Errorf("second, unexpected span associated with pipeline %q: %s",
					pipeline, key)
			}

			// Create new opentracing span from 'extendedTrace'
			var err error
			spanCtx, err = opentracing.GlobalTracer().Extract(opentracing.TextMap,
				opentracing.TextMapCarrier(extendedTrace.SerializedTrace))
			if err != nil {
				return fmt.Errorf("could not extract span context from ExtendedTrace proto: %v", err)
			}
			return nil
		}); err != nil {
		log.Errorf("tracing error: %v", err)
		return nil, ctx
	}
	if spanCtx == nil {
		return nil, ctx
	}

	// TODO(msteffen) we may leak a trace if a user updates a pipeline with
	// reprocess=false (so worker never restarts) AND the pipeline has no output
	// commits (so worker.jobSpawner never reads the trace)

	// return new span
	options := []opentracing.StartSpanOption{
		opentracing.FollowsFrom(spanCtx),
		opentracing.Tag{"pipeline", pipeline},
	}
	// Construct new span's tags
	for i := 0; i < len(kvs); i += 2 {
		if len(kvs) == i+1 {
			// likely forgot key or value--best effort
			options = append(options, opentracing.Tag{"extra", fmt.Sprintf("%v", kvs[i])})
			break
		}
		key, ok := kvs[i].(string) // common case--key is string
		if !ok {
			key = fmt.Sprintf("%v", kvs[i])
		}
		val, ok := kvs[i+1].(string) // common case--val is string
		if !ok {
			val = fmt.Sprintf("%v", kvs[i+1])
		}
		options = append(options, opentracing.Tag{key, val})
	}
	span := opentracing.StartSpan(operation, options...)
	return span, opentracing.ContextWithSpan(ctx, span)
}
