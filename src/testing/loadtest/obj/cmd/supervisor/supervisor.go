package main

import (
	"bytes"
	"context"
	"flag"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

const (
	// MB is a megabyte
	MB     = 1024 * 1024
	prefix = "/pach"
)

var (
	// Flags
	// Number of objects.
	// (bryce) change to int64 later.
	numObjects int
	// Size of the objects.
	// (bryce) change to int64 later.
	objectSize int
	// Maximum concurrent writes.
	concurrency int
)

func init() {
	flag.IntVar(&numObjects, "num-objects", 10, "number of objects")
	flag.IntVar(&objectSize, "object-size", MB, "size of the objects")
	flag.IntVar(&concurrency, "concurrency", 5, "maximum concurrent writes")
}

// PrintFlags just prints the flag values, set above, to stdout. Useful for
// comparing benchmark runs
func PrintFlags(ctx context.Context) {
	log.Debug(ctx, "flags", zap.Int("num-objects", numObjects), zap.Int("object-size", objectSize), zap.Int("concurrency", concurrency))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandSeq generates a random sequence of data (n is number of bytes)
func RandSeq(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}

func main() {
	log.InitPachctlLogger()
	ctx := pctx.Background("")

	flag.Parse()
	PrintFlags(ctx)
	// Setup client.
	storageBackend, ok := os.LookupEnv(obj.StorageBackendEnvVar)
	if !ok {
		log.Error(ctx, "storage backend environment variable not found")
		os.Exit(1)
	}
	b, err := obj.NewBucket(ctx, storageBackend, prefix)
	if err != nil {
		log.Error(ctx, "Error creating client", zap.Error(err))
		os.Exit(1)
	}
	// (bryce) It might make sense to clean up the bucket here before running the tests.
	// Run basic test.
	start := time.Now()
	log.Info(ctx, "Basic test started.")
	if err := basicTest(b); err != nil {
		log.Error(ctx, "Basic test error", zap.Error(err))
		os.Exit(1)
	}
	log.Info(ctx, "Basic test completed.", zap.Duration("time", time.Since(start)))
	// Run load test.
	start = time.Now()
	log.Info(ctx, "Load test started.")
	if err := loadTest(b); err != nil {
		log.Error(ctx, "Load test error", zap.Error(err))
		os.Exit(1)
	}
	log.Info(ctx, "Load test completed.", zap.Duration("time", time.Since(start)))
}

func basicTest(b *obj.Bucket) error {
	ctx := context.Background()
	name := "0"
	// Confirm that an existence check and deletion for a non-existent object works correctly.
	if exists, err := b.Exists(ctx, name); err != nil {
		return errors.EnsureStack(err)
	} else if !exists {
		return errors.Errorf("existence check returns true when the object should not exist")
	}
	if err := b.Delete(ctx, name); err != nil {
		return errors.Wrap(err, "deletion errored on non-existent object")
	}
	if err := walk(ctx, b, 0, nil); err != nil {
		return err
	}
	numObjects := 5
	basicObjectSize := 1024
	data := RandSeq(basicObjectSize)
	// Write then read objects.
	for i := 0; i < numObjects; i++ {
		name := strconv.Itoa(i)
		if err := writeObject(ctx, b, name, data); err != nil {
			return err
		}
		if err := readTest(ctx, b, name, data); err != nil {
			return err
		}
	}
	// Confirm range reads work correctly
	offset, size := basicObjectSize/2, 0
	if err := readTest(ctx, b, name, data[offset:]); err != nil {
		return err
	}
	offset, size = basicObjectSize/2, basicObjectSize/4
	if err := readTest(ctx, b, name, data[offset:offset+size]); err != nil {
		return err
	}
	// Walk the objects and for each check the existence and delete it.
	if err := walk(ctx, b, 5, func(name string) error {
		if exists, err := b.Exists(ctx, name); err != nil {
			return errors.EnsureStack(err)
		} else if !exists {
			return errors.Errorf("existence check returns false when the object should exist")
		}
		return errors.EnsureStack(b.Delete(ctx, name))
	}); err != nil {
		return err
	}
	// Test writing and reading a size zero object.
	data = []byte{}
	if err := writeObject(ctx, b, "zero", data); err != nil {
		return err
	}
	if err := readTest(ctx, b, "zero", data); err != nil {
		return err
	}
	if err := b.Delete(ctx, "zero"); err != nil {
		return errors.EnsureStack(err)
	}
	// Confirm that no objects exist after deletion.
	return walk(ctx, b, 0, nil)
}

func walk(ctx context.Context, b *obj.Bucket, expected int, f func(string) error) error {
	objCount := 0
	iter := b.List(nil)
	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.Wrap(err, "walk bucket")
		}
		objCount++
		if f != nil {
			return f(obj.Key)
		}
	}
	if objCount != expected {
		return errors.Errorf("walk should have returned %v objects, not %v", expected, objCount)
	}
	return nil
}

func writeObject(ctx context.Context, b *obj.Bucket, name string, data []byte) (retErr error) {
	w, err := b.NewWriter(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "write object")
	}
	defer func() {
		retErr = errors.Join(retErr, errors.Wrap(w.Close(), "close writer"))
	}()
	if _, err = w.Write(data); err != nil {
		return errors.Wrapf(err, "write data")
	}
	return nil
}

func readObject(ctx context.Context, b *obj.Bucket, name string, p []byte) (retErr error) {
	r, err := b.NewReader(ctx, name, nil)
	if err != nil {
		return errors.Wrapf(err, "read object")
	}
	content, err := io.ReadAll(r)
	if err != nil {
		return errors.Wrapf(err, "read data")
	}
	defer func() {
		retErr = errors.Join(retErr, errors.Wrap(r.Close(), "close reader"))
	}()
	copy(p, content)
	return nil
}

func readTest(ctx context.Context, b *obj.Bucket, name string, expected []byte) error {
	buf := make([]byte, len(expected))
	if err := readObject(ctx, b, name, buf); err != nil {
		return err
	}
	if !bytes.Equal(expected, buf) {
		return errors.Errorf("range read for object %v incorrect (offset: %v, size: %v)", name)
	}
	return nil
}

func loadTest(b *obj.Bucket) error {
	limiter := limit.New(concurrency)
	eg, ctx := errgroup.WithContext(context.Background())
	data := RandSeq(objectSize)
	bufPool := grpcutil.NewBufPool(objectSize)
	for i := 0; i < numObjects; i++ {
		i := i
		limiter.Acquire()
		eg.Go(func() error {
			defer limiter.Release()
			name := strconv.Itoa(i)
			if err := writeObject(ctx, b, name, data); err != nil {
				return err
			}
			buf := bufPool.GetBuffer()
			defer bufPool.PutBuffer(buf)
			if err := readObject(ctx, b, name, buf); err != nil {
				return err
			}
			if !bytes.Equal(data, buf) {
				return errors.Errorf("data written does not equal data read for object %v", i)
			}
			return errors.EnsureStack(b.Delete(ctx, name))
		})
	}
	return errors.EnsureStack(eg.Wait())
}
