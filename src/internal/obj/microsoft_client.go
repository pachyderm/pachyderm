package obj

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
)

const (
	// Maximum block size set to 4MB.
	maxBlockSize = 4 * 1024 * 1024
	// Concurrency is the maximum concurrent block writes per writer.
	concurrency = 10
)

var (
	bufPool = grpcutil.NewBufPool(maxBlockSize)
)

type microsoftClient struct {
	container *storage.Container
}

func newMicrosoftClient(container string, accountName string, accountKey string) (*microsoftClient, error) {
	client, err := storage.NewBasicClient(accountName, accountKey)
	if err != nil {
		return nil, err
	}
	client.HTTPClient.Transport = promutil.InstrumentRoundTripper("azure_storage", client.HTTPClient.Transport)
	blobSvc := client.GetBlobService()
	return &microsoftClient{container: (&blobSvc).GetContainerReference(container)}, nil
}

// TODO: remove the writer, and respect the context.
func (c *microsoftClient) Put(ctx context.Context, name string, r io.Reader) error {
	w := newMicrosoftWriter(ctx, c, name)
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// TODO: should respect context
func (c *microsoftClient) Get(_ context.Context, name string, w io.Writer) (retErr error) {
	r, err := c.container.GetBlobReference(name).Get(nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(w, r)
	return err
}

// TODO: should respect context
func (c *microsoftClient) Delete(_ context.Context, name string) error {
	_, err := c.container.GetBlobReference(name).DeleteIfExists(nil)
	return err
}

// TODO: should respect context
func (c *microsoftClient) Walk(_ context.Context, name string, f func(name string) error) error {
	var marker string
	for {
		blobList, err := c.container.ListBlobs(storage.ListBlobsParameters{
			Prefix: name,
			Marker: marker,
		})
		if err != nil {
			return err
		}
		for _, file := range blobList.Blobs {
			if err := f(file.Name); err != nil {
				return err
			}
		}
		// NextMarker is empty when all results have been returned
		if blobList.NextMarker == "" {
			break
		}
		marker = blobList.NextMarker
	}
	return nil
}

// TODO: should respect context
func (c *microsoftClient) Exists(ctx context.Context, name string) (bool, error) {
	exists, err := c.container.GetBlobReference(name).Exists()
	tracing.TagAnySpan(ctx, "exists", exists, "err", err)
	if err != nil {
		err = c.transformError(err, name)
		return false, err
	}
	return exists, nil
}

func (c *microsoftClient) transformError(err error, name string) error {
	const minWait = 250 * time.Millisecond
	microsoftErr := &storage.AzureStorageServiceError{}
	if !errors.As(err, &microsoftErr) {
		return err
	}
	if microsoftErr.StatusCode >= 500 {
		return pacherr.WrapTransient(err, minWait)
	}
	if microsoftErr.StatusCode == 404 {
		return pacherr.NewNotExist(c.container.Name, name)
	}
	return err
}

type microsoftWriter struct {
	ctx       context.Context
	blob      *storage.Blob
	w         *grpcutil.ChunkWriteCloser
	limiter   limit.ConcurrencyLimiter
	eg        *errgroup.Group
	numBlocks int
	err       error
}

func newMicrosoftWriter(ctx context.Context, client *microsoftClient, name string) *microsoftWriter {
	eg, cancelCtx := errgroup.WithContext(ctx)
	w := &microsoftWriter{
		ctx:     cancelCtx,
		blob:    client.container.GetBlobReference(name),
		limiter: limit.New(concurrency),
		eg:      eg,
	}
	w.w = grpcutil.NewChunkWriteCloser(bufPool, w.writeBlock)
	return w
}

func (w *microsoftWriter) Write(data []byte) (retN int, retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Microsoft.Writer/Write")
	defer func() {
		tracing.FinishAnySpan(span, "bytes", retN, "err", retErr)
	}()
	if w.err != nil {
		return 0, w.err
	}
	return w.w.Write(data)
}

func (w *microsoftWriter) writeBlock(block []byte) (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Microsoft.Writer/WriteBlock")
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	blockID := blockID(w.numBlocks)
	w.numBlocks++
	w.limiter.Acquire()

	w.eg.Go(func() error {
		defer w.limiter.Release()
		defer bufPool.Put(block[:cap(block)]) //nolint:staticcheck // []byte is sufficiently pointer-like for our purposes
		if err := w.blob.PutBlock(blockID, block, nil); err != nil {
			w.err = err
			return err
		}
		return nil
	})
	return nil
}

func blockID(n int) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%011d\n", n)))
}

func (w *microsoftWriter) Close() (retErr error) {
	span, _ := tracing.AddSpanToAnyExisting(w.ctx, "/Microsoft.Writer/Close")
	defer func() {
		tracing.FinishAnySpan(span, "err", retErr)
	}()
	if err := w.w.Close(); err != nil {
		return err
	}
	if err := w.eg.Wait(); err != nil {
		return err
	}
	// Finalize the blocks.
	blocks := make([]storage.Block, w.numBlocks)
	for i := range blocks {
		blocks[i] = storage.Block{ID: blockID(i), Status: storage.BlockStatusUncommitted}
	}
	return w.blob.PutBlockList(blocks, nil)
}
