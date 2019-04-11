package chunk

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"os"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	prefix = "chunks"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// generates random sequence of data (n is number of bytes)
func randSeq(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}

func localStorage(tb testing.TB) (obj.Client, *Storage) {
	wd, err := os.Getwd()
	require.NoError(tb, err)
	objC, err := obj.NewLocalClient(wd)
	require.NoError(tb, err)
	return objC, NewStorage(objC, prefix)
}

func TestWriteThenRead(t *testing.T) {
	objC, s := localStorage(t)
	defer func() {
		require.NoError(t, s.Clear(context.Background()))
		require.NoError(t, objC.Delete(context.Background(), prefix))
	}()
	var finalDataRefs []*DataRef
	w := s.NewWriter(context.Background(), func(dataRefs []*DataRef) error {
		finalDataRefs = append(finalDataRefs, dataRefs...)
		return nil
	})
	seq := randSeq(100 * MB)
	for i := 0; i < 100; i++ {
		w.StartRange()
		_, err := w.Write(seq[i*MB : (i+1)*MB])
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())
	buf := &bytes.Buffer{}
	for _, dataRef := range finalDataRefs {
		r, err := s.NewReader(context.Background(), dataRef)
		require.NoError(t, err)
		_, err = io.Copy(buf, r)
		require.NoError(t, err)
	}
	require.Equal(t, bytes.Compare(buf.Bytes(), seq), 0)
}

func BenchmarkWriter(b *testing.B) {
	objC, s := localStorage(b)
	defer func() {
		require.NoError(b, s.Clear(context.Background()))
		require.NoError(b, objC.Delete(context.Background(), prefix))
	}()
	seq := randSeq(100 * MB)
	b.SetBytes(100 * MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := s.NewWriter(context.Background(), func(_ []*DataRef) error {
			return nil
		})
		for i := 0; i < 100; i++ {
			w.StartRange()
			_, err := w.Write(seq[i*MB : (i+1)*MB])
			require.NoError(b, err)
		}
		require.NoError(b, w.Close())
	}
}
