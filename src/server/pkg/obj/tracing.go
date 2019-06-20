package obj

import (
	"context"
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
)

// TracingObjClient wraps the given object client 'c', adding tracing to all calls made
// by the returned interface
func TracingObjClient(provider string, c Client) Client {
	return &tracingObjClient{c, provider}
}

type tracingObjClient struct {
	Client
	provider string
}

// Writer implements the corresponding method in the Client interface
func (o *tracingObjClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Writer", "name", name)
	if span != nil {
		defer span.Finish()
	}
	return o.Client.Writer(ctx, name)
}

// Reader implements the corresponding method in the Client interface
func (o *tracingObjClient) Reader(ctx context.Context, name string, offset uint64, size uint64) (io.ReadCloser, error) {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Reader",
		"name", name,
		"offset", fmt.Sprintf("%d", offset),
		"size", fmt.Sprintf("%d", size))
	defer tracing.FinishAnySpan(span)
	return o.Client.Reader(ctx, name, offset, size)
}

// Delete implements the corresponding method in the Client interface
func (o *tracingObjClient) Delete(ctx context.Context, name string) error {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Delete",
		"name", name)
	defer tracing.FinishAnySpan(span)
	return o.Client.Delete(ctx, name)
}

// Walk implements the corresponding method in the Client interface
func (o *tracingObjClient) Walk(ctx context.Context, prefix string, fn func(name string) error) error {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Walk",
		"prefix", prefix)
	defer tracing.FinishAnySpan(span)
	return o.Client.Walk(ctx, prefix, fn)
}

// Exists implements the corresponding method in the Client interface
func (o *tracingObjClient) Exists(ctx context.Context, name string) bool {
	span, ctx := tracing.AddSpanToAnyExisting(ctx, "/"+o.provider+"/Exists",
		"name", name)
	defer tracing.FinishAnySpan(span)
	return o.Client.Exists(ctx, name)
}
