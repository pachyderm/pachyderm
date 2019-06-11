package chunk

import (
	"context"
	fmt "fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// LocalStorage creates a local chunk storage instance.
// Useful for storage layer tests.
func LocalStorage(tb testing.TB) (obj.Client, *Storage) {
	wd, err := os.Getwd()
	require.NoError(tb, err)
	objC, err := obj.NewLocalClient(wd)
	require.NoError(tb, err)
	return objC, NewStorage(objC)
}

// Cleanup cleans up a local chunk storage instance.
func Cleanup(objC obj.Client, chunks *Storage) {
	chunks.DeleteAll(context.Background())
	objC.Delete(context.Background(), prefix)
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

// CopyN is an efficient copy function that turns full chunk copies into data reference writes.
func CopyN(w *Writer, r *Reader, n int64) error {
	if r.Len() < n {
		// (bryce) this should be an internal error.
		return fmt.Errorf("number of bytes to copy (%v) is greater than reader length (%v)", n, r.Len())
	}
	for {
		// Read from the current data reference first.
		if r.r.Len() > 0 {
			nCopied, err := io.CopyN(w, r, int64(math.Min(float64(n), float64(r.r.Len()))))
			n -= nCopied
			if err != nil {
				return err
			}
		}
		// Done when there are no bytes left to read.
		if n == 0 {
			return nil
		}
		// A data reference can be cheaply copied when:
		// - The writer is at a split point.
		// - The data reference is a full chunk reference.
		// - The size of the chunk is less than or equal to the number of bytes left.
		for w.buf.Len() == 0 && r.dataRefs[0].Hash == "" && r.dataRefs[0].SizeBytes <= n {
			w.dataRefs[len(w.dataRefs)-1] = r.dataRefs[0]
			w.dataRefs = append(w.dataRefs, &DataRef{})
			w.rangeSize += r.dataRefs[0].SizeBytes
			n -= r.dataRefs[0].SizeBytes
			r.dataRefs = r.dataRefs[1:]
		}
		// Setup next data reference for reading.
		if err := r.nextDataRef(); err != nil {
			return err
		}
	}
}
