package chunk

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func TestWriteThenRead(t *testing.T) {
	objC, chunks := LocalStorage(t)
	defer func() {
		require.NoError(t, chunks.DeleteAll(context.Background()))
		require.NoError(t, objC.Delete(context.Background(), Prefix))
	}()
	var finalDataRefs []*DataRef
	var seq []byte
	t.Run("Write", func(t *testing.T) {
		w := chunks.NewWriter(context.Background())
		cb := func(dataRefs []*DataRef) error {
			finalDataRefs = append(finalDataRefs, dataRefs...)
			return nil
		}
		seq = RandSeq(100 * MB)
		for i := 0; i < 100; i++ {
			w.StartRange(cb)
			_, err := w.Write(seq[i*MB : (i+1)*MB])
			require.NoError(t, err)
		}
		require.NoError(t, w.Close())
	})
	mid := len(finalDataRefs) / 2
	initialRefs := finalDataRefs[:mid]
	streamRefs := finalDataRefs[mid:]
	r := chunks.NewReader(context.Background(), initialRefs...)
	buf := &bytes.Buffer{}
	t.Run("ReadInitial", func(t *testing.T) {
		_, err := io.Copy(buf, r)
		require.NoError(t, err)
		require.Equal(t, bytes.Compare(buf.Bytes(), seq[:buf.Len()]), 0)
	})
	seq = seq[buf.Len():]
	buf.Reset()
	t.Run("ReadStream", func(t *testing.T) {
		for _, ref := range streamRefs {
			r.NextRange([]*DataRef{ref})
			_, err := io.Copy(buf, r)
			require.NoError(t, err)
		}
		require.Equal(t, bytes.Compare(buf.Bytes(), seq), 0)
	})
}

func BenchmarkWriter(b *testing.B) {
	objC, chunks := LocalStorage(b)
	defer func() {
		require.NoError(b, chunks.DeleteAll(context.Background()))
		require.NoError(b, objC.Delete(context.Background(), Prefix))
	}()
	seq := RandSeq(100 * MB)
	b.SetBytes(100 * MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := chunks.NewWriter(context.Background())
		cb := func(dataRefs []*DataRef) error { return nil }
		for i := 0; i < 100; i++ {
			w.StartRange(cb)
			_, err := w.Write(seq[i*MB : (i+1)*MB])
			require.NoError(b, err)
		}
		require.NoError(b, w.Close())
	}
}
