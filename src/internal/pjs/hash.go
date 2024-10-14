package pjs

import (
	"archive/tar"
	"bytes"
	"context"
	"hash"
	"io"
	"io/fs"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/storage"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
)

// HashFS returns a stable hash of the contents of fileSystem.
func HashFS(fileSystem fs.FS) ([]byte, error) {
	var h hasher
	if err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) (retErr error) {
		if err != nil {
			return errors.Wrapf(err, "error walking %s", path)
		}
		if d.IsDir() {
			return nil
		}
		i, err := d.Info()
		if err != nil {
			return errors.Wrapf(err, "info for %s", path)
		}
		f, err := fileSystem.Open(path)
		if err != nil {
			return errors.Wrapf(err, "open %s", path)
		}
		defer errors.Close(&retErr, f, "closing %s", path)
		if err := h.writeFile("/"+path, i.Size(), f); err != nil {
			return errors.Wrapf(err, "hashing %s", path)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "walking testFS")
	}
	return h.sum()
}

// A FilesetReader provides a store.Fileset_ReadFilesetClient for hashing.
type FilesetReader interface {
	ReadFileset(context.Context, *storage.ReadFilesetRequest, ...grpc.CallOption) (storage.Fileset_ReadFilesetClient, error)
}

// HashFileset hashes the fileset identified by handle as read from a FilesetReader.
func HashFileset(ctx context.Context, fr FilesetReader, handle string) ([]byte, error) {
	var h hasher
	cc, err := fr.ReadFileset(ctx, &storage.ReadFilesetRequest{
		FilesetId: handle,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "read fileset %s", handle)
	}
	for {
		r, err := cc.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrap(err, "reading fileset")
		}
		if err := h.writeFile(r.Path, int64(len(r.Data.Value)), bytes.NewReader(r.Data.Value)); err != nil {
			return nil, errors.Wrapf(err, "hashing %s", r.Path)
		}
	}
	return h.sum()
}

type hasher struct {
	w *tar.Writer
	h hash.Hash
}

func (h *hasher) ensure() {
	if h.h == nil {
		h.h = pachhash.New()
	}
	if h.w == nil {
		h.w = tar.NewWriter(h.h)
	}
}

func (h *hasher) writeFile(path string, size int64, r io.Reader) error {
	h.ensure()
	if err := h.w.WriteHeader(tarutil.NewHeader(path, size)); err != nil {
		return errors.Wrapf(err, "writing tar header for %s", path)
	}
	if _, err := io.Copy(h.w, r); err != nil {
		return errors.Wrapf(err, "writing %s", path)
	}
	return nil
}

func (h *hasher) sum() ([]byte, error) {
	h.ensure()
	if err := h.w.Close(); err != nil {
		return nil, errors.Wrap(err, "closing tar writer")
	}
	return h.h.Sum(nil), nil
}
