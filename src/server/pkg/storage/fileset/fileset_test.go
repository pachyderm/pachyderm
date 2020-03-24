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

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

const (
	max      = 20 * chunk.MB
	maxTags  = 10
	testPath = "test"
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

func writeFile(t *testing.T, w *Writer, name string, f *testFile, msg string) {
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

func checkFile(t *testing.T, fr *FileReader, f *testFile, msg string) {
	// Check header.
	buf := &bytes.Buffer{}
	require.NoError(t, fr.Get(buf))
	tr := tar.NewReader(buf)
	hdr, err := tr.Next()
	require.NoError(t, err)
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

func seedStr(seed int64) string {
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

func TestWriteThenRead(t *testing.T) {
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunk.Cleanup(objC, chunks)
		objC.Delete(context.Background(), path.Join(prefix, testPath))
		objC.Delete(context.Background(), prefix)
	}()
	fileSets := NewStorage(objC, chunks)
	fileNames := index.Generate("abc")
	files := make(map[string]*testFile)
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	for _, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		files[fileName] = &testFile{
			name: fileName,
			data: data,
			tags: generateTags(len(data)),
		}
	}
	// Write out ten filesets where each subsequent fileset has the content of one random file changed.
	// Confirm that all of the content and hashes other than the changed file remain the same.
	for i := 0; i < 10; i++ {
		// Write files to fileset.
		w := fileSets.newWriter(context.Background(), testPath)
		for _, fileName := range fileNames {
			writeFile(t, w, fileName, files[fileName], msg)
		}
		require.NoError(t, w.Close(), msg)
		// Read files from fileset, checking against recorded files.
		r := fileSets.newReader(context.Background(), testPath)
		require.NoError(t, r.Iterate(func(fr *FileReader) error {
			checkFile(t, fr, files[fr.Index().Path], msg)
			return nil
		}))
		// Change one random file
		for fileName := range files {
			data := chunk.RandSeq(rand.Intn(max))
			files[fileName] = &testFile{
				name: fileName,
				data: data,
				tags: generateTags(len(data)),
			}
			break
		}
		require.NoError(t, chunks.DeleteAll(context.Background()), msg)
	}
}

func TestCopy(t *testing.T) {
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunk.Cleanup(objC, chunks)
		objC.Delete(context.Background(), path.Join(prefix, testPath))
		objC.Delete(context.Background(), path.Join(prefix, testPath+"Copy"))
		objC.Delete(context.Background(), prefix)
	}()
	fileSets := NewStorage(objC, chunks)
	fileNames := index.Generate("abc")
	files := make(map[string]*testFile)
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	// Write the initial fileset and count the chunks.
	w := fileSets.newWriter(context.Background(), testPath)
	for _, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		files[fileName] = &testFile{
			name: fileName,
			data: data,
			tags: generateTags(len(data)),
		}
		writeFile(t, w, fileName, files[fileName], msg)
	}
	require.NoError(t, w.Close(), msg)
	var initialChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		initialChunkCount++
		return nil
	}), msg)
	// Copy intial fileset to a new copy fileset.
	testPathCopy := testPath + "Copy"
	r := fileSets.newReader(context.Background(), testPath)
	wCopy := fileSets.newWriter(context.Background(), testPathCopy)
	require.NoError(t, r.Iterate(func(fr *FileReader) error {
		return wCopy.CopyFile(fr)
	}))
	require.NoError(t, wCopy.Close(), msg)
	// Compare initial fileset and copy fileset.
	rCopy := fileSets.newReader(context.Background(), testPathCopy)
	require.NoError(t, rCopy.Iterate(func(fr *FileReader) error {
		checkFile(t, fr, files[fr.Index().Path], msg)
		return nil
	}))
	// No new chunks should get created by the copy.
	var finalChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		finalChunkCount++
		return nil
	}), msg)
	require.Equal(t, initialChunkCount, finalChunkCount, msg)
}

func TestMergeReader(t *testing.T) {
	// Setup storage and seed.
	numFileSets := 5
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunk.Cleanup(objC, chunks)
		for i := 0; i < numFileSets; i++ {
			objC.Delete(context.Background(), path.Join(prefix, testPath+strconv.Itoa(i)))
		}
		objC.Delete(context.Background(), path.Join(prefix, testPath))
		objC.Delete(context.Background(), prefix)
	}()
	fileSets := NewStorage(objC, chunks)
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	// Generate filesets.
	files := generateFileSets(t, fileSets, numFileSets, testPath, msg)
	// Merge and check the files.
	mr, err := fileSets.NewMergeReader(context.Background(), []string{testPath})
	require.NoError(t, err)
	require.NoError(t, mr.Iterate(func(fmr *FileMergeReader) error {
		actualData := &bytes.Buffer{}
		tsmr, err := fmr.TagSetMergeReader()
		require.NoError(t, err)
		if err := tsmr.Get(actualData); err != nil {
			return err
		}
		require.Equal(t, 0, bytes.Compare(files[fmr.Index().Path].data, actualData.Bytes()), msg)
		return nil
	}))
}

func TestCompaction(t *testing.T) {
	// Setup storage and seed.
	numFileSets := 5
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunk.Cleanup(objC, chunks)
		for i := 0; i < numFileSets; i++ {
			objC.Delete(context.Background(), path.Join(prefix, testPath+strconv.Itoa(i)))
		}
		objC.Delete(context.Background(), path.Join(prefix, testPath, Compacted))
		objC.Delete(context.Background(), path.Join(prefix, testPath))
		objC.Delete(context.Background(), prefix)
	}()
	fileSets := NewStorage(objC, chunks)
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	// Generate filesets.
	files := generateFileSets(t, fileSets, numFileSets, testPath, msg)
	// Compact the files.
	require.NoError(t, fileSets.Compact(context.Background(), path.Join(testPath, Compacted), []string{testPath}), msg)
	// Check the files.
	r := fileSets.newReader(context.Background(), path.Join(testPath, Compacted))
	require.NoError(t, r.Iterate(func(fr *FileReader) error {
		checkFile(t, fr, files[fr.Index().Path], msg)
		return nil
	}), msg)
}

func generateFileSets(t *testing.T, fileSets *Storage, numFileSets int, prefix, msg string) map[string]*testFile {
	fileNames := index.Generate("abcd")
	files := make(map[string]*testFile)
	// Generate the files and randomly distribute them across the filesets.
	var ws []*Writer
	for i := 0; i < numFileSets; i++ {
		ws = append(ws, fileSets.newWriter(context.Background(), prefix+strconv.Itoa(i)))
	}
	for _, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		files[fileName] = &testFile{
			name: fileName,
			data: data,
			tags: generateTags(len(data)),
		}
		// Shallow copy for slicing as data is distributed.
		f := *files[fileName]
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
				writeFile(t, w, fileName, &f, msg)
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
			writeFile(t, w, fileName, &fWrite, msg)
		}
	}
	for _, w := range ws {
		require.NoError(t, w.Close(), msg)
	}
	return files
}
