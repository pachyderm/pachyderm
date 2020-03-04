package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"golang.org/x/sync/errgroup"
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
func PrintFlags() {
	fmt.Printf("num-objects: %v\n", numObjects)
	fmt.Printf("object-size: %v\n", objectSize)
	fmt.Printf("concurrency: %v\n", concurrency)
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
	flag.Parse()
	PrintFlags()
	// Setup client.
	c, err := obj.NewClientFromEnv(prefix)
	if err != nil {
		log.Fatalf("Error creating client (%v)", err)
	}
	// (bryce) It might make sense to clean up the bucket here before running the tests.
	// Run basic test.
	start := time.Now()
	fmt.Printf("Basic test started.\n")
	if err := basicTest(c); err != nil {
		log.Fatalf("Basic test error: %v", err)
	}
	fmt.Printf("Basic test completed. Total time: %.3f\n", time.Since(start).Seconds())
	// Run load test.
	start = time.Now()
	fmt.Printf("Load test started.\n")
	if err := loadTest(c); err != nil {
		log.Fatalf("Load test error: %v", err)
	}
	fmt.Printf("Load test completed. Total time: %.3f\n", time.Since(start).Seconds())
}

func basicTest(c obj.Client) error {
	ctx := context.Background()
	name := "0"
	// Confirm that an existence check and deletion for a non-existent object works correctly.
	if c.Exists(ctx, name) {
		return errors.Errorf("existence check returns true when the object should not exist")
	}
	if err := c.Delete(ctx, name); err != nil {
		return errors.Wrap(err, "deletion errored on non-existent object")
	}
	if err := walk(ctx, c, 0, nil); err != nil {
		return err
	}
	numObjects := 5
	basicObjectSize := 1024
	data := RandSeq(basicObjectSize)
	// Write then read objects.
	for i := 0; i < numObjects; i++ {
		name := strconv.Itoa(i)
		if err := writeObject(ctx, c, name, data); err != nil {
			return err
		}
		if err := readTest(ctx, c, name, 0, 0, data); err != nil {
			return err
		}
	}
	// Confirm range reads work correctly
	offset, size := basicObjectSize/2, 0
	if err := readTest(ctx, c, name, offset, size, data[offset:]); err != nil {
		return err
	}
	offset, size = basicObjectSize/2, basicObjectSize/4
	if err := readTest(ctx, c, name, offset, size, data[offset:offset+size]); err != nil {
		return err
	}
	// Walk the objects and for each check the existence and delete it.
	if err := walk(ctx, c, 5, func(name string) error {
		if !c.Exists(ctx, name) {
			return errors.Errorf("existence check returns false when the object should exist")
		}
		return c.Delete(ctx, name)
	}); err != nil {
		return err
	}
	// Test writing and reading a size zero object.
	data = []byte{}
	if err := writeObject(ctx, c, "zero", data); err != nil {
		return err
	}
	if err := readTest(ctx, c, "zero", 0, 0, data); err != nil {
		return err
	}
	if err := c.Delete(ctx, "zero"); err != nil {
		return err
	}
	// Confirm that no objects exist after deletion.
	return walk(ctx, c, 0, nil)
}

func walk(ctx context.Context, c obj.Client, expected int, f func(string) error) error {
	objCount := 0
	if err := c.Walk(ctx, "", func(name string) error {
		objCount++
		if f != nil {
			return f(name)
		}
		return nil
	}); err != nil {
		return err
	}
	if objCount != expected {
		return errors.Errorf("walk should have returned %v objects, not %v", expected, objCount)
	}
	return nil
}

func writeObject(ctx context.Context, c obj.Client, name string, data []byte) (retErr error) {
	w, err := c.Writer(ctx, name)
	if err != nil {
		return err
	}
	defer func() {
		if err := w.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	r := bytes.NewReader(data)
	_, err = io.Copy(w, r)
	return err
}

func readObject(ctx context.Context, c obj.Client, name string, offset, size int, buf []byte) (retErr error) {
	r, err := c.Reader(ctx, name, uint64(offset), uint64(size))
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	_, err = io.ReadFull(r, buf)
	return err
}

func readTest(ctx context.Context, c obj.Client, name string, offset, size int, expected []byte) error {
	buf := make([]byte, len(expected))
	if err := readObject(ctx, c, name, offset, size, buf); err != nil {
		return err
	}
	if !bytes.Equal(expected, buf) {
		return errors.Errorf("range read for object %v incorrect (offset: %v, size: %v)", name, offset, size)
	}
	return nil
}

func loadTest(c obj.Client) error {
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
			if err := writeObject(ctx, c, name, data); err != nil {
				return err
			}
			buf := bufPool.GetBuffer()
			defer bufPool.PutBuffer(buf)
			if err := readObject(ctx, c, name, 0, 0, buf); err != nil {
				return err
			}
			if !bytes.Equal(data, buf) {
				return errors.Errorf("data writen does not equal data read for object %v", i)
			}
			return c.Delete(ctx, name)
		})
	}
	return eg.Wait()
}
