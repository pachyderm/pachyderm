package kv

import (
	"context"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"go.uber.org/zap"
)

var _ Store = &FSStore{}

type FSStore struct {
	dir      string
	initOnce sync.Once
}

func NewFSStore(dir string) *FSStore {
	return &FSStore{
		dir: dir,
	}
}

func (s *FSStore) Put(ctx context.Context, key, value []byte) (retErr error) {
	log.Debug(ctx, "put", zap.ByteString("key", key), zap.Int("value_len", len(value)))
	if err := s.ensureInit(ctx); err != nil {
		return err
	}
	staging := s.stagingPathFor(key)
	final := s.finalPathFor(key)
	defer s.cleanupFile(ctx, &retErr, staging)
	if err := os.WriteFile(staging, value, 0o644); err != nil {
		return s.transformError(err, key)
	}
	return os.Rename(staging, final)
}

func (s *FSStore) Get(ctx context.Context, key, buf []byte) (_ int, retErr error) {
	f, err := os.Open(s.finalPathFor(key))
	if err != nil {
		return 0, s.transformError(err, key)
	}
	defer s.closeFile(ctx, &retErr, f)
	var n int
	for {
		n2, err := f.Read(buf[n:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				n += n2
				break
			}
			return 0, s.transformError(err, key)
		}
		n += n2
	}
	return n, nil
}

func (s *FSStore) Exists(ctx context.Context, key []byte) (bool, error) {
	_, err := os.Stat(s.finalPathFor(key))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, s.transformError(err, key)
	}
	return true, nil
}

func (s *FSStore) Delete(ctx context.Context, key []byte) error {
	if err := s.ensureInit(ctx); err != nil {
		return err
	}
	err := os.Remove(s.finalPathFor(key))
	if os.IsNotExist(err) {
		err = nil
	}
	return err
}

func (c *FSStore) stagingPathFor(k []byte) string {
	return filepath.Join(c.dir, "staging", uuid.NewWithoutDashes())
}

func (c *FSStore) finalPathFor(k []byte) string {
	return filepath.Join(c.dir, "objects", hex.EncodeToString(k))
}

func (s *FSStore) NewKeyIterator(span Span) stream.Iterator[[]byte] {
	return &fsIterator{s: s, span: span}
}

type fsIterator struct {
	s    *FSStore
	span Span
	keys [][]byte
	pos  int
}

func (it *fsIterator) Next(ctx context.Context, dst *[]byte) error {
	if it.keys == nil {
		dirEnts, err := os.ReadDir(filepath.Join(it.s.dir, "objects"))
		if err != nil {
			return err
		}
		var keys [][]byte
		for i := range dirEnts {
			key, err := hex.DecodeString(dirEnts[i].Name())
			if err != nil {
				return err
			}
			keys = append(keys, key)
		}
		it.keys = keys
	}
	for it.pos < len(it.keys) {
		k := it.keys[it.pos]
		if it.span.Contains(k) {
			*dst = append((*dst)[:0], k...)
			it.pos++
			return nil
		} else {
			it.pos++
		}
	}
	return stream.EOS
}

func (s *FSStore) ensureInit(ctx context.Context) (err error) {
	s.initOnce.Do(func() {
		err = s.init(ctx)
	})
	return err
}

func (s *FSStore) init(ctx context.Context) error {
	if err := os.RemoveAll(filepath.Join(s.dir, "staging")); err != nil {
		return errors.EnsureStack(err)
	}
	if err := os.MkdirAll(filepath.Join(s.dir, "staging"), 0755); err != nil {
		return errors.EnsureStack(err)
	}
	if err := os.MkdirAll(filepath.Join(s.dir, "objects"), 0755); err != nil {
		return errors.EnsureStack(err)
	}
	log.Info(pctx.TODO(), "successfully initialized fs-backed object store", zap.String("root", s.dir))
	return nil
}

func (s *FSStore) transformError(err error, key []byte) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) || strings.HasSuffix(err.Error(), ": no such file or directory") {
		return pacherr.NewNotExist(s.dir, string(key))
	}
	return err
}

func (c *FSStore) closeFile(ctx context.Context, retErr *error, f *os.File) {
	err := f.Close()
	if err != nil && !strings.Contains(err.Error(), "file already closed") {
		if retErr == nil {
			*retErr = err
		} else {
			log.Error(ctx, "error closing file", zap.Error(err))
		}
	}
}

// cleanupFile is called to cleanup files from the staging area
func (c *FSStore) cleanupFile(ctx context.Context, retErr *error, p string) {
	err := os.Remove(p)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		if retErr == nil {
			*retErr = err
		} else {
			log.Error(ctx, "error deleting file", zap.Error(err))
		}
	}
}
