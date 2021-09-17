package obj

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
)

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	return miscutil.WithPipe(func(w io.Writer) error {
		return src.Get(ctx, srcPath, w)
	}, func(r io.Reader) error {
		return dst.Put(ctx, dstPath, r)
	})
}
