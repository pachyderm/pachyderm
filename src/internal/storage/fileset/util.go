package fileset

import (
	"archive/tar"
	"context"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/track"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
)

// NewTestStorage constructs a local storage instance scoped to the lifetime of the test
func NewTestStorage(ctx context.Context, t testing.TB, db *pachsql.DB, tr track.Tracker, opts ...StorageOption) *Storage {
	_, chunks := chunk.NewTestStorage(ctx, t, db, tr)
	store := NewTestStore(ctx, t, db)
	return NewStorage(store, tr, chunks, opts...)
}

// CopyFiles copies files from a file set to a file set writer.
func CopyFiles(ctx context.Context, w *Writer, fs FileSet, opts ...index.Option) error {
	return errors.EnsureStack(fs.Iterate(ctx, func(f File) error {
		return w.Copy(f, f.Index().File.Datum)
	}, opts...))
}

// CopyDeletedFiles copies the deleted files from a file set to a file set writer.
func CopyDeletedFiles(ctx context.Context, w *Writer, fs FileSet, opts ...index.Option) error {
	return errors.EnsureStack(fs.IterateDeletes(ctx, func(f File) error {
		idx := f.Index()
		return w.Delete(idx.Path, idx.File.Datum)
	}, opts...))
}

// WriteTarEntry writes an tar entry for f to w
func WriteTarEntry(ctx context.Context, w io.Writer, f File) error {
	idx := f.Index()
	tw := tar.NewWriter(w)
	if err := tw.WriteHeader(tarutil.NewHeader(idx.Path, index.SizeBytes(idx))); err != nil {
		return errors.EnsureStack(err)
	}
	if err := f.Content(ctx, tw); err != nil {
		return err
	}
	return errors.EnsureStack(tw.Flush())
}

// WriteTarStream writes an entire tar stream to w
// It will contain an entry for each File in fs
func WriteTarStream(ctx context.Context, w io.Writer, fs FileSet) error {
	if err := fs.Iterate(ctx, func(f File) error {
		return WriteTarEntry(ctx, w, f)
	}); err != nil {
		return err
	}
	return errors.EnsureStack(tar.NewWriter(w).Close())
}

// Clean cleans a file path.
func Clean(p string, isDir bool) string {
	p = path.Clean(p)
	if p == "." {
		return "/"
	}
	y := "/" + strings.Trim(p, "/")
	if isDir && !IsDir(y) {
		y += "/"
	}
	return y
}

// IsClean checks if a file path is clean.
func IsClean(x string, isDir bool) bool {
	y := Clean(x, isDir)
	return y == x
}

// IsDir determines if a path is for a directory.
func IsDir(p string) bool {
	return strings.HasSuffix(p, "/")
}

func computeFileHash(hashes [][]byte) ([]byte, error) {
	h := pachhash.New()
	for _, hash := range hashes {
		_, err := h.Write(hash)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	return h.Sum(nil), nil
}

func SizeFromIndex(idx *index.Index) (size int64) {
	for _, dr := range idx.File.DataRefs {
		size += dr.SizeBytes
	}
	return size
}
