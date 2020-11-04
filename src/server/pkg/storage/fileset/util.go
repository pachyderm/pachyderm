package fileset

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/tracker"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
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

// CopyFiles iterates over s and writes all the Files to w
func CopyFiles(ctx context.Context, w *Writer, s FileSet) error {
	switch s := s.(type) {
	case *Reader:
		return s.iterate(func(fr *FileReader) error {
			return w.CopyFile(fr)
		})
	default:
		return s.Iterate(ctx, func(f File) error {
			hdr, err := f.Header()
			if err != nil {
				return err
			}
			if err := w.WriteHeader(hdr); err != nil {
				return err
			}
			return f.Content(w)
		})
	}
}

// WriteTarEntry writes an tar entry for f to w
func WriteTarEntry(w io.Writer, f File) error {
	h, err := f.Header()
	if err != nil {
		return err
	}
	tw := tar.NewWriter(w)
	if err := tw.WriteHeader(h); err != nil {
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

// CleanTarPath ensures that the path is in the canonical format for tar header names.
// This includes ensuring a prepending /'s and ensure directory paths
// have a trailing slash.
func CleanTarPath(x string, isDir bool) string {
	y := "/" + strings.Trim(x, "/")
	if isDir && !IsDir(y) {
		y += "/"
	}
	return y
}

// IsCleanTarPath determines if the path is a valid tar path.
func IsCleanTarPath(x string, isDir bool) bool {
	y := CleanTarPath(x, isDir)
	return y == x
}

func getSortedKeys(set map[string]struct{}) []string {
	keys := make([]string, len(set))
	var i int
	for key := range set {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}

// IsDir determines if a path is for a directory.
func IsDir(p string) bool {
	return strings.HasSuffix(p, "/")
}

// DirUpperBound returns the immediate next path after a directory in a lexicographical ordering.
func DirUpperBound(p string) string {
	if !IsDir(p) {
		panic(fmt.Sprintf("%v is not a directory path", p))
	}
	return strings.TrimRight(p, "/") + "0"
}
