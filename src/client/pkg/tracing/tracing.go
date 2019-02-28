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
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
)

// JaegerServiceName is the service name used when the client reports traces
// to Jaeger
const JaegerServiceName = "pachd"

// jaegerOnce is used to ensure that the Jaeger tracer is only initialized once
var jaegerOnce sync.Once

var enableTracing bool

// EnableTracing causes our opentracing GRPC interceptor to attach traces to
// outgoing RPCs (once an RPC has a trace attached to it, the trace will cause
// downstream RPCs to be traced as well). This is called by pachctl (controlled
// by a flag), but not by pachd, so that the only way to trace a Pachyderm RPC
// is by setting the appropriate flag in pachctl
func EnableTracing(enabled bool) {
	enableTracing = enabled
}

// AddSpanToAnyExisting checks 'ctx' for Jaeger tracing information, and if
// tracing metadata is present, it generates a new span for 'operation', marks
// it as a child of the existing span, and returns it.
func AddSpanToAnyExisting(ctx context.Context, operation string) (opentracing.Span, context.Context) {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span := opentracing.StartSpan(operation, opentracing.ChildOf(parentSpan.Context()))
		return span, opentracing.ContextWithSpan(ctx, span)
	}
	return nil, ctx
}

// FinishAnySpan calls span.Finish() if span is not nil. Pairs with
// AddSpanToAnyExisting
func FinishAnySpan(span opentracing.Span) {
	if span != nil {
		span.Finish()
	}
}

// InstallJaegerTracerFromEnv installs a Jaeger client as then opentracing
// global tracer, relying on environment variables to configure the client. It
// returns the address used to initialize the global tracer, if any
// initialization occurred
func InstallJaegerTracerFromEnv() string {
	var jaegerEndpoint string
	jaegerOnce.Do(func() {
		var onUserMachine bool
		jaegerEndpoint, onUserMachine = os.LookupEnv("JAEGER_ENDPOINT")
		if !onUserMachine {
			if host, ok := os.LookupEnv("JAEGER_COLLECTOR_SERVICE_HOST"); ok {
				port := os.Getenv("JAEGER_COLLECTOR_SERVICE_PORT_JAEGER_COLLECTOR_HTTP")
				jaegerEndpoint = fmt.Sprintf("%s:%s", host, port)
			}
		}

		// canonicalize jaegerEndpoint as http://<hostport>/api/traces
		jaegerEndpoint = strings.TrimPrefix(jaegerEndpoint, "http://")
		jaegerEndpoint = strings.TrimSuffix(jaegerEndpoint, "/api/traces")
		jaegerEndpoint = fmt.Sprintf("http://%s/api/traces", jaegerEndpoint)
		cfg := jaegercfg.Configuration{
			// Configure Jaeger to sample every call, but use the SpanInclusionFunc
			// addTraceIfTracingEnabled (defined below) to skip sampling every RPC
			// unless EnableTracing is set
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
		// don't try to keep or use the closer, as this tracer will run for the
		// duration of the pachd binary
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
		tracer, _, err := cfg.New(JaegerServiceName, jaegercfg.Logger(logger))
		if err != nil {
			panic(fmt.Sprintf("could not install Jaeger tracer: %v", err))
		}
		opentracing.SetGlobalTracer(tracer)
	})
	return jaegerEndpoint
}

// addTraceIfTracingEnabled is an otgrpc interceptor option that propagates
// existing traces, but won't start any new ones
var addTraceIfTracingEnabled otgrpc.SpanInclusionFunc = func(
	parentSpanCtx opentracing.SpanContext,
	method string,
	req, resp interface{}) bool {
	// Always trace if enableTracing is on
	if enableTracing {
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
	// propagating e.g.  Zipkin traces through the Pachyderm client. In that
	// case, we wouldn't know where to report traces anyway
	return false
}

// UnaryClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(addTraceIfTracingEnabled))
}

// StreamClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(addTraceIfTracingEnabled))
}

// UnaryServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(addTraceIfTracingEnabled))
}

// StreamServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer(),
		otgrpc.IncludingSpans(addTraceIfTracingEnabled))
}

// CloseAndReportTraces tries to close the global tracer, which, in the case of
// the Jaeger tracer, causes it to send any unreported traces to the collector
func CloseAndReportTraces() {
	if c, ok := opentracing.GlobalTracer().(io.Closer); ok {
		c.Close()
	}
}
