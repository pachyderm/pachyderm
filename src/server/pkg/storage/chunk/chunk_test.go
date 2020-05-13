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
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"modernc.org/mathutil"
)

type test struct {
	maxAnnotationSize int
	maxTagSize        int
	n                 int
}

func (t test) name() string {
	return fmt.Sprintf("Max Annotation Size: %v, Max Tag Size: %v, Data Size: %v", units.HumanSize(float64(t.maxAnnotationSize)), units.HumanSize(float64(t.maxTagSize)), units.HumanSize(float64(t.n)))
}

var tests = []test{
	test{1 * units.KB, 1 * units.KB, 1 * units.KB},
	test{1 * units.KB, 1 * units.KB, 1 * units.MB},
	test{1 * units.MB, 1 * units.KB, 100 * units.MB},
	test{1 * units.MB, 1 * units.MB, 100 * units.MB},
	test{10 * units.MB, 1 * units.MB, 100 * units.MB},
}

// (bryce) this should be somewhere else (probably testutil).
func seedRand() string {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func TestWriteThenRead(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(objC obj.Client, chunks *Storage) error {
		msg := seedRand()
		for _, test := range tests {
			t.Run(test.name(), func(t *testing.T) {
				// Generate set of annotations.
				as := generateAnnotations(test)
				// Write then read the set of annotations.
				writeAnnotations(t, chunks, as, msg)
				readAnnotations(t, chunks, as, msg)
			})
		}
		return nil
	}))
}

func TestCopy(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(objC obj.Client, chunks *Storage) error {
		msg := seedRand()
		for _, test := range tests {
			t.Run(test.name(), func(t *testing.T) {
				// Generate two sets of annotations.
				as1 := generateAnnotations(test)
				as2 := generateAnnotations(test)
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
				w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), f)
				copyAnnotations(t, chunks, w, as, msg)
				require.NoError(t, w.Close(), msg)
				// Check that the annotations were correctly copied.
				readAnnotations(t, chunks, as, msg)
				// Check that at least one chunk was copied when connecting the two sets of annotations.
				var finalChunkCount int64
				require.NoError(t, chunks.List(context.Background(), func(_ string) error {
					finalChunkCount++
					return nil
				}), msg)
				require.True(t, finalChunkCount < initialChunkCount*2, msg)
			})
		}
		return nil
	}))
}

func BenchmarkWriter(b *testing.B) {
	require.NoError(b, WithLocalStorage(func(objC obj.Client, chunks *Storage) error {
		seq := RandSeq(100 * units.MB)
		b.SetBytes(100 * units.MB)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f := func(_ []*Annotation) error { return nil }
			w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), f)
			for i := 0; i < 100; i++ {
				w.Annotate(&Annotation{
					NextDataRef: &DataRef{},
				})
				w.Tag(strconv.Itoa(i))
				_, err := w.Write(seq[i*units.MB : (i+1)*units.MB])
				require.NoError(b, err)
			}
			require.NoError(b, w.Close())
		}
		return nil
	}))
}

func BenchmarkRollingHash(b *testing.B) {
	seq := RandSeq(100 * units.MB)
	b.SetBytes(100 * units.MB)
	hash := buzhash64.New()
	splitMask := uint64((1 << uint64(23)) - 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(initialWindow)
		for _, bt := range seq {
			hash.Roll(bt)
			//lint:ignore SA9003 benchmark is simulating exact usecase
			if hash.Sum64()&splitMask == 0 {
			}
		}
	}
}

type testAnnotation struct {
	data     []byte
	tags     []*Tag
	dataRefs []*DataRef
}

func generateAnnotations(t test) []*testAnnotation {
	var as []*testAnnotation
	for t.n > 0 {
		a := &testAnnotation{}
		a.data = RandSeq(mathutil.Min(rand.Intn(t.maxAnnotationSize)+1, t.n))
		a.tags = generateTags(t.maxTagSize, len(a.data))
		t.n -= len(a.data)
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
		w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), f)
		for _, a := range annotations {
			w.Annotate(&Annotation{
				NextDataRef: &DataRef{},
				Data:        a,
			})
			var offset int
			for _, tag := range a.tags {
				w.Tag(tag.Id)
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
				return w.Copy(dr.BoundReader())
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
