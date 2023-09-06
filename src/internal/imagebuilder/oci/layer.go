package oci

import (
	"archive/tar"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/klauspost/compress/zstd"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const LayerMediaType = "application/vnd.oci.image.layer.v1.tar+zstd"

type Layer struct {
	// Path to the underlying file.
	Underlying string
	// OCI Layer descriptor.
	Descriptor v1.Descriptor
}

type countWriter struct{ n int }

func (w *countWriter) Write(b []byte) (int, error) {
	w.n += len(b)
	return len(b), nil
}

func sha256Hex(b []byte) digest.Digest {
	// The digest.Digest library is crazy.
	return digest.Digest(fmt.Sprintf("sha256:%x", b))
}

func NewLayerFromFS(f fs.FS) (layer *Layer, retErr error) {
	outfh, err := os.CreateTemp("", "layer-*.tar.zst")
	if err != nil {
		return nil, errors.Wrap(err, "create underlying file")
	}
	sha := sha256.New()
	cw := new(countWriter)
	out := io.MultiWriter(outfh, sha, cw)
	var outputOK bool
	defer func() {
		if layer != nil {
			layer.Descriptor.Size = int64(cw.n)
			layer.Descriptor.Digest = sha256Hex(sha.Sum(nil))
		}
		if !outputOK {
			errors.JoinInto(&retErr, errors.Wrap(os.Remove(outfh.Name()), "remove partial output file"))
		}
	}()
	defer errors.Close(&retErr, outfh, "close underlying file")

	zst, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, errors.Wrap(err, "create zstd writer")
	}
	defer errors.Close(&retErr, zst, "close zstd writer")

	tw := tar.NewWriter(zst)
	defer errors.Close(&retErr, tw, "close tar writer")

	if err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		infh, err := f.Open(path)
		if err != nil {
			return errors.Wrapf(err, "open input file %v", path)
		}
		info, err := fs.Stat(f, path)
		if err != nil {
			return errors.Wrapf(err, "stat %v", path)
		}
		if err := tw.WriteHeader(&tar.Header{
			Name:    path,
			Size:    info.Size(),
			Mode:    int64(info.Mode()),
			ModTime: info.ModTime(),
		}); err != nil {
			return errors.Wrapf(err, "write tar header for %v", path)
		}
		if _, err := io.Copy(tw, infh); err != nil {
			return errors.Wrapf(err, "copy input file %v to archive", path)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk fs")
	}

	outputOK = true
	return &Layer{
		Underlying: outfh.Name(),
		Descriptor: v1.Descriptor{
			MediaType: LayerMediaType,
			Size:      int64(cw.n),
		},
	}, nil
}

func (l *Layer) Close() error {
	if l.Underlying != "" {
		if err := os.Remove(l.Underlying); err != nil {
			return errors.Wrap(err, "remove underlying file")
		}
		l.Underlying = ""
	}
	return nil
}
