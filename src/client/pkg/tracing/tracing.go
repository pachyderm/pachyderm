package tracing

import (
	"fmt"
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

// If true, don't log from tracing/tracing.go
// TODO(msteffen) >>> remove this var
var DisableLogs bool

// InstallJaegerTracerFromEnv installs a Jaeger client as then opentracing
// global tracer, relying on environment variables to configure the client
func InstallJaegerTracerFromEnv() {
	jaegerOnce.Do(func() {
		jaegerEndpoint, ok := os.LookupEnv("JAEGER_ENDPOINT")
		onUserMachine := ok
		if !ok {
			if host, ok := os.LookupEnv("JAEGER_COLLECTOR_SERVICE_HOST"); ok {
				port := os.Getenv("JAEGER_COLLECTOR_SERVICE_PORT_JAEGER_COLLECTOR_HTTP")
				jaegerEndpoint = fmt.Sprintf("%s:%s", host, port)
			}
		}

		// canonicalize jaegerEndpoint as http://<hostport>/api/traces
		jaegerEndpoint = strings.TrimPrefix(jaegerEndpoint, "http://")
		jaegerEndpoint = strings.TrimSuffix(jaegerEndpoint, "/api/traces")
		jaegerEndpoint = fmt.Sprintf("http://%s/api/traces", jaegerEndpoint)
		if !onUserMachine {
			fmt.Printf("\n>>> CollectorEndpoint: %s\n", jaegerEndpoint)
		}
		if !DisableLogs {
			fmt.Printf("\n>>> CollectorEndpoint: %s\n", jaegerEndpoint)
		}
		cfg := jaegercfg.Configuration{
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
		// TODO(msteffen) respect log verbosity, particularly on the client
		// TODO(msteffen) I wrote this in a ridiculous way so that we didn't get
		// any superfluous output from 'pachctl version --client-only'. Come up
		// with a better way of doing that.
		logger := jaeger.Logger(jaeger.NullLogger)
		if !onUserMachine {
			logger = jaeger.StdLogger
		}
		if !DisableLogs {
			logger = jaeger.StdLogger
		}
		tracer, _, err := cfg.New(JaegerServiceName, jaegercfg.Logger(logger))
		if err != nil {
			panic(fmt.Sprintf("could not install Jaeger tracer: %v", err))
		}
		opentracing.SetGlobalTracer(tracer)
	})
}

// UnaryClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())
}

// StreamClientInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return otgrpc.OpenTracingStreamClientInterceptor(opentracing.GlobalTracer())
}

// UnaryServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())
}

// StreamServerInterceptor returns a GRPC interceptor for non-streaming GRPC RPCs
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return otgrpc.OpenTracingStreamServerInterceptor(opentracing.GlobalTracer())
}
