package fileset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"

	units "github.com/docker/go-units"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
)

const (
	max       = 20 * units.MB
	maxDatums = 10
)

type testFile struct {
	path  string
	datum string
	data  []byte
}

func writeFileSet(t *testing.T, s *Storage, files []*testFile) ID {
	w := s.NewWriter(context.Background())
	for _, file := range files {
		require.NoError(t, w.Add(file.path, file.datum, bytes.NewReader(file.data)))
	}
	id, err := w.Close()
	require.NoError(t, err)
	return *id
}

func checkFile(t *testing.T, f File, tf *testFile) {
	require.NoError(t, miscutil.WithPipe(func(w io.Writer) error {
		return errors.EnsureStack(f.Content(context.Background(), w))
	}, func(r io.Reader) error {
		actual := make([]byte, len(tf.data))
		_, err := io.ReadFull(r, actual)
		return errors.EnsureStack(err)
	}))
}

// newTestStorage creates a storage object with a test db and test tracker
// both of those components are kept hidden, so this is only appropriate for testing this package.
func newTestStorage(tb testing.TB) *Storage {
	db := dockertestenv.NewTestDB(tb)
	tr := track.NewTestTracker(tb, db)
	return NewTestStorage(tb, db, tr)
}

func TestWriteThenRead(t *testing.T) {
	ctx := context.Background()
	storage := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var datums []string
		for _, datumInt := range random.Perm(maxDatums) {
			datums = append(datums, fmt.Sprintf("%08x", datumInt))
		}
		sort.Strings(datums)
		for _, datum := range datums {
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path:  "/" + fileName,
				datum: datum,
				data:  data,
			})
		}
	}

	// Write the files to the fileset.
	id := writeFileSet(t, storage, files)

	// Read the files from the fileset, checking against the recorded files.
	fs, err := storage.Open(ctx, []ID{id})
	require.NoError(t, err)
	fileIter := files
	err = fs.Iterate(ctx, func(f File) error {
		tf := fileIter[0]
		fileIter = fileIter[1:]
		checkFile(t, f, tf)
		return nil
	})
	require.NoError(t, err)
}

func TestWriteThenReadFuzz(t *testing.T) {
	ctx := context.Background()
	storage := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var datums []string
		for _, datumInt := range random.Perm(maxDatums) {
			datums = append(datums, fmt.Sprintf("%08x", datumInt))
		}
		sort.Strings(datums)
		for _, datum := range datums {
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path:  "/" + fileName,
				datum: datum,
				data:  data,
			})
		}
	}

	// Write out ten filesets where each subsequent fileset has the content of one random file changed.
	// Confirm that all of the content and hashes other than the changed file remain the same.
	for i := 0; i < 10; i++ {
		// Write the files to the fileset.
		id := writeFileSet(t, storage, files)
		r, err := storage.Open(ctx, []ID{id})
		require.NoError(t, err)
		filesIter := files
		require.NoError(t, r.Iterate(ctx, func(f File) error {
			checkFile(t, f, filesIter[0])
			filesIter = filesIter[1:]
			return nil
		}))
		idx := random.Intn(len(files))
		data := randutil.Bytes(random, random.Intn(max))
		files[idx] = &testFile{
			path:  files[idx].path,
			datum: files[idx].datum,
			data:  data,
		}
		require.NoError(t, storage.Drop(ctx, id))
	}
}

func TestCopy(t *testing.T) {
	ctx := context.Background()
	fileSets := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var datums []string
		for _, datumInt := range random.Perm(maxDatums) {
			datums = append(datums, fmt.Sprintf("%08x", datumInt))
		}
		sort.Strings(datums)
		for _, datum := range datums {
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path:  "/" + fileName,
				datum: datum,
				data:  data,
			})
		}
	}
	originalID := writeFileSet(t, fileSets, files)

	initialChunkCount := countChunks(t, fileSets)
	// Copy intial fileset to a new copy fileset.
	r := fileSets.newReader(originalID)
	wCopy := fileSets.newWriter(context.Background())
	require.NoError(t, CopyFiles(ctx, wCopy, r))
	copyID, err := wCopy.Close()
	require.NoError(t, err)

	// Compare initial fileset and copy fileset.
	rCopy := fileSets.newReader(*copyID)
	require.NoError(t, rCopy.Iterate(ctx, func(f File) error {
		checkFile(t, f, files[0])
		files = files[1:]
		return nil
	}))
	// No new chunks should get created by the copy.
	finalChunkCount := countChunks(t, fileSets)
	require.Equal(t, initialChunkCount, finalChunkCount)
}

func countChunks(t *testing.T, s *Storage) (count int64) {
	require.NoError(t, s.ChunkStorage().List(context.Background(), func(chunk.ID) error {
		count++
		return nil
	}))
	return count
}

// This test ensures that future changes do not affect the stable hashes of files since that would be a breaking change.
func TestStableHash(t *testing.T) {
	type testData struct {
		seed     int64
		expected []string
	}
	tds := []*testData{
		&testData{
			seed: 1648577872380609229,
			expected: []string{
				"27e12145099615b6bf0364a4472452dfe0e8105e6d58d7fbc5d0c038c7a50736",
				"5672e6f3e1841f3f1e284c2d4b7c12dc213ffc88878c9d3e2302be8acd0198ef",
			},
		},
		&testData{
			seed: 1648742949150704545,
			expected: []string{
				"020e7fda5b3d81b000918ac4fa808a6569fc01000d15322247206aeeea1761a2",
				"d1842da7cfdc60c4a63c8ab58bd8eee4fa765a270743f3aa36ece6cbef786423",
			},
		},
		&testData{
			seed: 1648742961991348032,
			expected: []string{
				"8a801a53ba2933cccea69fdb61eb3aa04b2019eea541aa848c6b42ef4c7262d5",
				"d578bcae30c790722048f7d698a6279049012a42b339e802da07bca269831b30",
			},
		},
		&testData{
			seed: 1648742974827637769,
			expected: []string{
				"70e7271fc45251fcf788175df481ab6fd090f37860fcdc63b9f25c26834adf52",
				"b480bf9cffce51b52615a59dc254647145c6376551cf864d2eb87c63bd11344d",
			},
		},
		&testData{
			seed: 1648742988445537518,
			expected: []string{
				"ca46ed3b9bc6a9b090d94a71d4d308ca05580ab467681cdb236b097206feba5a",
				"79d4bc9c45b07aecaf19b0fe3cea273b059d70c23dabbe3f9d575c1051dddd44",
			},
		},
	}
	for _, td := range tds {
		msg := fmt.Sprint("seed: ", strconv.FormatInt(td.seed, 10))
		output, err := pachhash.ParseHex([]byte(td.expected[0]))
		require.NoError(t, err)
		random := rand.New(rand.NewSource(td.seed))
		testStableHash(t, oldRandomBytes(random, 100*units.KB), output[:], msg)
		output, err = pachhash.ParseHex([]byte(td.expected[1]))
		require.NoError(t, err)
		random = rand.New(rand.NewSource(td.seed))
		testStableHash(t, oldRandomBytes(random, 100*units.MB), output[:], msg)
	}
}

func TestStableHashFuzz(t *testing.T) {
	seed := time.Now().UTC().UnixNano()
	msg := fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
	random := rand.New(rand.NewSource(seed))
	testStableHash(t, randutil.Bytes(random, 100*units.KB), nil, msg)
	testStableHash(t, randutil.Bytes(random, 100*units.MB), nil, msg)
}

var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// This is an older way of generating random bytes; what randutil.Bytes used to do.  The algorithm
// it uses has changed in favor of a faster one, but breaks the above test.
func oldRandomBytes(random *rand.Rand, n int) []byte {
	bs := make([]byte, n)
	for i := range bs {
		bs[i] = letters[random.Intn(len(letters))]
	}
	return bs
}

func testStableHash(t *testing.T, data, expected []byte, msg string) {
	ctx := context.Background()
	storage := newTestStorage(t)
	var ids []ID
	write := func(data []byte) {
		w := storage.NewWriter(ctx)
		require.NoError(t, w.Add("test", DefaultFileDatum, bytes.NewReader(data)), msg)
		id, err := w.Close()
		require.NoError(t, err, msg)
		ids = append(ids, *id)
	}
	getHash := func() []byte {
		fs, err := storage.Open(ctx, ids)
		require.NoError(t, err, msg)
		var found bool
		var hash []byte
		require.NoError(t, fs.Iterate(ctx, func(f File) error {
			if found {
				return errors.New("more than one file found")
			}
			found = true
			var err error
			hash, err = f.Hash(ctx)
			return errors.EnsureStack(err)
		}), msg)
		if !found {
			t.Fatal("file not found")
		}
		return hash

	}
	// Compute hash after writing to one writer.
	write(data)
	stableHash := getHash()
	if expected != nil {
		require.True(t, bytes.Equal(expected, stableHash))
	}
	// Compute hash after writing to two writers.
	ids = nil
	size := len(data) / 2
	for offset := 0; offset < len(data); offset += size {
		write(data[offset : offset+size])
	}
	require.True(t, bytes.Equal(stableHash, getHash()), msg)
	// Compute hash after writing to ten writers.
	ids = nil
	size = len(data) / 10
	for offset := 0; offset < len(data); offset += size {
		write(data[offset : offset+size])
	}
	require.True(t, bytes.Equal(stableHash, getHash()), msg)
	// Compute hash after writing to one hundred writers.
	ids = nil
	size = len(data) / 100
	for offset := 0; offset < len(data); offset += size {
		write(data[offset : offset+size])
	}
	require.True(t, bytes.Equal(stableHash, getHash()), msg)
}
