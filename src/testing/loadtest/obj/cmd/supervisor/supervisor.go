package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
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
	if exists, err := c.Exists(ctx, name); err != nil {
		return err
	} else if !exists {
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
		if err := readTest(ctx, c, name, data); err != nil {
			return err
		}
	}
	// Confirm range reads work correctly
	offset, size := basicObjectSize/2, 0
	if err := readTest(ctx, c, name, data[offset:]); err != nil {
		return err
	}
	offset, size = basicObjectSize/2, basicObjectSize/4
	if err := readTest(ctx, c, name, data[offset:offset+size]); err != nil {
		return err
	}
	// Walk the objects and for each check the existence and delete it.
	if err := walk(ctx, c, 5, func(name string) error {
		if exists, err := c.Exists(ctx, name); err != nil {
			return err
		} else if !exists {
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
	if err := readTest(ctx, c, "zero", data); err != nil {
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
	return c.Put(ctx, name, bytes.NewReader(data))
}

func readObject(ctx context.Context, c obj.Client, name string, p []byte) (retErr error) {
	buf := &bytes.Buffer{}
	if err := c.Get(ctx, name, buf); err != nil {
		return err
	}
	copy(p, buf.Bytes())
	return nil
}

func readTest(ctx context.Context, c obj.Client, name string, expected []byte) error {
	buf := make([]byte, len(expected))
	if err := readObject(ctx, c, name, buf); err != nil {
		return err
	}
	if !bytes.Equal(expected, buf) {
		return errors.Errorf("range read for object %v incorrect (offset: %v, size: %v)", name)
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
			if err := readObject(ctx, c, name, buf); err != nil {
				return err
			}
			if !bytes.Equal(data, buf) {
				return errors.Errorf("data written does not equal data read for object %v", i)
			}
			return c.Delete(ctx, name)
		})
	}
	return eg.Wait()
}
