package obj

import (
	"context"
	"io"
)

// Copy copys an object from src at srcPath to dst at dstPath
func Copy(ctx context.Context, src, dst Client, srcPath, dstPath string) (retErr error) {
	rc, err := src.Reader(ctx, srcPath, 0, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := rc.Close(); retErr == nil {
			retErr = err
		}
	}()
	wc, err := dst.Writer(ctx, dstPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := wc.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(wc, rc)
	return err
}
