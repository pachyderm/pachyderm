package fileset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	max     = 20 * units.MB
	maxTags = 10
)

type testFile struct {
	path string
	tag  string
	data []byte
}

func writeFileSet(t *testing.T, s *Storage, files []*testFile) ID {
	w := s.NewWriter(context.Background())
	for _, file := range files {
		require.NoError(t, w.Add(file.path, file.tag, bytes.NewReader(file.data)))
	}
	id, err := w.Close()
	require.NoError(t, err)
	return *id
}

func checkFile(t *testing.T, f File, tf *testFile) {
	r, w := io.Pipe()
	eg := errgroup.Group{}
	eg.Go(func() error {
		return f.Content(w)
	})
	eg.Go(func() (retErr error) {
		defer func() {
			if retErr != nil {
				r.CloseWithError(retErr)
			}
		}()
		actual := make([]byte, len(tf.data))
		_, err := io.ReadFull(r, actual)
		if err != nil && err != io.EOF {
			return err
		}
		return r.Close()
	})
	require.NoError(t, eg.Wait())
}

// newTestStorage creates a storage object with a test db and test tracker
// both of those components are kept hidden, so this is only appropriate for testing this package.
func newTestStorage(t *testing.T) *Storage {
	db := testutil.NewTestDB(t)
	tr := track.NewTestTracker(t, db)
	return NewTestStorage(t, db, tr)
}

func TestWriteThenRead(t *testing.T) {
	ctx := context.Background()
	storage := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	random := rand.New(rand.NewSource(seed))
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		for _, tagInt := range random.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path: "/" + fileName,
				tag:  tag,
				data: data,
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
		for _, tagInt := range random.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path: "/" + fileName,
				tag:  tag,
				data: data,
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
			path: files[idx].path,
			tag:  files[idx].tag,
			data: data,
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
		for _, tagInt := range random.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := randutil.Bytes(random, random.Intn(max))
			files = append(files, &testFile{
				path: "/" + fileName,
				tag:  tag,
				data: data,
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

func TestStableHash(t *testing.T) {
	ctx := context.Background()
	storage := newTestStorage(t)
	seed := time.Now().UTC().UnixNano()
	msg := fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
	random := rand.New(rand.NewSource(seed))
	data := randutil.Bytes(random, 100*units.MB)
	var ids []ID
	write := func(data []byte) {
		w := storage.NewWriter(ctx)
		require.NoError(t, w.Add("test", DefaultFileTag, bytes.NewReader(data)), msg)
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
			hash, err = f.Hash()
			return err
		}), msg)
		if !found {
			t.Fatal("file not found")
		}
		return hash

	}
	// Compute hash after writing to one writer.
	write(data)
	stableHash := getHash()
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
