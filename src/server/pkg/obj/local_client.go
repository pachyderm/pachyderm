package obj

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func NewLocalClient(root string) (Client, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	return &localClient{root}, nil
}

type localClient struct {
	root string
}

func (c *localClient) Writer(path string) (io.WriteCloser, error) {
	fullPath := filepath.Join(c.root, path)

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

func (c *localClient) Reader(path string, offset uint64, size uint64) (io.ReadCloser, error) {
	file, err := os.Open(filepath.Join(c.root, path))
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (c *localClient) Delete(path string) error {
	return os.Remove(filepath.Join(c.root, path))
}

func (c *localClient) Walk(dir string, walkFn func(name string) error) error {
	return filepath.Walk(filepath.Join(c.root, dir), func(path string, _ os.FileInfo, err error) error {
		relPath, _ := filepath.Rel(c.root, path)
		return walkFn(relPath)
	})
}

func (c *localClient) Exists(path string) bool {
	_, err := os.Stat(filepath.Join(c.root, path))
	return err == nil
}

func (c *localClient) IsRetryable(err error) bool {
	return false
}

func (c *localClient) IsNotExist(err error) bool {
	return strings.Contains(err.Error(), "no such file or directory")
}

func (c *localClient) IsIgnorable(err error) bool {
	return false
}
