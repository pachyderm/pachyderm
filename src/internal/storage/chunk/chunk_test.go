package chunk

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"modernc.org/mathutil"
)

type test struct {
	maxAnnotationSize int
	n                 int
}

func (t test) name() string {
	return fmt.Sprintf("Max Annotation Size: %v, Data Size: %v", units.HumanSize(float64(t.maxAnnotationSize)), units.HumanSize(float64(t.n)))
}

var tests = []test{
	test{1 * units.KB, 1 * units.KB},
	test{1 * units.KB, 1 * units.MB},
	test{1 * units.MB, 100 * units.MB},
	test{10 * units.MB, 100 * units.MB},
}

func TestWriteThenRead(t *testing.T) {
	_, chunks := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	msg := fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
	random := rand.New(rand.NewSource(seed))
	for _, test := range tests {
		t.Run(test.name(), func(t *testing.T) {
			// Generate set of annotations.
			as := generateAnnotations(random, test)
			// Write then read the set of annotations.
			writeAnnotations(t, chunks, as, msg)
			readAnnotations(t, chunks, as, msg)
		})
	}
}

func TestCopy(t *testing.T) {
	_, chunks := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	msg := fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
	random := rand.New(rand.NewSource(seed))
	for _, test := range tests {
		t.Run(test.name(), func(t *testing.T) {
			// Generate two sets of annotations.
			as1 := generateAnnotations(random, test)
			as2 := generateAnnotations(random, test)
			// Write the two sets of annotations.
			writeAnnotations(t, chunks, as1, msg)
			writeAnnotations(t, chunks, as2, msg)
			// Initial chunk count.
			var initialChunkCount int64
			require.NoError(t, chunks.List(context.Background(), func(_ ID) error {
				initialChunkCount++
				return nil
			}), msg)
			// Copy the annotations from the two sets of annotations.
			as := append(as1, as2...)
			cb := func(annotations []*Annotation) error {
				for _, a := range annotations {
					testA := a.Data.(*testAnnotation)
					if a.NextDataRef != nil {
						testA.dataRefs = append(testA.dataRefs, a.NextDataRef)
					}
				}
				return nil
			}
			w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), cb)
			copyAnnotations(t, chunks, w, as, msg)
			require.NoError(t, w.Close(), msg)
			// Check that the annotations were correctly copied.
			readAnnotations(t, chunks, as, msg)
			// Check that at least one chunk was copied when connecting the two sets of annotations.
			var finalChunkCount int64
			require.NoError(t, chunks.List(context.Background(), func(_ ID) error {
				finalChunkCount++
				return nil
			}), msg)
			require.True(t, finalChunkCount < initialChunkCount*2, msg)
		})
	}
}

func BenchmarkWriter(b *testing.B) {
	_, chunks := newTestStorage(b)
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	data := randutil.Bytes(random, 100*units.MB)
	b.SetBytes(100 * units.MB)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb := func(_ []*Annotation) error { return nil }
		w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), cb)
		for i := 0; i < 100; i++ {
			w.Annotate(&Annotation{})
			_, err := w.Write(data[i*units.MB : (i+1)*units.MB])
			require.NoError(b, err)
		}
		require.NoError(b, w.Close())
	}
}

func BenchmarkRollingHash(b *testing.B) {
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	data := randutil.Bytes(random, 100*units.MB)
	b.SetBytes(100 * units.MB)
	hash := buzhash64.New()
	splitMask := uint64((1 << uint64(23)) - 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.Reset()
		hash.Write(initialWindow)
		for _, bt := range data {
			hash.Roll(bt)
			//lint:ignore SA9003 benchmark is simulating exact usecase
			if hash.Sum64()&splitMask == 0 {
			}
		}
	}
}

type testAnnotation struct {
	data     []byte
	dataRefs []*DataRef
}

func generateAnnotations(random *rand.Rand, t test) []*testAnnotation {
	var as []*testAnnotation
	for t.n > 0 {
		a := &testAnnotation{}
		a.data = randutil.Bytes(random, mathutil.Min(rand.Intn(t.maxAnnotationSize)+1, t.n))
		t.n -= len(a.data)
		as = append(as, a)
	}
	return as
}

func writeAnnotations(t *testing.T, chunks *Storage, annotations []*testAnnotation, msg string) {
	t.Run("Write", func(t *testing.T) {
		cb := func(annotations []*Annotation) error {
			for _, a := range annotations {
				testA := a.Data.(*testAnnotation)
				// TODO: Document why NextDataRef can be nil.
				if a.NextDataRef != nil {
					testA.dataRefs = append(testA.dataRefs, a.NextDataRef)
				}
			}
			return nil
		}
		w := chunks.NewWriter(context.Background(), uuid.NewWithoutDashes(), cb)
		for _, a := range annotations {
			require.NoError(t, w.Annotate(&Annotation{
				Data: a,
			}))
			_, err := w.Write(a.data)
			require.NoError(t, err, msg)
		}
		require.NoError(t, w.Close(), msg)
	})
}

func copyAnnotations(t *testing.T, chunks *Storage, w *Writer, annotations []*testAnnotation, msg string) {
	t.Run("Copy", func(t *testing.T) {
		for _, a := range annotations {
			dataRefs := a.dataRefs
			a.dataRefs = nil
			require.NoError(t, w.Annotate(&Annotation{
				Data: a,
			}))
			for _, dataRef := range dataRefs {
				require.NoError(t, w.Copy(dataRef))
			}
		}
	})
}

func readAnnotations(t *testing.T, chunks *Storage, annotations []*testAnnotation, msg string) {
	t.Run("Read", func(t *testing.T) {
		for _, a := range annotations {
			r := chunks.NewReader(context.Background(), a.dataRefs)
			buf := &bytes.Buffer{}
			require.NoError(t, r.Get(buf), msg)
			require.Equal(t, 0, bytes.Compare(a.data, buf.Bytes()), msg)
		}
	})
}

// newTestStorage is like NewTestStorage except it doesn't need an external tracker
// it is for testing this package, not for reuse.
func newTestStorage(t testing.TB) (obj.Client, *Storage) {
	db := testutil.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	return NewTestStorage(t, db, tr)
}
