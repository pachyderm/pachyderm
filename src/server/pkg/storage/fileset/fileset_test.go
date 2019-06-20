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

type file struct {
	hashes []string
	data   []byte
	tags   []*index.Tag
}

func generateTags(n int) []*index.Tag {
	numTags := rand.Intn(maxTags) + 1
	tags := []*index.Tag{}
	tagSize := n / numTags
	for i := 0; i < numTags-1; i++ {
		tags = append(tags, &index.Tag{
			Id:        strconv.Itoa(i),
			SizeBytes: int64(tagSize),
		})
	}
	tags = append(tags, &index.Tag{
		Id:        strconv.Itoa(numTags - 1),
		SizeBytes: int64(n - (numTags-1)*tagSize),
	})
	return tags
}

func writeFile(t *testing.T, w *Writer, name string, f *file, msg string) {
	// Write header.
	hdr := &index.Header{
		Hdr: &tar.Header{
			Name: name,
			Size: int64(len(f.data)),
		},
	}
	require.NoError(t, w.WriteHeader(hdr), msg)
	// Write data with tags.
	data := f.data
	for _, tag := range f.tags {
		w.StartTag(tag.Id)
		_, err := w.Write(data[:tag.SizeBytes])
		require.NoError(t, err, msg)
		data = data[tag.SizeBytes:]
	}
	_, err := w.Write(data)
	require.NoError(t, err, msg)
}

func checkNextFile(t *testing.T, r *Reader, f *file, msg string) {
	hdr, err := r.Next()
	require.NoError(t, err, msg)
	actualHashes := dataRefsToHashes(hdr.Idx.DataOp.DataRefs)
	// If no hashes are recorded then set them based on what was read.
	if len(f.hashes) == 0 {
		f.hashes = actualHashes
	}
	require.Equal(t, f.hashes, actualHashes, msg)
	actualData := &bytes.Buffer{}
	_, err = io.Copy(actualData, r)
	require.NoError(t, err, msg)
	require.Equal(t, f.data, actualData.Bytes(), msg)
	// Slice of tags that excludes the first element is necessary to remove the header tag.
	// (bryce) should the header tag be exposed?
	require.Equal(t, f.tags, hdr.Idx.DataOp.Tags[1:])
}

func dataRefsToHashes(dataRefs []*chunk.DataRef) []string {
	var hashes []string
	for _, dataRef := range dataRefs {
		if dataRef.Hash == "" {
			hashes = append(hashes, dataRef.Chunk.Hash)
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
	files := make(map[string]*file)
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	for _, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		files[fileName] = &file{
			data: data,
			tags: generateTags(len(data)),
		}
	}
	// Write out ten filesets where each subsequent fileset has the content of one random file changed.
	// Confirm that all of the content and hashes other than the changed file remain the same.
	// (bryce) we are going to want a dedupe test somewhere, not sure if it makes sense here or in the chunk
	// storage layer (probably in the chunk storage layer).
	for i := 0; i < 10; i++ {
		// Write files to file set.
		w := fileSets.NewWriter(context.Background(), testPath)
		for _, fileName := range fileNames {
			writeFile(t, w, fileName, files[fileName], msg)
		}
		require.NoError(t, w.Close(), msg)
		// Read files from file set, checking against recorded files.
		r := fileSets.NewReader(context.Background(), testPath, "")
		for _, fileName := range fileNames {
			checkNextFile(t, r, files[fileName], msg)
		}
		// Change one random file
		for fileName := range files {
			data := chunk.RandSeq(rand.Intn(max))
			files[fileName] = &file{
				data: data,
				tags: generateTags(len(data)),
			}
			break
		}
		require.NoError(t, chunks.DeleteAll(context.Background()), msg)
	}
}

// (bryce) there is a lot of common functionality between this and
// the WriteThenRead test. This common functionality should get refactored
// into helper functions.
func TestCopyN(t *testing.T) {
	objC, chunks := chunk.LocalStorage(t)
	defer func() {
		chunk.Cleanup(objC, chunks)
		objC.Delete(context.Background(), path.Join(prefix, testPath))
		objC.Delete(context.Background(), path.Join(prefix, testPath+"Copy"))
		objC.Delete(context.Background(), prefix)
	}()
	fileSets := NewStorage(objC, chunks)
	fileNames := index.Generate("abc")
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	msg := seedStr(seed)
	// Write the initial file set and count the chunks.
	w := fileSets.NewWriter(context.Background(), testPath)
	for _, fileName := range fileNames {
		data := chunk.RandSeq(rand.Intn(max))
		hdr := &index.Header{
			Hdr: &tar.Header{
				Name: fileName,
				Size: int64(len(data)),
			},
		}
		require.NoError(t, w.WriteHeader(hdr), msg)
		_, err := w.Write(data)
		require.NoError(t, err, msg)
	}
	require.NoError(t, w.Close(), msg)
	var initialChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		initialChunkCount++
		return nil
	}), msg)
	// Copy intial file set to a new copy file set.
	testPathCopy := testPath + "Copy"
	r := fileSets.NewReader(context.Background(), testPath, "")
	wCopy := fileSets.NewWriter(context.Background(), testPathCopy)
	for range fileNames {
		hdr, err := r.Next()
		require.NoError(t, err, msg)
		require.NoError(t, wCopy.WriteHeader(hdr), msg)
		mid := hdr.Hdr.Size / 2
		require.NoError(t, CopyN(wCopy, r, mid), msg)
		require.NoError(t, CopyN(wCopy, r, hdr.Hdr.Size-mid), msg)
	}
	require.NoError(t, wCopy.Close(), msg)
	// Compare initial file set and copy file set.
	r = fileSets.NewReader(context.Background(), testPath, "")
	rCopy := fileSets.NewReader(context.Background(), testPathCopy, "")
	for _, fileName := range fileNames {
		hdr, err := r.Next()
		require.NoError(t, err, msg)
		require.Equal(t, fileName, hdr.Hdr.Name, msg)
		hdrCopy, err := rCopy.Next()
		require.NoError(t, err, msg)
		require.Equal(t, hdr.Hdr, hdrCopy.Hdr, msg)
		rData := &bytes.Buffer{}
		_, err = io.Copy(rData, r)
		require.NoError(t, err, msg)
		rDataCopy := &bytes.Buffer{}
		_, err = io.Copy(rDataCopy, rCopy)
		require.NoError(t, err, msg)
		require.Equal(t, rData.Bytes(), rDataCopy.Bytes(), msg)
	}
	// No new chunks should get created by the copy.
	var finalChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		finalChunkCount++
		return nil
	}), msg)
	require.Equal(t, initialChunkCount, finalChunkCount, msg)
}
