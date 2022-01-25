package obj

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
)

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	return miscutil.WithPipe(func(w io.Writer) error {
		return errors.EnsureStack(src.Get(ctx, srcPath, w))
	}, func(r io.Reader) error {
		return errors.EnsureStack(dst.Put(ctx, dstPath, r))
	})
}

type testURL struct {
	Client
}

// WrapWithTestURL marks client as a test client and will prepend test- to the url.
// the consturctors in this package know how to parse test urls, and assume default credentials.
func WrapWithTestURL(c Client) Client {
	return testURL{c}
}

func (c testURL) BucketURL() ObjectStoreURL {
	u := c.Client.BucketURL()
	u.Scheme = "test-" + u.Scheme
	return u
}
