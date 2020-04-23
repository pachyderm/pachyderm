package fileset

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"path"
	"strconv"
	"testing"
	"time"

	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

const (
	max         = 20 * units.MB
	maxTags     = 10
	testPath    = "test"
	scratchPath = "scratch"
)

type testFile struct {
	name   string
	data   []byte
	hashes []string
	tags   []*chunk.Tag
}

func generateTags(n int) []*chunk.Tag {
	numTags := rand.Intn(maxTags) + 1
	tags := []*chunk.Tag{}
	tagSize := n / numTags
	for i := 0; i < numTags-1; i++ {
		tags = append(tags, &chunk.Tag{
			Id:        strconv.Itoa(i),
			SizeBytes: int64(tagSize),
		})
	}
	tags = append(tags, &chunk.Tag{
		Id:        strconv.Itoa(numTags - 1),
		SizeBytes: int64(n - (numTags-1)*tagSize),
	})
	return tags
}

func writeFile(t *testing.T, w *Writer, f *testFile, msg string) {
	// Write header.
	hdr := &tar.Header{
		Name: f.name,
		Size: int64(len(f.data)),
	}
	require.NoError(t, w.WriteHeader(hdr), msg)
	// Write content and tags.
	data := f.data
	for _, tag := range f.tags {
		w.Tag(tag.Id)
		_, err := w.Write(data[:tag.SizeBytes])
		require.NoError(t, err, msg)
		data = data[tag.SizeBytes:]
	}
	_, err := w.Write(data)
	require.NoError(t, err, msg)
}

func writeFileSet(t *testing.T, fileSets *Storage, fileSet string, files []*testFile, msg string) {
	w := fileSets.newWriter(context.Background(), fileSet, nil)
	for _, file := range files {
		writeFile(t, w, file, msg)
	}
	require.NoError(t, w.Close(), msg)
}

type fileReader interface {
	Index() *index.Index
	Get(w io.Writer) error
}

func checkFile(t *testing.T, fr fileReader, f *testFile, msg string) {
	// Check header.
	buf := &bytes.Buffer{}
	require.NoError(t, fr.Get(buf), msg)
	tr := tar.NewReader(buf)
	hdr, err := tr.Next()
	require.NoError(t, err, msg)
	require.Equal(t, f.name, hdr.Name, msg)
	require.Equal(t, int64(len(f.data)), hdr.Size, msg)
	// Check content.
	actualData := &bytes.Buffer{}
	_, err = io.Copy(actualData, tr)
	require.NoError(t, err, msg)
	require.Equal(t, 0, bytes.Compare(f.data, actualData.Bytes()), msg)
	// Check hashes.
	actualHashes := dataRefsToHashes(fr.Index().DataOp.DataRefs)
	// If no hashes are recorded then set them based on what was read.
	if len(f.hashes) == 0 {
		f.hashes = actualHashes
	}
	require.Equal(t, f.hashes, actualHashes, msg)
}

func dataRefsToHashes(dataRefs []*chunk.DataRef) []string {
	var hashes []string
	for _, dataRef := range dataRefs {
		if dataRef.Hash == "" {
			hashes = append(hashes, dataRef.ChunkInfo.Chunk.Hash)
			continue
		}
		hashes = append(hashes, dataRef.Hash)
	}
	return hashes
}

// (bryce) this should be somewhere else (probably testutil).
func seedRand() string {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func TestWriteThenRead(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(fileSets *Storage) error {
		msg := seedRand()
		fileNames := index.Generate("abc")
		files := []*testFile{}
		for _, fileName := range fileNames {
			data := chunk.RandSeq(rand.Intn(max))
			files = append(files, &testFile{
				name: fileName,
				data: data,
				tags: generateTags(len(data)),
			})
		}
		// Write out ten filesets where each subsequent fileset has the content of one random file changed.
		// Confirm that all of the content and hashes other than the changed file remain the same.
		fileSet := path.Join(testPath, "0")
		for i := 0; i < 10; i++ {
			// Write the files to the fileset.
			writeFileSet(t, fileSets, fileSet, files, msg)
			// Read the files from the fileset, checking against the recorded files.
			r := fileSets.newReader(context.Background(), fileSet)
			filesIter := files
			require.NoError(t, r.Iterate(func(fr *FileReader) error {
				checkFile(t, fr, filesIter[0], msg)
				filesIter = filesIter[1:]
				return nil
			}), msg)
			// Change one random file
			data := chunk.RandSeq(rand.Intn(max))
			i := rand.Intn(len(files))
			files[i] = &testFile{
				name: files[i].name,
				data: data,
				tags: generateTags(len(data)),
			}
			require.NoError(t, fileSets.Delete(context.Background(), fileSet), msg)
		}
		return nil
	}))
}

func TestCopy(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(fileSets *Storage) error {
		msg := seedRand()
		fileNames := index.Generate("abc")
		files := []*testFile{}
		// Write the initial fileset and count the chunks.
		for _, fileName := range fileNames {
			data := chunk.RandSeq(rand.Intn(max))
			files = append(files, &testFile{
				name: fileName,
				data: data,
				tags: generateTags(len(data)),
			})
		}
		originalPath := path.Join(testPath, "original")
		writeFileSet(t, fileSets, originalPath, files, msg)
		var initialChunkCount int64
		require.NoError(t, fileSets.ChunkStorage().List(context.Background(), func(_ string) error {
			initialChunkCount++
			return nil
		}), msg)
		// Copy intial fileset to a new copy fileset.
		r := fileSets.newReader(context.Background(), originalPath)
		copyPath := path.Join(testPath, "copy")
		wCopy := fileSets.newWriter(context.Background(), copyPath, nil)
		require.NoError(t, r.Iterate(func(fr *FileReader) error {
			return wCopy.CopyFile(fr)
		}), msg)
		require.NoError(t, wCopy.Close(), msg)
		// Compare initial fileset and copy fileset.
		rCopy := fileSets.newReader(context.Background(), copyPath)
		require.NoError(t, rCopy.Iterate(func(fr *FileReader) error {
			checkFile(t, fr, files[0], msg)
			files = files[1:]
			return nil
		}), msg)
		// No new chunks should get created by the copy.
		var finalChunkCount int64
		require.NoError(t, fileSets.ChunkStorage().List(context.Background(), func(_ string) error {
			finalChunkCount++
			return nil
		}), msg)
		require.Equal(t, initialChunkCount, finalChunkCount, msg)
		return nil
	}))
}

func TestResolveIndexes(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(fileSets *Storage) error {
		msg := seedRand()
		numFileSets := 5
		// Generate filesets.
		files := generateFileSets(t, fileSets, numFileSets, testPath, msg)
		// Get the file hashes.
		getHashes(t, fileSets, files, msg)
		// Merge and check the files.
		mr, err := fileSets.NewMergeReader(context.Background(), []string{testPath})
		require.NoError(t, err)
		require.NoError(t, fileSets.ResolveIndexes(context.Background(), []string{testPath}, func(idx *index.Index) error {
			fmr, err := mr.Next()
			require.NoError(t, err)
			fmr.fullIdx = idx
			checkFile(t, fmr, files[0], msg)
			files = files[1:]
			return nil
		}), msg)
		return nil
	}))
}

func TestCompaction(t *testing.T) {
	require.NoError(t, WithLocalStorage(func(fileSets *Storage) error {
		msg := seedRand()
		numFileSets := 5
		// Generate filesets.
		files := generateFileSets(t, fileSets, numFileSets, testPath, msg)
		// Get the file hashes.
		getHashes(t, fileSets, files, msg)
		// Compact the files.
		require.NoError(t, fileSets.Compact(context.Background(), path.Join(testPath, Compacted), []string{testPath}), msg)
		// Check the files.
		r := fileSets.newReader(context.Background(), path.Join(testPath, Compacted))
		require.NoError(t, r.Iterate(func(fr *FileReader) error {
			checkFile(t, fr, files[0], msg)
			files = files[1:]
			return nil
		}), msg)
		return nil
	}))
}

func generateFileSets(t *testing.T, fileSets *Storage, numFileSets int, prefix, msg string) []*testFile {
	fileNames := index.Generate("abcd")
	files := []*testFile{}
	// Generate the files and randomly distribute them across the filesets.
	var ws []*Writer
	for i := 0; i < numFileSets; i++ {
		ws = append(ws, fileSets.newWriter(context.Background(), path.Join(prefix, strconv.Itoa(i)), nil))
	}
	for i, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		files = append(files, &testFile{
			name: fileName,
			data: data,
			tags: generateTags(len(data)),
		})
		// Shallow copy for slicing as data is distributed.
		f := *files[i]
		wsCopy := make([]*Writer, len(ws))
		copy(wsCopy, ws)
		// Randomly distribute tagged data among filesets.
		for len(f.tags) > 0 {
			// Randomly select fileset to write to.
			i := rand.Intn(len(wsCopy))
			w := wsCopy[i]
			wsCopy = append(wsCopy[:i], wsCopy[i+1:]...)
			// Write the rest of the file if this is the last fileset.
			if len(wsCopy) == 0 {
				writeFile(t, w, &f, msg)
				break
			}
			// Choose a random number of the tags left.
			numTags := rand.Intn(len(f.tags)) + 1
			var size int
			for _, tag := range f.tags[:numTags] {
				size += int(tag.SizeBytes)
			}
			// Create file for writing and remove data/tags from rest of the file.
			fWrite := f
			fWrite.data = fWrite.data[:size]
			fWrite.tags = fWrite.tags[:numTags]
			f.data = f.data[size:]
			f.tags = f.tags[numTags:]
			writeFile(t, w, &fWrite, msg)
		}
	}
	for _, w := range ws {
		require.NoError(t, w.Close(), msg)
	}
	return files
}

func getHashes(t *testing.T, fileSets *Storage, files []*testFile, msg string) {
	writeFileSet(t, fileSets, path.Join(scratchPath, Compacted), files, msg)
	r := fileSets.newReader(context.Background(), path.Join(scratchPath, Compacted))
	require.NoError(t, r.Iterate(func(fr *FileReader) error {
		files[0].hashes = dataRefsToHashes(fr.Index().DataOp.DataRefs)
		files = files[1:]
		return nil
	}), msg)
}
