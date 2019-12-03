package chunk

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	// KB is Kilobyte.
	KB = 1024
	// MB is Megabyte.
	MB = 1024 * KB
)

// LocalStorage creates a local chunk storage instance.
// Useful for storage layer tests.
func LocalStorage(tb testing.TB) (obj.Client, *Storage) {
	wd, err := os.Getwd()
	require.NoError(tb, err)
	objC, err := obj.NewLocalClient(wd)
	require.NoError(tb, err)
	return objC, NewStorage(objC)
}

// Cleanup cleans up a local chunk storage instance.
func Cleanup(objC obj.Client, chunks *Storage) {
	chunks.DeleteAll(context.Background())
	objC.Delete(context.Background(), prefix)
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

func joinAnnotations(as []*Annotation, a *Annotation) []*Annotation {
	// If the annotation being added is the same as the
	// last, then they are merged.
	if as != nil {
		lastA := as[len(as)-1]
		if lastA.Meta == a.Meta {
			lastA.buf.Write(a.buf.Bytes())
			if lastA.tags != nil {
				lastA.tags = joinTags(lastA.tags, a.tags)
			}
			return as
		}
	}
	return append(as, a)
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

func splitAnnotation(a *Annotation, size int) (*Annotation, *Annotation) {
	a1 := copyAnnotation(a)
	a2 := copyAnnotation(a)
	if a.buf != nil {
		a1.buf = bytes.NewBuffer(a.buf.Bytes()[:size])
		a2.buf = bytes.NewBuffer(a.buf.Bytes()[size:])
	}
	if a.tags != nil {
		a1.tags, a2.tags = splitTags(a.tags, size)
	}
	return a1, a2
}

func copyAnnotation(a *Annotation) *Annotation {
	copyA := &Annotation{Meta: a.Meta}
	if a.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	if a.buf != nil {
		copyA.buf = &bytes.Buffer{}
	}
	return copyA
}

func splitTags(ts []*Tag, size int) ([]*Tag, []*Tag) {
	var ts1, ts2 []*Tag
	for _, t := range ts {
		ts2 = append(ts2, proto.Clone(t).(*Tag))
	}
	for {
		if int(ts2[0].SizeBytes) >= size {
			t := proto.Clone(ts2[0]).(*Tag)
			t.SizeBytes = int64(size)
			ts1 = append(ts1, t)
			ts2[0].SizeBytes -= int64(size)
			break
		}
		size -= int(ts2[0].SizeBytes)
		ts1 = append(ts1, ts2[0])
		ts2 = ts2[1:]
	}
	return ts1, ts2
}
