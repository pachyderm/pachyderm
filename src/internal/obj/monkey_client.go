package obj

import (
	"context"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
)

var (
	monkeyTest bool
	enabled    bool
	localRand  *rand.Rand
	failProb   float64
	errMsg     = pacherr.WrapTransient(errors.Errorf("object storage slipped on a banana"), 200*time.Millisecond)
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

type monkeyClient struct {
	c Client
}

// Get wraps the get operation.
func (c *monkeyClient) Get(ctx context.Context, path string, w io.Writer) error {
	if enabled && localRand.Float64() < failProb {
		return errMsg
	}
	return c.c.Get(ctx, path, w)
}

// Put wraps the put operation.
func (c *monkeyClient) Put(ctx context.Context, path string, r io.Reader) error {
	if enabled && localRand.Float64() < failProb {
		return errMsg
	}
	return c.c.Put(ctx, path, r)
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
func (c *monkeyClient) Exists(ctx context.Context, path string) (bool, error) {
	if enabled && localRand.Float64() < failProb {
		return false, errMsg
	}
	return c.c.Exists(ctx, path)
}
