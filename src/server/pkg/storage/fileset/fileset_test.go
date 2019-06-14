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
	testPath = "test"
)

type file struct {
	hashes []string
	data   []byte
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
	for _, fileName := range fileNames {
		files[fileName] = &file{
			data: chunk.RandSeq(rand.Intn(max)),
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
			hdr := &index.Header{
				Hdr: &tar.Header{
					Name: fileName,
					Size: int64(len(files[fileName].data)),
				},
			}
			require.NoError(t, w.WriteHeader(hdr), seedStr(seed))
			_, err := w.Write(files[fileName].data)
			require.NoError(t, err, seedStr(seed))
		}
		require.NoError(t, w.Close(), seedStr(seed))
		// Read files from file set, checking against recorded data and hashes.
		r := fileSets.NewReader(context.Background(), testPath, "")
		for _, fileName := range fileNames {
			hdr, err := r.Next()
			require.NoError(t, err, seedStr(seed))
			actualHashes := dataRefsToHashes(hdr.Idx.DataOp.DataRefs)
			// If no hashes are recorded (first iteration or changed file),
			// then set them based on what was read.
			if len(files[fileName].hashes) == 0 {
				files[fileName].hashes = actualHashes
			}
			require.Equal(t, files[fileName].hashes, actualHashes, seedStr(seed))
			actualData := &bytes.Buffer{}
			_, err = io.Copy(actualData, r)
			require.NoError(t, err, seedStr(seed))
			require.Equal(t, files[fileName].data, actualData.Bytes(), seedStr(seed))
		}
		// Change one random file
		for fileName := range files {
			files[fileName] = &file{
				data: chunk.RandSeq(rand.Intn(max)),
			}
			break
		}
		require.NoError(t, chunks.DeleteAll(context.Background()), seedStr(seed))
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
		require.NoError(t, w.WriteHeader(hdr), seedStr(seed))
		_, err := w.Write(data)
		require.NoError(t, err, seedStr(seed))
	}
	require.NoError(t, w.Close(), seedStr(seed))
	var initialChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		initialChunkCount++
		return nil
	}), seedStr(seed))
	// Copy intial file set to a new copy file set.
	r := fileSets.NewReader(context.Background(), testPath, "")
	wCopy := fileSets.NewWriter(context.Background(), testPath+"Copy")
	for _ = range fileNames {
		hdr, err := r.Next()
		require.NoError(t, err, seedStr(seed))
		hdr.Idx = nil
		require.NoError(t, wCopy.WriteHeader(hdr), seedStr(seed))
		mid := hdr.Hdr.Size / 2
		require.NoError(t, CopyN(wCopy, r, mid), seedStr(seed))
		require.NoError(t, CopyN(wCopy, r, hdr.Hdr.Size-mid), seedStr(seed))
	}
	require.NoError(t, wCopy.Close(), seedStr(seed))
	// Compare initial file set and copy file set.
	r = fileSets.NewReader(context.Background(), testPath, "")
	rCopy := fileSets.NewReader(context.Background(), testPath+"Copy", "")
	for _, fileName := range fileNames {
		hdr, err := r.Next()
		require.NoError(t, err, seedStr(seed))
		require.Equal(t, fileName, hdr.Hdr.Name, seedStr(seed))
		hdrCopy, err := rCopy.Next()
		require.NoError(t, err, seedStr(seed))
		require.Equal(t, hdr.Hdr, hdrCopy.Hdr, seedStr(seed))
		rData := &bytes.Buffer{}
		_, err = io.Copy(rData, r)
		require.NoError(t, err, seedStr(seed))
		rCopyData := &bytes.Buffer{}
		_, err = io.Copy(rCopyData, rCopy)
		require.NoError(t, err, seedStr(seed))
		require.Equal(t, rData.Bytes(), rCopyData.Bytes(), seedStr(seed))
	}
	// No new chunks should get created by the copy.
	var finalChunkCount int64
	require.NoError(t, chunks.List(context.Background(), func(_ string) error {
		finalChunkCount++
		return nil
	}), seedStr(seed))
	require.Equal(t, initialChunkCount, finalChunkCount, seedStr(seed))
}
