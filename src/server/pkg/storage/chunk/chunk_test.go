package chunk

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

const (
	averageBits = 23
)

func Write(t *testing.T, chunks *Storage, n, annotationSize int) ([]*DataRef, []byte) {
	var finalDataRefs []*DataRef
	var seq []byte
	t.Run("Write", func(t *testing.T) {
		f := func(_ *DataRef, annotations []*Annotation) error {
			for _, a := range annotations {
				finalDataRefs = append(finalDataRefs, a.NextDataRef)
			}
			return nil
		}
		w := chunks.NewWriter(context.Background(), averageBits, f, 0)
		seq = RandSeq(n * MB)
		for i := 0; i < n/annotationSize; i++ {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
				Meta:        i,
			})
			w.StartTag(strconv.Itoa(i))
			_, err := w.Write(seq[i*MB*annotationSize : (i+1)*MB*annotationSize])
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
	r.NextDataRefs(initialRefs)
	buf := &bytes.Buffer{}
	t.Run("ReadInitial", func(t *testing.T) {
		require.NoError(t, r.Get(buf))
		require.Equal(t, 0, bytes.Compare(buf.Bytes(), seq[:buf.Len()]))
	})
	seq = seq[buf.Len():]
	buf.Reset()
	t.Run("ReadStream", func(t *testing.T) {
		for _, ref := range streamRefs {
			r.NextDataRefs([]*DataRef{ref})
			require.NoError(t, r.Get(buf))
		}
		require.Equal(t, 0, bytes.Compare(buf.Bytes(), seq))
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
		w := chunks.NewWriter(context.Background(), averageBits, f, 0)
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

func BenchmarkRollingHash(b *testing.B) {
	seq := RandSeq(100 * MB)
	b.SetBytes(100 * MB)
	hash := buzhash64.New()
	splitMask := uint64((1 << uint64(23)) - 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(initialWindow)
		for _, bt := range seq {
			hash.Roll(bt)
			if hash.Sum64()&splitMask == 0 {
			}
		}
	}
}

func TestCopy(t *testing.T) {
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
	w := chunks.NewWriter(context.Background(), averageBits, f, 0)
	r1 := chunks.NewReader(context.Background())
	r1.NextDataRefs(dataRefs1)
	r2 := chunks.NewReader(context.Background())
	r2.NextDataRefs(dataRefs2)
	w.Annotate(&Annotation{
		NextDataRef: &DataRef{},
	})
	require.NoError(t, r1.Iterate(func(dr *DataReader) error {
		return w.Copy(dr.LimitReader())
	}))
	require.NoError(t, r2.Iterate(func(dr *DataReader) error {
		return w.Copy(dr.LimitReader())
	}))
	require.NoError(t, w.Close())
	// Check that the initial data equals the final data.
	buf := &bytes.Buffer{}
	finalR := chunks.NewReader(context.Background())
	finalR.NextDataRefs(finalDataRefs)
	require.NoError(t, finalR.Get(buf))
	require.Equal(t, append(seq1, seq2...), buf.Bytes())
	// Only one extra chunk should get created when connecting the two sets of data.
	var finalChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		finalChunkCount++
		return nil
	}))
	require.Equal(t, initialChunkCount+1, finalChunkCount)
}
