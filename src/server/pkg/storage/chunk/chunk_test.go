package chunk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"modernc.org/mathutil"
)

const (
	averageBits = 23
)

var tests = []struct {
	maxAnnotationSize int
	maxTagSize        int
	n                 int
}{
	{1 * KB, 1 * KB, 1 * KB},
	{1 * KB, 1 * KB, 1 * MB},
	{1 * MB, 1 * KB, 100 * MB},
	{1 * MB, 1 * MB, 100 * MB},
	{10 * MB, 1 * MB, 100 * MB},
}

func TestWriteThenRead(t *testing.T) {
	objC, chunks := LocalStorage(t)
	defer Cleanup(objC, chunks)
	// Setup seed.
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	for _, test := range tests {
		// Generate set of annotations.
		as := generateAnnotations(test.maxAnnotationSize, test.maxTagSize, test.n)
		// Write then read the set of annotations.
		writeAnnotations(t, chunks, as, msg)
		readAnnotations(t, chunks, as, msg)
	}
}

func TestCopy(t *testing.T) {
	objC, chunks := LocalStorage(t)
	defer Cleanup(objC, chunks)
	// Setup seed.
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	for _, test := range tests {
		// Generate two sets of annotations.
		as1 := generateAnnotations(test.maxAnnotationSize, test.maxTagSize, test.n)
		as2 := generateAnnotations(test.maxAnnotationSize, test.maxTagSize, test.n)
		// Write the two sets of annotations.
		writeAnnotations(t, chunks, as1, msg)
		writeAnnotations(t, chunks, as2, msg)
		// Initial chunk count.
		var initialChunkCount int64
		require.NoError(t, chunks.List(context.Background(), func(_ string) error {
			initialChunkCount++
			return nil
		}), msg)
		// Copy the annotations from the two sets of annotations.
		as := append(as1, as2...)
		f := func(annotations []*Annotation) error {
			for _, a := range annotations {
				testA := a.Data.(*testAnnotation)
				testA.dataRefs = append(testA.dataRefs, a.NextDataRef)
			}
			return nil
		}
		w := chunks.NewWriter(context.Background(), averageBits, f, 0)
		copyAnnotations(t, chunks, w, as, msg)
		require.NoError(t, w.Close(), msg)
		// Check that the annotations were correctly copied.
		readAnnotations(t, chunks, as, msg)
		// Check that only one new chunk was created when connecting the two sets of annotations.
		var finalChunkCount int64
		require.NoError(t, chunks.List(context.Background(), func(_ string) error {
			finalChunkCount++
			return nil
		}), msg)
		require.Equal(t, initialChunkCount+1, finalChunkCount, msg)
	}
}

func BenchmarkWriter(b *testing.B) {
	objC, chunks := LocalStorage(b)
	defer Cleanup(objC, chunks)
	seq := RandSeq(100 * MB)
	b.SetBytes(100 * MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := func(_ []*Annotation) error { return nil }
		w := chunks.NewWriter(context.Background(), averageBits, f, 0)
		for i := 0; i < 100; i++ {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
			})
			w.StartTag(strconv.Itoa(i))
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

func seedStr(seed int64) string {
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

type testAnnotation struct {
	data     []byte
	tags     []*Tag
	dataRefs []*DataRef
}

func generateAnnotations(maxAnnotationSize, maxTagSize, n int) []*testAnnotation {
	var as []*testAnnotation
	for n > 0 {
		a := &testAnnotation{}
		a.data = RandSeq(mathutil.Min(rand.Intn(maxAnnotationSize)+1, n))
		a.tags = generateTags(maxTagSize, len(a.data))
		n -= len(a.data)
		as = append(as, a)
	}
	return as
}

func generateTags(maxTagSize, n int) []*Tag {
	var tags []*Tag
	for n > 0 {
		tagSize := mathutil.Min(rand.Intn(maxTagSize)+1, n)
		n -= tagSize
		tags = append(tags, &Tag{
			Id:        strconv.Itoa(len(tags)),
			SizeBytes: int64(tagSize),
		})
	}
	return tags
}

func writeAnnotations(t *testing.T, chunks *Storage, annotations []*testAnnotation, msg string) {
	t.Run("Write", func(t *testing.T) {
		f := func(annotations []*Annotation) error {
			for _, a := range annotations {
				testA := a.Data.(*testAnnotation)
				testA.dataRefs = append(testA.dataRefs, a.NextDataRef)
			}
			return nil
		}
		w := chunks.NewWriter(context.Background(), averageBits, f, 0)
		for _, a := range annotations {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
				Data:        a,
			})
			var offset int
			for _, tag := range a.tags {
				w.StartTag(tag.Id)
				_, err := w.Write(a.data[offset : offset+int(tag.SizeBytes)])
				require.NoError(t, err, msg)
				offset += int(tag.SizeBytes)
			}
		}
		require.NoError(t, w.Close(), msg)
	})
}

func copyAnnotations(t *testing.T, chunks *Storage, w *Writer, annotations []*testAnnotation, msg string) {
	t.Run("Copy", func(t *testing.T) {
		r := chunks.NewReader(context.Background())
		for _, a := range annotations {
			r.NextDataRefs(a.dataRefs)
			a.dataRefs = nil
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
				Data:        a,
			})
			require.NoError(t, r.Iterate(func(dr *DataReader) error {
				return w.Copy(dr.LimitReader())
			}), msg)
		}
	})
}

func readAnnotations(t *testing.T, chunks *Storage, annotations []*testAnnotation, msg string) {
	t.Run("Read", func(t *testing.T) {
		r := chunks.NewReader(context.Background())
		for _, a := range annotations {
			r.NextDataRefs(a.dataRefs)
			data := &bytes.Buffer{}
			var tags []*Tag
			require.NoError(t, r.Iterate(func(dr *DataReader) error {
				return dr.Iterate(func(tag *Tag, r io.Reader) error {
					tags = joinTags(tags, []*Tag{tag})
					_, err := io.Copy(data, r)
					return err
				})
			}), msg)
			require.Equal(t, 0, bytes.Compare(a.data, data.Bytes()), msg)
			require.Equal(t, a.tags, tags, msg)
		}
	})
}
