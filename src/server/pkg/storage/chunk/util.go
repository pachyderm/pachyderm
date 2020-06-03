package chunk

import (
	"math/rand"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// WithLocalStorage creates a local storage instance for testing during the lifetime of
// the callback.
func WithLocalStorage(f func(obj.Client, *Storage) error, opts ...StorageOption) error {
	return obj.WithLocalClient(func(objClient obj.Client) error {
		return f(objClient, NewStorage(objClient, opts...))
	})
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandSeq generates a random sequence of data (n is number of bytes)
func RandSeq(n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return []byte(string(b))
}

// Reference creates a data reference for the full chunk referenced by a data reference.
func Reference(dataRef *DataRef, tag string) *DataRef {
	chunkRef := &DataRef{}
	chunkRef.ChunkInfo = dataRef.ChunkInfo
	chunkRef.SizeBytes = dataRef.ChunkInfo.SizeBytes
	chunkRef.Tags = []*Tag{
		&Tag{
			Id:        tag,
			SizeBytes: dataRef.ChunkInfo.SizeBytes,
		},
	}
	return chunkRef
}

func joinTags(ts1, ts2 []*Tag) []*Tag {
	if ts1 != nil {
		lastT := ts1[len(ts1)-1]
		if lastT.Id == ts2[0].Id {
			lastT.SizeBytes += ts2[0].SizeBytes
			ts2 = ts2[1:]
		}
	}
	return append(ts1, ts2...)
}
