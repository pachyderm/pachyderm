package obj

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"gocloud.dev/blob"
)

type keyIterator struct {
	it blob.ListIterator
}

// NewKeyIterator iterates over the keys in a blob.Bucket.
func NewKeyIterator(b *blob.Bucket, prefix string) stream.Iterator[string] {
	return &keyIterator{it: *b.List(&blob.ListOptions{Prefix: prefix})}
}

func (it *keyIterator) Next(ctx context.Context, dst *string) error {
	lo, err := it.it.Next(ctx)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return stream.EOS()
		}
		return errors.EnsureStack(err)
	}
	*dst = lo.Key
	return nil
}
