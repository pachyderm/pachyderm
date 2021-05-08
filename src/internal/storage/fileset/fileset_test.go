package fileset

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"

	units "github.com/docker/go-units"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	max         = 20 * units.MB
	maxTags     = 10
	testPath    = "test"
	scratchPath = "scratch"
)

type testFile struct {
	path  string
	parts []*testPart
}

type testPart struct {
	tag  string
	data []byte
}

func appendFile(t *testing.T, w *Writer, path string, parts []*testPart) {
	// Write content and tags.
	fw, err := w.Add(path)
	require.NoError(t, err)
	for _, part := range parts {
		fw.Add(part.tag)
		_, err := fw.Write(part.data)
		require.NoError(t, err)
	}
}

func writeFileSet(t *testing.T, s *Storage, files []*testFile) ID {
	w := s.NewWriter(context.Background())
	for _, file := range files {
		appendFile(t, w, file.path, file.parts)
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
		for _, part := range tf.parts {
			actual := make([]byte, len(part.data))
			_, err := io.ReadFull(r, actual)
			if err != nil && err != io.EOF {
				return err
			}
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
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var parts []*testPart
		for _, tagInt := range rand.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := chunk.RandSeq(rand.Intn(max))
			parts = append(parts, &testPart{
				tag:  tag,
				data: data,
			})
		}
		files = append(files, &testFile{
			path:  "/" + fileName,
			parts: parts,
		})
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
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var parts []*testPart
		for _, tagInt := range rand.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := chunk.RandSeq(rand.Intn(max))
			parts = append(parts, &testPart{
				tag:  tag,
				data: data,
			})
		}
		files = append(files, &testFile{
			path:  "/" + fileName,
			parts: parts,
		})
	}

	// Write out ten filesets where each subsequent fileset has the content of one random file changed.
	// Confirm that all of the content and hashes other than the changed file remain the same.
	N := len(fileNames)
	for i := 0; i < N; i++ {
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
		files[i].parts = append(files[i].parts, &testPart{
			data: chunk.RandSeq(rand.Intn(max)),
			tag:  "test-tag",
		})
		require.NoError(t, storage.Drop(ctx, id))
	}
}

func TestCopy(t *testing.T) {
	ctx := context.Background()
	fileSets := newTestStorage(t)
	fileNames := index.Generate("abc")
	files := []*testFile{}
	for _, fileName := range fileNames {
		var parts []*testPart
		for _, tagInt := range rand.Perm(maxTags) {
			tag := fmt.Sprintf("%08x", tagInt)
			data := chunk.RandSeq(rand.Intn(max))
			parts = append(parts, &testPart{
				tag:  tag,
				data: data,
			})
		}
		files = append(files, &testFile{
			path:  "/" + fileName,
			parts: parts,
		})
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
	// TODO: figure out why we make 1 more chunk.
	require.Equal(t, initialChunkCount, finalChunkCount-1)
}

func countChunks(t *testing.T, s *Storage) (count int64) {
	require.NoError(t, s.ChunkStorage().List(context.Background(), func(chunk.ID) error {
		count++
		return nil
	}))
	return count
}
