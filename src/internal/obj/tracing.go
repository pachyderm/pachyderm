package obj

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	objectOperationMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage",
		Name:      "operation_count_total",
		Help:      "Number of object storage operations, by storage type and operation name",
	}, []string{"provider", "op"})

	objectBytesReadMetrics = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage",
		Name:      "read_bytes_total",
		Help:      "Number of bytes read from object storage, by storage type",
	}, []string{"provider"})

	objectBytesWrittenMetrics = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "pfs_object_storage",
		Name:      "written_bytes_total",
		Help:      "Number of bytes written to object storage, by storage type",
	}, []string{"provider"})
)

func prettyProvider(provider string) string {
	switch provider {
	case "s3", Amazon:
		return "Amazon"
	case "gcs", "gs", Google:
		return "Google"
	case "as", "wasb", Microsoft:
		return "Microsoft"
	case "local", Local:
		return "Local"
	case Minio:
		return "Minio"
	}
	if provider == "" {
		return "Unknown"
	}
	return provider
}

// TracingObjClient wraps the given object client 'c', adding tracing and monitoring to all calls
// made by the returned interface.
func TracingObjClient(provider string, c Client) Client {
	return &tracingObjClient{c, prettyProvider(provider)}
}

var _ Client = &tracingObjClient{}

type tracingObjClient struct {
	Client
	provider string
}

// Writer implements the corresponding method in the Client interface
func (o *tracingObjClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	objectOperationMetric.WithLabelValues(o.provider, "put").Inc()
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Put", "name", name)
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	return o.Client.Put(ctx, name, &promutil.CountingReader{
		Reader: r,
		// The bytes are written to storage after being read from this reader.  Thus,
		// they're "bytes written", not "bytes read".
		Counter: objectBytesWrittenMetrics.WithLabelValues(o.provider),
	})
}

// Get implements the corresponding method in the Client interface
func (o *tracingObjClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	objectOperationMetric.WithLabelValues(o.provider, "get").Inc()
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Get", "name", name)
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	return o.Client.Get(ctx, name, &promutil.CountingWriter{
		Writer: w,
		// The bytes are read from storage, and then written to this writer.  Thus, they're
		// "bytes read", not "bytes written".
		Counter: objectBytesReadMetrics.WithLabelValues(o.provider),
	})
}

// Delete implements the corresponding method in the Client interface
func (o *tracingObjClient) Delete(ctx context.Context, name string) (retErr error) {
	objectOperationMetric.WithLabelValues(o.provider, "delete").Inc()
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Delete",
		"name", name)
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	return o.Client.Delete(ctx, name)
}

// Walk implements the corresponding method in the Client interface
func (o *tracingObjClient) Walk(ctx context.Context, prefix string, fn func(name string) error) (retErr error) {
	objectOperationMetric.WithLabelValues(o.provider, "walk").Inc()
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Walk",
		"prefix", prefix)
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	return o.Client.Walk(ctx, prefix, fn)
}

// Exists implements the corresponding method in the Client interface
func (o *tracingObjClient) Exists(ctx context.Context, name string) (retVal bool, retErr error) {
	objectOperationMetric.WithLabelValues(o.provider, "exists").Inc()
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Exists",
		"name", name)
	defer func() {
		tracing.FinishAnySpan(span, "exists", retVal)
	}()
	defer tracing.FinishAnySpan(span)
	return o.Client.Exists(ctx, name)
}
