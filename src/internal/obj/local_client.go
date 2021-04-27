package obj

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
)

// NewLocalClient returns a Client that stores data on the local file system
func NewLocalClient(root string) (c Client, err error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, errors.EnsureStack(err)
	}
	client := &localClient{filepath.Clean(root)}
	if monkeyTest {
		return &monkeyClient{client}, nil
	}
	return newUniformClient(client), nil
}

type localClient struct {
	root string
}

func (c *localClient) normPath(path string) string {
	path = filepath.Clean(path)
	if !filepath.IsAbs(path) {
		return filepath.Join(c.root, path)
	}
	return path
}

func (c *localClient) Put(ctx context.Context, path string, r io.Reader) (retErr error) {
	defer func() { retErr = c.transformError(retErr, path) }()
	fullPath := c.normPath(path)

	// Create the directory since it may not exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); retErr == nil {
			retErr = err
		}
	}()
	if deadline, ok := ctx.Deadline(); ok {
		if err := file.SetWriteDeadline(deadline); err != nil {
			return err
		}
	}
	if _, err := io.Copy(file, r); err != nil {
		return err
	}
	return nil
}

func (c *localClient) Get(ctx context.Context, path string, w io.Writer) (retErr error) {
	defer func() { retErr = c.transformError(retErr, path) }()
	file, err := os.Open(c.normPath(path))
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); retErr == nil {
			retErr = err
		}
	}()
	if deadline, ok := ctx.Deadline(); ok {
		if err := file.SetReadDeadline(deadline); err != nil {
			return err
		}
	}
	_, err = io.Copy(w, file)
	return err
}

func (c *localClient) Delete(_ context.Context, path string) (retErr error) {
	defer func() { retErr = c.transformError(retErr, path) }()
	return os.Remove(c.normPath(path))
}

func (c *localClient) Walk(_ context.Context, dir string, walkFn func(name string) error) error {
	dir = c.normPath(dir)
	fi, _ := os.Stat(dir)
	prefix := ""
	if fi == nil || !fi.IsDir() {
		dir, prefix = filepath.Split(dir)
	}
	err := filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if fileInfo.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(c.root, path)
		if !strings.HasPrefix(filepath.Base(relPath), prefix) {
			return nil
		}
		return walkFn(relPath)
	})
	return err
}

func (c *localClient) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(c.normPath(path))
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return false, err
	}
	tracing.TagAnySpan(ctx, "err", err)
	return true, nil
}

func (c *localClient) transformError(err error, name string) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) || strings.HasSuffix(err.Error(), ": no such file or directory") {
		return pacherr.NewNotExist(c.root, name)
	}
	return err
}
