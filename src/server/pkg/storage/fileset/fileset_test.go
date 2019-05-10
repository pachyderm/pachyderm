package fileset

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

const (
	max = 20 * chunk.MB
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
		chunks.DeleteAll(context.Background())
		objC.Delete(context.Background(), chunk.Prefix)
	}()
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
		w := NewWriter(context.Background(), chunks)
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
		idx, err := w.Close()
		require.NoError(t, err, seedStr(seed))
		// Read files from file set, checking against recorded data and hashes.
		r := NewReader(context.Background(), chunks, idx, "")
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
