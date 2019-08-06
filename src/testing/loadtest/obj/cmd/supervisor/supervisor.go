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
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"golang.org/x/sync/errgroup"
)

const (
	// MB is a megabyte
	MB = 1024 * 1024
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
	// Setup object to write and client.
	data := RandSeq(objectSize)
	c, err := obj.NewClientFromEnv("/pach")
	if err != nil {
		log.Fatalf("Error creating client (%v)", err)
	}
	// Start timing load test.
	var start = time.Now()
	defer func() {
		fmt.Printf("Benchmark complete. Total time: %.3f\n", time.Now().Sub(start).Seconds())
	}()
	// Start writing objects.
	eg, ctx := errgroup.WithContext(context.Background())
	limiter := limit.New(concurrency)
	for i := 0; i < numObjects; i++ {
		i := i
		limiter.Acquire()
		eg.Go(func() error {
			defer limiter.Release()
			w, err := c.Writer(ctx, strconv.Itoa(i))
			if err != nil {
				return fmt.Errorf("Error creating writer for object %v (%v)", i, err)
			}
			r := bytes.NewReader(data)
			if _, err := io.Copy(w, r); err != nil {
				return fmt.Errorf("Error writing to object %v (%v)", i, err)
			}
			if err := w.Close(); err != nil {
				return fmt.Errorf("Error closing object %v (%v)", i, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatalf(err.Error())
	}
}
