package jobs

import (
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/blake3"
)

// Cache is an Artifact cache.
type Cache struct {
	Path string
}

type cacheableFile struct {
	io.Writer
	f    *os.File
	h    hash.Hash
	c    *Cache
	name string
}

type CacheFile interface {
	io.WriteCloser
	Sum(b []byte) []byte
	Path() string
}

func (c *Cache) NewCacheableFile(name string) (CacheFile, error) {
	fh, err := os.CreateTemp("", "cache-*")
	if err != nil {
		return nil, errors.Wrap(err, "create temporary file")
	}
	blake := blake3.New()
	tee := io.MultiWriter(fh, blake)
	return &cacheableFile{Writer: tee, f: fh, h: blake, c: c, name: name}, nil
}

func (f *cacheableFile) Sum(b []byte) []byte {
	return f.h.Sum(b)
}

func (f *cacheableFile) Path() string {
	hash := hex.EncodeToString(f.Sum(nil))
	return filepath.Join(f.c.Path, hash, f.name)
}

func (f *cacheableFile) Close() error {
	if err := f.f.Close(); err != nil {
		return errors.Wrapf(err, "close underlying os file %v for %v", f.f.Name(), f.name)
	}
	hash := hex.EncodeToString(f.Sum(nil))
	if err := os.MkdirAll(filepath.Join(f.c.Path, hash), 0o755); err != nil {
		return errors.Wrapf(err, "make cache directory for %v (id %v)", f.name, hash)
	}
	if err := os.Rename(f.f.Name(), f.Path()); err != nil {
		return errors.Wrap(err, "move file to final destination")
	}
	return nil
}
