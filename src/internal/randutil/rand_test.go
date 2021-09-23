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

func BenchmarkReader(b *testing.B) {
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	b.SetBytes(100 * units.MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		r := NewBytesReader(random, 100*units.MB)
		_, err := io.Copy(buf, r)
		require.NoError(b, err)
	}
}
