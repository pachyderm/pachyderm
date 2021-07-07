package obj

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	ghttp "google.golang.org/api/transport/http"
)

type googleClient struct {
	bucketName string
	bucket     *storage.BucketHandle
}

func newGoogleClient(bucket string, opts []option.ClientOption) (*googleClient, error) {
	ctx := context.Background()
	opts = append(opts, option.WithScopes(storage.ScopeFullControl))

	// We have to build the transport and supply it to NewClient in order to instrument the
	// roundtripper.  If we pass the instrumented client directly to NewClient with
	// WithHTTPClient, the client won't have any credentials and won't be able to make requests.
	tr, err := ghttp.NewTransport(ctx, promutil.InstrumentRoundTripper("google_cloud_storage", http.DefaultTransport), opts...)
	if err != nil {
		return nil, fmt.Errorf("init google transport: %w", err)
	}
	opts = append(opts, option.WithHTTPClient(&http.Client{
		Transport: tr,
	}))
	client, err := storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	return &googleClient{bucketName: bucket, bucket: client.Bucket(bucket)}, nil
}

func (c *googleClient) Exists(ctx context.Context, name string) (bool, error) {
	_, err := c.bucket.Object(name).Attrs(ctx)
	if err != nil {
		err = c.transformError(err, name)
		if pacherr.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	tracing.TagAnySpan(ctx, "err", err)
	return true, nil
}

func (c *googleClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	ctx, cf := context.WithCancel(ctx)
	defer cf() // this aborts the write if the writer is not already closed
	wc := c.bucket.Object(name).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return err
	}
	return wc.Close()
}

func (c *googleClient) Walk(ctx context.Context, name string, fn func(name string) error) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	objectIter := c.bucket.Objects(ctx, &storage.Query{Prefix: name})
	for {
		objectAttrs, err := objectIter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return err
		}
		if err := fn(objectAttrs.Name); err != nil {
			return err
		}
	}
	return nil
}

func (c *googleClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	reader, err := c.bucket.Object(name).NewReader(ctx)
	defer func() {
		if err := reader.Close(); retErr == nil {
			retErr = err
		}
	}()
	if err != nil {
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

func (c *googleClient) Delete(ctx context.Context, name string) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	return c.bucket.Object(name).Delete(ctx)
}

func (c *googleClient) transformError(err error, objectPath string) error {
	const minWait = 250 * time.Millisecond
	if err == nil {
		return nil
	}
	if errors.Is(err, storage.ErrObjectNotExist) {
		return pacherr.NewNotExist(c.bucketName, objectPath)
	}
	// https://github.com/pachyderm/pachyderm/v2/issues/912
	if strings.Contains(err.Error(), "ParseError") {
		return err
	}

	var googleErr googleapi.Error
	if !errors.As(err, &googleErr) {
		return err
	}
	switch googleErr.Code {
	case 429:
		return pacherr.WrapTransient(err, minWait)

	// https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
	case 500, 502, 503, 504:
		return pacherr.WrapTransient(err, minWait)

	}
	return err
}
