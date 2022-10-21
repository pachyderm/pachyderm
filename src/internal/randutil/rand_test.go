package randutil

import (
	"bytes"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var random = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

func BenchmarkReader(b *testing.B) {
	b.SetBytes(100 * units.MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		r := NewBytesReader(random, 100*units.MB)
		_, err := io.Copy(buf, r)
		require.NoError(b, err)
	}
}

func testLetters(t *testing.T, buf []byte) {
	t.Helper()
	got := map[byte]int{}
	for _, b := range buf {
		got[byte(b)]++
	}
	for _, letter := range letters {
		if n := got[letter]; n < 1 {
			t.Errorf("letter %s never appears", string(letter))
		}
	}
}

func TestBytes(t *testing.T) {
	testLetters(t, Bytes(random, 1000))
}

func TestBytesReader(t *testing.T) {
	r := NewBytesReader(random, 1000)
	buf, err := io.ReadAll(r)
	require.NoError(t, err, "should have read bytes from BytesReader")
	testLetters(t, buf)
}
