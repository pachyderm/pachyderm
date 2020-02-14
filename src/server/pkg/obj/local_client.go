package obj

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
)

// NewLocalClient returns a Client that stores data on the local file system
func NewLocalClient(root string) (Client, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	c := &localClient{filepath.Clean(root)}
	if monkeyTest {
		return &monkeyClient{c}, nil
	}
	return c, nil
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

func (c *localClient) Writer(_ context.Context, path string) (io.WriteCloser, error) {
	fullPath := c.normPath(path)

	// Create the directory since it may not exist
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *localClient) Reader(_ context.Context, path string, offset uint64, size uint64) (io.ReadCloser, error) {
	file, err := os.Open(c.normPath(path))
	if err != nil {
		return nil, err
	}
	if _, err := file.Seek(int64(offset), 0); err != nil {
		return nil, err
	}

	if size == 0 {
		if _, err := file.Seek(int64(offset), 0); err != nil {
			return nil, err
		}
		return file, nil
	}
	return newSectionReadCloser(file, offset, size), nil
}

func (c *localClient) Delete(_ context.Context, path string) error {
	return os.Remove(c.normPath(path))
}

func (c *localClient) Walk(_ context.Context, dir string, walkFn func(name string) error) error {
	dir = c.normPath(dir)
	fi, _ := os.Stat(dir)
	prefix := ""
	if fi == nil || !fi.IsDir() {
		dir, prefix = filepath.Split(dir)
	}
	return filepath.Walk(dir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			if c.IsNotExist(err) {
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
}

func (c *localClient) Exists(ctx context.Context, path string) bool {
	_, err := os.Stat(c.normPath(path))
	tracing.TagAnySpan(ctx, "err", err)
	return err == nil
}

func (c *localClient) IsRetryable(err error) bool {
	return false
}

func (c *localClient) IsNotExist(err error) bool {
	return strings.Contains(err.Error(), "no such file or directory") ||
		strings.Contains(err.Error(), "cannot find the file specified")
}

func (c *localClient) IsIgnorable(err error) bool {
	return false
}

type sectionReadCloser struct {
	*io.SectionReader
	f *os.File
}

func newSectionReadCloser(f *os.File, offset uint64, size uint64) *sectionReadCloser {
	return &sectionReadCloser{
		SectionReader: io.NewSectionReader(f, int64(offset), int64(size)),
		f:             f,
	}
}

func (s *sectionReadCloser) Close() error {
	return s.f.Close()
}
