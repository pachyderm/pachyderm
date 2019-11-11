package chunk

import (
	"context"
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
