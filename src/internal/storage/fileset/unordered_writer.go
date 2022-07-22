package fileset

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// UnorderedWriter allows writing Files, unordered by path, into multiple ordered filesets.
// This may be a full filesystem or a subfilesystem (e.g. datum / datum set / shard).
type UnorderedWriter struct {
	ctx                        context.Context
	storage                    *Storage
	memAvailable, memThreshold int64
	fileThreshold              int64
	buffer                     *Buffer
	subFileSet                 int64
	ttl                        time.Duration
	renewer                    *Renewer
	ids                        []ID
	getParentID                func() (*ID, error)
	validator                  func(string) error
	maxFanIn                   int
}

func newUnorderedWriter(ctx context.Context, storage *Storage, memThreshold, fileThreshold int64, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	if err := storage.filesetSem.Acquire(ctx, 1); err != nil {
		return nil, errors.EnsureStack(err)
	}
	// Half of the memory will be for buffering in the unordered writer.
	// The other half will be for buffering in the chunk writer.
	memThreshold /= 2
	uw := &UnorderedWriter{
		ctx:           ctx,
		storage:       storage,
		fileThreshold: fileThreshold,
		memAvailable:  memThreshold,
		memThreshold:  memThreshold,
		buffer:        NewBuffer(),
		maxFanIn:      math.MaxInt32,
	}
	for _, opt := range opts {
		opt(uw)
	}
	return uw, nil
}

func (uw *UnorderedWriter) Put(ctx context.Context, p, datum string, appendFile bool, r io.Reader) (retErr error) {
	if err := uw.validate(p); err != nil {
		return err
	}
	if datum == "" {
		datum = DefaultFileDatum
	}
	if !appendFile {
		uw.buffer.Delete(p, datum)
	}
	w := uw.buffer.Add(p, datum)
	for {
		n, err := io.CopyN(w, r, uw.memAvailable)
		uw.memAvailable -= n
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return errors.EnsureStack(err)
		}
		if uw.memAvailable == 0 {
			if err := uw.serialize(ctx); err != nil {
				return err
			}
			w = uw.buffer.Add(p, datum)
		}
	}
	if int64(uw.buffer.Count()) >= uw.fileThreshold {
		return uw.serialize(ctx)
	}
	return nil
}

func (uw *UnorderedWriter) validate(p string) error {
	if uw.validator != nil {
		return uw.validator(p)
	}
	return nil
}

// serialize will be called whenever the in-memory file set is past the memory threshold.
// A new in-memory file set will be created for the following operations.
func (uw *UnorderedWriter) serialize(ctx context.Context) error {
	if uw.buffer.Empty() {
		return nil
	}
	return miscutil.LogStep(ctx, "UnorderedWriter.serialize", func() error {
		return uw.withWriter(func(w *Writer) error {
			if err := uw.buffer.WalkAdditive(func(path, datum string, r io.Reader) error {
				return w.Add(path, datum, r)
			}, func(f File, datum string) error {
				return w.Copy(f, datum)
			}); err != nil {
				return err
			}
			return uw.buffer.WalkDeletive(func(path, datum string) error {
				return w.Delete(path, datum)
			})
		})
	})
}

func (uw *UnorderedWriter) withWriter(cb func(*Writer) error) error {
	// Serialize file set.
	var writerOpts []WriterOption
	if uw.ttl > 0 {
		writerOpts = append(writerOpts, WithTTL(uw.ttl))
	}
	w := uw.storage.newWriter(uw.ctx, writerOpts...)
	if err := cb(w); err != nil {
		return err
	}
	id, err := w.Close()
	if err != nil {
		return err
	}
	uw.ids = append(uw.ids, *id)
	if uw.renewer != nil {
		if err := uw.renewer.Add(uw.ctx, *id); err != nil {
			return err
		}
	}
	// Reset fileset buffer.
	uw.buffer = NewBuffer()
	uw.memAvailable = uw.memThreshold
	uw.subFileSet++
	return nil
}

// Delete deletes a file from the file set.
func (uw *UnorderedWriter) Delete(ctx context.Context, p, datum string) error {
	if err := uw.validate(p); err != nil {
		return err
	}
	if datum == "" {
		datum = DefaultFileDatum
	}
	p = Clean(p, IsDir(p))
	if IsDir(p) {
		uw.buffer.Delete(p, datum)
		var ids []ID
		if uw.getParentID != nil {
			parentID, err := uw.getParentID()
			if err != nil {
				return err
			}
			ids = []ID{*parentID}
		}
		fs, err := uw.storage.Open(uw.ctx, append(ids, uw.ids...), index.WithPrefix(p))
		if err != nil {
			return err
		}
		err = fs.Iterate(uw.ctx, func(f File) error {
			return uw.Delete(ctx, f.Index().Path, datum)
		})
		return errors.EnsureStack(err)
	}
	uw.buffer.Delete(p, datum)
	if int64(uw.buffer.Count()) >= uw.fileThreshold {
		return uw.serialize(ctx)
	}
	return nil
}

func (uw *UnorderedWriter) Copy(ctx context.Context, fs FileSet, datum string, appendFile bool) error {
	if datum == "" {
		datum = DefaultFileDatum
	}
	return errors.EnsureStack(fs.Iterate(ctx, func(f File) error {
		if !appendFile {
			uw.buffer.Delete(f.Index().Path, datum)
		}
		uw.buffer.Copy(f, datum)
		if int64(uw.buffer.Count()) >= uw.fileThreshold {
			return uw.serialize(ctx)
		}
		return nil
	}))
}

// Close closes the writer.
func (uw *UnorderedWriter) Close(ctx context.Context) (*ID, error) {
	defer uw.storage.filesetSem.Release(1)
	if err := uw.serialize(ctx); err != nil {
		return nil, err
	}
	if err := miscutil.LogStep(ctx, fmt.Sprintf("directly compacting %d file sets", len(uw.ids)), func() error {
		return uw.compact()
	}); err != nil {
		return nil, err
	}
	return uw.storage.newComposite(uw.ctx, &Composite{
		Layers: idsToHex(uw.ids),
	}, uw.ttl)
}

func (uw *UnorderedWriter) compact() error {
	for len(uw.ids) > uw.maxFanIn {
		var ids []ID
		for start := 0; start < len(uw.ids); start += int(uw.maxFanIn) {
			end := start + int(uw.maxFanIn)
			if end > len(uw.ids) {
				end = len(uw.ids)
			}
			id, err := uw.storage.Compact(uw.ctx, uw.ids[start:end], uw.ttl)
			if err != nil {
				return err
			}
			if uw.renewer != nil {
				if err := uw.renewer.Add(uw.ctx, *id); err != nil {
					return err
				}
			}
			ids = append(ids, *id)
		}
		uw.ids = ids
	}
	return nil
}
