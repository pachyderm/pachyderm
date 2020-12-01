package obj

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

type memClient struct {
	objects sync.Map
}

func NewMem() Client {
	return &memClient{}
}

func (c *memClient) Writer(ctx context.Context, p string) (io.WriteCloser, error) {
	return &memWriter{
		c: c,
	}, nil
}

func (c *memClient) Reader(ctx context.Context, p string, offset, size uint64) (io.ReadCloser, error) {
	v, exists := c.objects.Load(p)
	if !exists {
		return nil, os.ErrNotExist
	}
	return ioutil.NopCloser(bytes.NewReader(v.([]byte))), nil
}

func (c *memClient) Exists(ctx context.Context, p string) bool {
	_, exists := c.objects.Load(p)
	return exists
}

func (c *memClient) Delete(ctx context.Context, p string) error {
	c.objects.Delete(p)
	return nil
}

func (c *memClient) IsIgnorable(error) bool {
	return false
}

func (c *memClient) IsNotExist(err error) bool {
	return err == os.ErrNotExist
}

func (c *memClient) IsRetryable(error) bool {
	return false
}

func (c *memClient) Walk(ctx context.Context, prefix string, cb func(p string) error) error {
	var err error
	c.objects.Range(func(k, _ interface{}) bool {
		p := k.(string)
		if !strings.HasPrefix(p, prefix) {
			return true
		}
		err = cb(k.(string))
		return err == nil
	})
	return err
}

type memWriter struct {
	c        *memClient
	path     string
	buf      bytes.Buffer
	isClosed bool
}

func (w *memWriter) Write(p []byte) (int, error) {
	if w.isClosed {
		return 0, io.ErrClosedPipe
	}
	return w.buf.Write(p)
}

func (w *memWriter) Close() error {
	if !w.isClosed {
		w.c.objects.Store(w.path, w.buf.Bytes())
		w.isClosed = true
	}
	return nil
}
