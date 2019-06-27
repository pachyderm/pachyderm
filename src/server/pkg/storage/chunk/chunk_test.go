package chunk

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

const (
	averageBits = 23
)

func Write(t *testing.T, chunks *Storage, n, rangeSize int) ([]*DataRef, []byte) {
	var finalDataRefs []*DataRef
	var seq []byte
	t.Run("Write", func(t *testing.T) {
		f := func(_ *DataRef, annotations []*Annotation) error {
			for _, a := range annotations {
				finalDataRefs = append(finalDataRefs, a.NextDataRef)
			}
			return nil
		}
		w := chunks.NewWriter(context.Background(), averageBits, f)
		seq = RandSeq(n * MB)
		for i := 0; i < n/rangeSize; i++ {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
			})
			_, err := w.Write(seq[i*MB*rangeSize : (i+1)*MB*rangeSize])
			require.NoError(t, err)
		}
		require.NoError(t, w.Close())
	})
	return finalDataRefs, seq
}

func TestWriteThenRead(t *testing.T) {
	objC, chunks := LocalStorage(t)
	defer Cleanup(objC, chunks)
	finalDataRefs, seq := Write(t, chunks, 100, 1)
	mid := len(finalDataRefs) / 2
	initialRefs := finalDataRefs[:mid]
	streamRefs := finalDataRefs[mid:]
	r := chunks.NewReader(context.Background())
	r.NextRange(initialRefs)
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
	defer Cleanup(objC, chunks)
	seq := RandSeq(100 * MB)
	b.SetBytes(100 * MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := func(_ *DataRef, _ []*Annotation) error { return nil }
		w := chunks.NewWriter(context.Background(), averageBits, f)
		for i := 0; i < 100; i++ {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
			})
			_, err := w.Write(seq[i*MB : (i+1)*MB])
			require.NoError(b, err)
		}
		require.NoError(b, w.Close())
	}
}

func TestWriteToN(t *testing.T) {
	objC, chunks := LocalStorage(t)
	defer Cleanup(objC, chunks)
	// Write the initial data and count the chunks.
	dataRefs1, seq1 := Write(t, chunks, 60, 20)
	dataRefs2, seq2 := Write(t, chunks, 60, 20)
	var initialChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		initialChunkCount++
		return nil
	}))
	// Copy data from readers into new writer.
	var finalDataRefs []*DataRef
	f := func(_ *DataRef, annotations []*Annotation) error {
		for _, a := range annotations {
			finalDataRefs = append(finalDataRefs, a.NextDataRef)
		}
		return nil
	}
	w := chunks.NewWriter(context.Background(), averageBits, f)
	r1 := chunks.NewReader(context.Background())
	r1.NextRange(dataRefs1)
	r2 := chunks.NewReader(context.Background())
	r2.NextRange(dataRefs2)
	w.Annotate(&Annotation{
		NextDataRef: &DataRef{},
	})
	mid := r1.Len() / 2
	require.NoError(t, r1.WriteToN(w, r1.Len()-mid))
	require.NoError(t, r1.WriteToN(w, mid))
	mid = r2.Len() / 2
	require.NoError(t, r2.WriteToN(w, r2.Len()-mid))
	require.NoError(t, r2.WriteToN(w, mid))
	require.NoError(t, w.Close())
	// Check that the initial data equals the final data.
	buf := &bytes.Buffer{}
	finalR := chunks.NewReader(context.Background())
	finalR.NextRange(finalDataRefs)
	_, err := io.Copy(buf, finalR)
	require.NoError(t, err)
	require.Equal(t, append(seq1, seq2...), buf.Bytes())
	// Only one extra chunk should get created when connecting the two sets of data.
	var finalChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		finalChunkCount++
		return nil
	}))
	require.Equal(t, initialChunkCount+1, finalChunkCount)
}
