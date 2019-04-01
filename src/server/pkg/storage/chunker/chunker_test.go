package chunker

import (
	"bytes"
	"io"
	"math/rand"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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

// generates data for chunker (n is number of MB)
func genData(n int) [][]byte {
	data := [][]byte{}
	for i := 0; i < n; i++ {
		data = append(data, randSeq(MB))
	}
	return data
}

func TestChunker(t *testing.T) {
	buf := &bytes.Buffer{}
	chnkr := New(func(r io.Reader) error {
		io.Copy(buf, r)
		require.True(t, buf.Len() >= MinSize && buf.Len() <= MaxSize)
		buf.Reset()
		return nil
	})
	for _, seq := range genData(100) {
		chnkr.Write(seq)
	}
}

func BenchmarkChunker(b *testing.B) {
	data := genData(100)
	b.SetBytes(100 * MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chnkr := New(func(r io.Reader) error {
			return nil
		})
		for _, seq := range data {
			chnkr.Write(seq)
		}
	}
}
