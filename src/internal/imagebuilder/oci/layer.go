package oci

import (
	"archive/tar"
	"crypto/sha256"
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
	// Digest of the compressed archive.
	SHA256 [sha256.Size]byte
	// Digest of the uncompressed archive.
	DiffID [sha256.Size]byte
}

func (l *Layer) DiffIDAsDigest() digest.Digest {
	return digest.NewDigestFromBytes(digest.SHA256, l.DiffID[:])
}

type countWriter struct{ n int }

func (w *countWriter) Write(b []byte) (int, error) {
	w.n += len(b)
	return len(b), nil
}

func NewLayerFromFS(f fs.FS) (layer *Layer, retErr error) {
	outfh, err := os.CreateTemp("", "layer-*.tar.zst")
	if err != nil {
		return nil, errors.Wrap(err, "create underlying file")
	}
	compsha := sha256.New()
	uncompsha := sha256.New() // Nothing better than having to use the slowest hash TWICE.
	cw := new(countWriter)
	out := io.MultiWriter(outfh, compsha, cw)
	var outputOK bool
	defer func() {
		if layer != nil {
			d := compsha.Sum(nil)
			copy(layer.SHA256[:], d)
			copy(layer.DiffID[:], uncompsha.Sum(nil))
			layer.Descriptor.Size = int64(cw.n)
			layer.Descriptor.Digest = digest.NewDigestFromBytes(digest.SHA256, d)
		}
		if !outputOK {
			errors.JoinInto(&retErr, errors.Wrap(os.Remove(outfh.Name()), "remove partial output file"))
		}
	}()
	defer errors.Close(&retErr, outfh, "close underlying file")

	// TODO: let a slower option be passesd in for release builds
	zst, err := zstd.NewWriter(out, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		return nil, errors.Wrap(err, "create zstd writer")
	}
	defer errors.Close(&retErr, zst, "close zstd writer")

	comp := io.MultiWriter(zst, uncompsha)
	tw := tar.NewWriter(comp)
	defer errors.Close(&retErr, tw, "close tar writer")

	if err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) (walkErr error) {
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
		defer errors.Close(&walkErr, infh, "close input file %v", infh)

		info, err := infh.Stat()
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
		// SHA256 and DiffID added after closing output.
		SHA256:     [32]byte(compsha.Sum(nil)),
		DiffID:     [32]byte(uncompsha.Sum(nil)),
		Underlying: outfh.Name(),
		Descriptor: v1.Descriptor{
			MediaType: LayerMediaType,
			// Size and Digest added after closing output.
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
