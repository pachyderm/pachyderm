package fileset

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

// WithTestStorage constructs a local storage instance for testing during the lifetime of
// the callback.
func WithTestStorage(t testing.TB, f func(*Storage) error) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		chunk.SetupPostgresStore(db)
		tracker.PGTrackerApplySchema(db)
		SetupPostgresStore(db)
		tr := tracker.NewPGTracker(db)
		obj.WithLocalClient(func(objC obj.Client) error {
			chunkStorage := chunk.NewStorage(objC, chunk.NewPostgresStore(db), tr)
			return f(NewStorage(NewPostgresStore(db), tr, chunkStorage))
		})
	})
}

// CopyFiles copies files from a file set to a file set writer.
func CopyFiles(ctx context.Context, w *Writer, fs FileSet, deletive ...bool) error {
	if len(deletive) > 0 && deletive[0] {
		if err := fs.Iterate(ctx, func(f File) error {
			return deleteIndex(w, f.Index())
		}, deletive...); err != nil {
			return err
		}
	}
	return fs.Iterate(ctx, func(f File) error {
		return w.Copy(f)
	})
}

func deleteIndex(w *Writer, idx *index.Index) error {
	p := idx.Path
	if len(idx.File.Parts) == 0 {
		return w.Delete(p)
	}
	var tags []string
	for _, part := range idx.File.Parts {
		tags = append(tags, part.Tag)
	}
	return w.Delete(p, tags...)
}

// WriteTarEntry writes an tar entry for f to w
func WriteTarEntry(w io.Writer, f File) error {
	idx := f.Index()
	tw := tar.NewWriter(w)
	if err := tw.WriteHeader(tarutil.NewHeader(idx.Path, idx.SizeBytes)); err != nil {
		return err
	}
	if err := f.Content(tw); err != nil {
		return err
	}
	return tw.Flush()
}

// WriteTarStream writes an entire tar stream to w
// It will contain an entry for each File in fs
func WriteTarStream(ctx context.Context, w io.Writer, fs FileSet) error {
	if err := fs.Iterate(ctx, func(f File) error {
		return WriteTarEntry(w, f)
	}); err != nil {
		return err
	}
	return tar.NewWriter(w).Close()
}

func Clean(x string, isDir bool) string {
	y := "/" + strings.Trim(x, "/")
	if isDir && !IsDir(y) {
		y += "/"
	}
	return y
}

func IsClean(x string, isDir bool) bool {
	y := Clean(x, isDir)
	return y == x
}

// IsDir determines if a path is for a directory.
func IsDir(p string) bool {
	return strings.HasSuffix(p, "/")
}

// Iterator provides functionality for imperative iteration over a file set.
type Iterator struct {
	peek     File
	fileChan chan File
	errChan  chan error
}

// NewIterator creates a new iterator.
func NewIterator(ctx context.Context, fs FileSet, deletive ...bool) *Iterator {
	fileChan := make(chan File)
	errChan := make(chan error, 1)
	go func() {
		if err := fs.Iterate(ctx, func(f File) error {
			fileChan <- f
			return nil
		}, deletive...); err != nil {
			errChan <- err
			return
		}
		close(fileChan)
	}()
	return &Iterator{
		fileChan: fileChan,
		errChan:  errChan,
	}
}

// Peek returns the next file without progressing the iterator.
func (i *Iterator) Peek() (File, error) {
	if i.peek != nil {
		return i.peek, nil
	}
	var err error
	i.peek, err = i.Next()
	return i.peek, err
}

// Next returns the next file and progresses the iterator.
func (i *Iterator) Next() (File, error) {
	if i.peek != nil {
		tmp := i.peek
		i.peek = nil
		return tmp, nil
	}
	select {
	case file, more := <-i.fileChan:
		if !more {
			return nil, io.EOF
		}
		return file, nil
	case err := <-i.errChan:
		return nil, err
	}
}

func getDataRefs(parts []*index.Part) []*chunk.DataRef {
	var dataRefs []*chunk.DataRef
	for _, part := range parts {
		dataRefs = append(dataRefs, part.DataRefs...)
	}
	return dataRefs
}
