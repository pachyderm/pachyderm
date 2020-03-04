package obj

import (
	"context"
	"io"
	"math/rand"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

var (
	monkeyTest bool
	enabled    bool
	localRand  *rand.Rand
	failProb   float64
	errMsg     = errors.Errorf("object storage slipped on a banana")
)

// InitMonkeyTest sets up this package for monkey testing.
// Object storage clients will be wrapped with a client that
// sporadically fails requests.
func InitMonkeyTest(seed int64) {
	monkeyTest = true
	localRand = rand.New(rand.NewSource(seed))
	failProb = 0.05
}

// EnableMonkeyTest enables sporadic request failures.
func EnableMonkeyTest() {
	enabled = true
}

// DisableMonkeyTest disables sporadic request failures.
func DisableMonkeyTest() {
	enabled = false
}

// IsMonkeyError checks if an error was caused by a monkey client.
func IsMonkeyError(err error) bool {
	return strings.Contains(err.Error(), errMsg.Error())
}

type monkeyReadWriteCloser struct {
	rc io.ReadCloser
	wc io.WriteCloser
}

// Read wraps the read operation.
func (rwc *monkeyReadWriteCloser) Read(data []byte) (int, error) {
	if enabled && localRand.Float64() < failProb {
		return 0, errMsg
	}
	return rwc.rc.Read(data)
}

// Write wraps the write operation.
func (rwc *monkeyReadWriteCloser) Write(data []byte) (int, error) {
	if enabled && localRand.Float64() < failProb {
		return 0, errMsg
	}
	return rwc.wc.Write(data)
}

// Close wraps the close operation.
func (rwc *monkeyReadWriteCloser) Close() error {
	if enabled && localRand.Float64() < failProb {
		return errMsg
	}
	if rwc.wc != nil {
		return rwc.wc.Close()
	}
	return rwc.rc.Close()
}

type monkeyClient struct {
	c Client
}

// Reader wraps the reader operation.
func (c *monkeyClient) Reader(ctx context.Context, path string, offset uint64, size uint64) (io.ReadCloser, error) {
	if enabled && localRand.Float64() < failProb {
		return nil, errMsg
	}
	rc, err := c.c.Reader(ctx, path, offset, size)
	if err != nil {
		return nil, err
	}
	return &monkeyReadWriteCloser{rc: rc}, nil
}

// Writer wraps the writer operation.
func (c *monkeyClient) Writer(ctx context.Context, path string) (io.WriteCloser, error) {
	if enabled && localRand.Float64() < failProb {
		return nil, errMsg
	}
	wc, err := c.c.Writer(ctx, path)
	if err != nil {
		return nil, err
	}
	return &monkeyReadWriteCloser{wc: wc}, nil
}

// Delete wraps the delete operation.
func (c *monkeyClient) Delete(ctx context.Context, path string) error {
	if enabled && localRand.Float64() < failProb {
		return errMsg
	}
	return c.c.Delete(ctx, path)
}

// Walk wraps the walk operation.
func (c *monkeyClient) Walk(ctx context.Context, dir string, walkFn func(name string) error) error {
	if enabled && localRand.Float64() < failProb {
		return errMsg
	}
	return c.c.Walk(ctx, dir, walkFn)
}

// Exists wraps the existance check.
func (c *monkeyClient) Exists(ctx context.Context, path string) bool {
	return c.c.Exists(ctx, path)
}

// IsRetryable wraps the is retryable check.
func (c *monkeyClient) IsRetryable(err error) bool {
	return c.c.IsRetryable(err)
}

// IsNotExist wraps the does not exist check.
func (c *monkeyClient) IsNotExist(err error) bool {
	return c.c.IsNotExist(err)
}

// IsIgnorable wraps the is ignorable check.
func (c *monkeyClient) IsIgnorable(err error) bool {
	return c.c.IsIgnorable(err)
}
