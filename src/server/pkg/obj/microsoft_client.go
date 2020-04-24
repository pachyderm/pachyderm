package obj

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/storage"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
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
	blobSvc := client.GetBlobService()
	return &microsoftClient{container: (&blobSvc).GetContainerReference(container)}, nil
}

func (c *microsoftClient) Writer(ctx context.Context, name string) (io.WriteCloser, error) {
	return newMicrosoftWriter(ctx, c, name), nil
}

func (c *microsoftClient) Reader(ctx context.Context, name string, offset uint64, size uint64) (io.ReadCloser, error) {
	blobRange := blobRange(offset, size)
	if blobRange == nil {
		reader, err := c.container.GetBlobReference(name).Get(nil)
		if err != nil {
			return nil, err
		}
		return newCheckedReadCloser(size, reader), nil
	}

	reader, err := c.container.GetBlobReference(name).GetRange(&storage.GetBlobRangeOptions{Range: blobRange})
	if err != nil {
		return nil, err
	}
	return newCheckedReadCloser(size, reader), nil
}

func blobRange(offset, size uint64) *storage.BlobRange {
	if offset == 0 && size == 0 {
		return nil
	} else if size == 0 {
		return &storage.BlobRange{Start: offset}
	}
	return &storage.BlobRange{Start: offset, End: offset + size - 1}
}

func (c *microsoftClient) Delete(_ context.Context, name string) error {
	_, err := c.container.GetBlobReference(name).DeleteIfExists(nil)
	return err
}

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

func (c *microsoftClient) Exists(ctx context.Context, name string) bool {
	exists, err := c.container.GetBlobReference(name).Exists()
	tracing.TagAnySpan(ctx, "exists", exists, "err", err)
	return exists
}

func (c *microsoftClient) IsRetryable(err error) (ret bool) {
	microsoftErr, ok := err.(storage.AzureStorageServiceError)
	if !ok {
		return false
	}
	return microsoftErr.StatusCode >= 500
}

func (c *microsoftClient) IsNotExist(err error) bool {
	microsoftErr, ok := err.(storage.AzureStorageServiceError)
	if !ok {
		return false
	}
	return microsoftErr.StatusCode == 404
}

func (c *microsoftClient) IsIgnorable(err error) bool {
	return false
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

	//lint:ignore SA6002 []byte is sufficiently pointer-like for our purposes
	w.eg.Go(func() error {
		defer w.limiter.Release()
		defer bufPool.Put(block[:cap(block)])
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
