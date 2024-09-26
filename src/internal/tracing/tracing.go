// Package tracing provides utilities related to Jaeger tracing.
package tracing

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// JaegerServiceName is the name pachd (and the pachyderm client) uses to
// describe itself when it reports traces to Jaeger
const JaegerServiceName = "pachd"

// If you have Jaeger deployed and the JAEGER_ENDPOINT environment variable set
// to the address of your Jaeger instance's HTTP collection API, setting this
// environment variable to "true" will cause pachyderm to attach a Jaeger trace
// to any RPCs that it sends (this is primarily intended to be set in pachctl
// though any binary that includes our go client library will be able to use
// this env var)
//
// Note that tracing calls can slow them down somewhat and make interesting
// traces hard to find in Jaeger, so you may not want this variable set for
// every call.
const jaegerEndpointEnvVar = "JAEGER_ENDPOINT"

// ShortTraceEnvVar is what the client reads to decide whether to send a trace.
// Below, this is implemented by the span inclusion func
// addTraceIfTracingEnabled() (which is itself used by the GRPC interceptor)
const ShortTraceEnvVar = "PACH_TRACE"

// jaegerOnce is used to ensure that the Jaeger tracer is only initialized once
var jaegerOnce sync.Once

// jaegerEndpoint is set using jaegerOnce on startup, and then returned by
// future calls to InstallJaegerTracerFromEnv
var jaegerEndpoint string

// TagAnySpan tags any span associated with 'spanBox' (which must be either a
// span itself or a context.Context) with 'kvs'
func TagAnySpan(spanBox interface{}, kvs ...interface{}) opentracing.Span {
	if spanBox == nil {
		return nil
	}

	// extract span from 'spanBox'
	var span opentracing.Span
	switch v := spanBox.(type) {
	case opentracing.Span:
		span = v
	case context.Context:
		span = opentracing.SpanFromContext(v) // may return nil
	default:
		log.Error(pctx.TODO(), "invalid type passed to TagAnySpan", zap.Any("value", spanBox))
	}
	if span == nil {
		return nil
	}

	// tag 'span'
	for i := 0; i < len(kvs); i += 2 {
		if len(kvs) == i+1 {
			span = span.SetTag("extra", kvs[i]) // likely forgot key or value--best effort
			break
		}
		if key, ok := kvs[i].(string); ok {
			span = span.SetTag(key, kvs[i+1]) // common case -- skip printf
		} else {
			span = span.SetTag(fmt.Sprintf("%v", kvs[i]), kvs[i+1])
		}
	}
	return span
}

// AddSpanToAnyExisting checks 'ctx' for Jaeger tracing information, and if
// tracing metadata is present, it generates a new span for 'operation', marks
// it as a child of the existing span, and returns it.
func AddSpanToAnyExisting(ctx context.Context, operation string, kvs ...interface{}) (opentracing.Span, context.Context) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span := opentracing.StartSpan(operation, opentracing.ChildOf(parentSpan.Context()))
		span = TagAnySpan(span, kvs...)
		return span, opentracing.ContextWithSpan(ctx, span)
	}
	return nil, ctx
}

// FinishAnySpan calls span.Finish() if span is not nil. Pairs with
// AddSpanToAnyExisting
func FinishAnySpan(span opentracing.Span, kvs ...interface{}) {
	span = TagAnySpan(span, kvs...)
	if span != nil {
		span.Finish()
	}
}

// InstallJaegerTracerFromEnv installs a Jaeger client as the opentracing global
// tracer, relying on environment variables to configure the client
func InstallJaegerTracerFromEnv() string {
	jaegerOnce.Do(func() {
		var onUserMachine bool
		jaegerEndpoint, onUserMachine = os.LookupEnv(jaegerEndpointEnvVar)
		if !onUserMachine {
			if host, ok := os.LookupEnv("JAEGER_COLLECTOR_SERVICE_HOST"); ok {
				port := os.Getenv("JAEGER_COLLECTOR_SERVICE_PORT_JAEGER_COLLECTOR_HTTP")
				jaegerEndpoint = fmt.Sprintf("%s:%s", host, port)
			}
		}
		if jaegerEndpoint == "" {
			return // break early -- not using Jaeger
		}

		// canonicalize jaegerEndpoint as http://<hostport>/api/traces
		jaegerEndpoint = strings.TrimPrefix(jaegerEndpoint, "http://")
		jaegerEndpoint = strings.TrimSuffix(jaegerEndpoint, "/api/traces")
		jaegerEndpoint = fmt.Sprintf("http://%s/api/traces", jaegerEndpoint)
		cfg := jaegercfg.Configuration{
			ServiceName: JaegerServiceName,
			// Configure Jaeger to sample every call, but use the SpanInclusionFunc
			// addTraceIfTracingEnabled (defined below) to skip sampling every RPC
			// unless the PACH_TRACE environment variable is set
			Sampler: &jaegercfg.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:            true,
				BufferFlushInterval: 1 * time.Second,
				CollectorEndpoint:   jaegerEndpoint,
			},
		}

		// configure jaeger logger
		logger := jaeger.Logger(jaeger.NullLogger)
		if !onUserMachine {
			logger = jaeger.StdLogger
		}

		// Hack: ignore second argument (io.Closer) because the Jaeger
		// implementation of opentracing.Tracer also implements io.Closer (i.e. the
		// first and second return values from cfg.New(), here, are two interfaces
		// that wrap the same underlying type). Instead of storing the second return
		// value here, just cast the tracer to io.Closer in CloseAndReportTraces()
		// (below) and call 'Close()' on it there.
		tracer, _, err := cfg.NewTracer(jaegercfg.Logger(logger))
		if err != nil {
			log.Error(pctx.TODO(), "jaeger-collector service is deployed, but Pachyderm could not install Jaeger tracer", zap.Error(err))
			return
		}
		opentracing.SetGlobalTracer(tracer)
		log.Info(pctx.TODO(), "jaeger setup ok")
	})
	return jaegerEndpoint
}

// addTraceIfTracingEnabled is an otgrpc span inclusion func that propagates
// existing traces, but won't start any new ones
func addTraceIfTracingEnabled(
	parentSpanCtx opentracing.SpanContext,
	method string,
	req, resp interface{}) bool {
	// Always trace if PACH_TRACE is on
	if _, shortTracingOn := os.LookupEnv(ShortTraceEnvVar); shortTracingOn {
		if !IsActive() {
			log.Error(pctx.TODO(), "PACH_TRACE is set, indicating tracing is requested, but no connection to Jaeger has been established")
		}
		return true
	}

	// Otherwise, only propagate an existing trace
	if parentSpanCtx == nil {
		return false
	}
	if jaegerCtx, ok := parentSpanCtx.(jaeger.SpanContext); ok {
		return jaegerCtx.IsValid()
	}
	// Non-Jaeger context. This shouldn't happen, unless some Pachyderm user is
	// propagating e.g. Zipkin traces through the Pachyderm client. In that
	// case, we wouldn't know where to report traces anyway
	return false
}

// IsActive returns true if a connection to Jaeger has been established and a
// global tracer has been installed
func IsActive() bool {
	return opentracing.IsGlobalTracerRegistered()
}

// UnaryClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(otgrpc.SpanInclusionFunc(addTraceIfTracingEnabled)))
}

// StreamClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(otgrpc.SpanInclusionFunc(addTraceIfTracingEnabled)))
}

// UnaryServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(otgrpc.SpanInclusionFunc(addTraceIfTracingEnabled)))
}

// StreamServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(otgrpc.SpanInclusionFunc(addTraceIfTracingEnabled)))
}

// CloseAndReportTraces tries to close the global tracer, which, in the case of
// the Jaeger tracer, causes it to send any unreported traces to the collector
func CloseAndReportTraces() {
	if c, ok := opentracing.GlobalTracer().(io.Closer); ok {
		if err := c.Close(); err != nil {
			log.Error(pctx.Background("opentracing"), "problem closing global tracer", zap.Error(err))
		}
	}
}
