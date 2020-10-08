package fileset

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

// WithLocalStorage constructs a local storage instance for testing during the lifetime of
// the callback.
func WithLocalStorage(t *testing.T, f func(*Storage) error) (err error) {
	WithTestPathStore(t, func(store PathStore) {
		err = chunk.WithLocalStorage(func(objC obj.Client, chunks *chunk.Storage) error {
			return f(NewStorage(store, chunks))
		})
	})
	return err
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
	if isDir && !strings.HasSuffix(y, "/") {
		y += "/"
	}
	return y
}

// IsCleanTarPath determines if the path is a valid tar path.
func IsCleanTarPath(x string, isDir bool) bool {
	y := CleanTarPath(x, isDir)
	return y == x
}
