package fileset

import (
	"bytes"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"golang.org/x/exp/slices"
)

// MergeReader is an abstraction for reading merged file sets.
type MergeReader struct {
	chunks   *chunk.Storage
	fileSets []FileSet
}

func newMergeReader(chunks *chunk.Storage, fileSets []FileSet) *MergeReader {
	return &MergeReader{
		chunks:   chunks,
		fileSets: fileSets,
	}
}

// Iterate iterates over the files in the merge reader.
func (mr *MergeReader) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	type fileEntry struct {
		File       File
		IsDeletive bool
	}
	copyFileEntry := func(dst, src *fileEntry) { *dst = *src }
	var ss []stream.Peekable[fileEntry]
	for _, fs := range mr.fileSets {
		// additive
		itAdd := stream.NewFromForEach(ctx, copyFileEntry, func(fn func(fileEntry) error) error {
			return fs.Iterate(ctx, func(f File) error {
				return fn(fileEntry{
					File:       f,
					IsDeletive: false,
				})
			}, opts...)
		})
		pkAdd := stream.NewPeekable(itAdd, copyFileEntry)
		ss = append(ss, pkAdd)
		// deletive
		itDel := stream.NewFromForEach(ctx, copyFileEntry, func(fn func(fileEntry) error) error {
			return fs.IterateDeletes(ctx, func(f File) error {
				return fn(fileEntry{
					File:       f,
					IsDeletive: true,
				})
			}, opts...)
		})
		pkDel := stream.NewPeekable(itDel, copyFileEntry)
		ss = append(ss, pkDel)
	}
	m := stream.NewMerger(ss, func(a, b fileEntry) bool {
		// path-wise merge, equal paths are equal.
		return a.File.Index().Path < b.File.Index().Path
	})
	return stream.ForEach[stream.Merged[fileEntry]](ctx, m, func(x stream.Merged[fileEntry]) error {
		var ents []fileEntry
		for _, fe := range x.Values {
			if fe.IsDeletive {
				continue
			}
			ents = append(ents, fe)
		}
		if len(ents) == 0 {
			return nil
		}
		if len(ents) == 1 {
			return cb(newFileReader(mr.chunks, ents[0].File.Index()))
		}
		slices.SortStableFunc(ents, func(a, b fileEntry) bool {
			// all of these will have the same path, so now we are only sorting by tag
			return a.File.Index().File.Datum < b.File.Index().File.Datum
		})
		var dataRefs []*chunk.DataRef
		for _, fe := range ents {
			idx := fe.File.Index()
			dataRefs = append(dataRefs, idx.File.DataRefs...)
		}
		mergeIdx := ents[0].File.Index()
		mergeIdx.File.DataRefs = dataRefs
		return cb(newMergeFileReader(mr.chunks, mergeIdx))
	})
}

func (mr *MergeReader) IterateDeletes(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	copyFile := func(dst, src *File) { *dst = *src }
	var ss []stream.Peekable[File]
	for _, fs := range mr.fileSets {
		it := stream.NewFromForEach(ctx, copyFile, func(fn func(File) error) error {
			return fs.IterateDeletes(ctx, cb, opts...)
		})
		pk := stream.NewPeekable(it, copyFile)
		ss = append(ss, pk)
	}
	m := stream.NewMerger(ss, fileLessThan)
	return stream.ForEach[stream.Merged[File]](ctx, m, func(x stream.Merged[File]) error {
		f, _ := x.First()
		return cb(newFileReader(mr.chunks, f.Index()))
	})
}

// TODO: Look at the sizes?
// TODO: Come up with better heuristics for sharding.
func (mr *MergeReader) Shards(ctx context.Context, opts ...index.Option) ([]*index.PathRange, error) {
	shards, err := mr.fileSets[0].Shards(ctx, opts...)
	return shards, errors.EnsureStack(err)
}

// MergeFileReader is an abstraction for reading a merged file.
type MergeFileReader struct {
	chunks *chunk.Storage
	idx    *index.Index
}

func newMergeFileReader(chunks *chunk.Storage, idx *index.Index) *MergeFileReader {
	return &MergeFileReader{
		chunks: chunks,
		idx:    idx,
	}
}

// Index returns the index for the merged file.
// TODO: Removed clone because it had a significant performance impact for small files.
// May want to revisit.
func (mfr *MergeFileReader) Index() *index.Index {
	return mfr.idx
}

// Content returns the content of the merged file.
func (mfr *MergeFileReader) Content(ctx context.Context, w io.Writer, opts ...chunk.ReaderOption) error {
	r := mfr.chunks.NewReader(ctx, mfr.idx.File.DataRefs, opts...)
	return r.Get(w)
}

// Hash returns the hash of the file.
// TODO: It would be good to remove this potential performance footgun, but it would require removing the append functionality.
func (mfr *MergeFileReader) Hash(ctx context.Context) ([]byte, error) {
	var hashes [][]byte
	size := index.SizeBytes(mfr.idx)
	if size >= DefaultBatchThreshold {
		// TODO: Optimize to handle large files that can mostly be copy by reference?
		if err := miscutil.WithPipe(func(w io.Writer) error {
			r := mfr.chunks.NewReader(ctx, mfr.idx.File.DataRefs)
			return r.Get(w)
		}, func(r io.Reader) error {
			uploader := mfr.chunks.NewUploader(ctx, "chunk-uploader-resolver", true, func(_ interface{}, dataRefs []*chunk.DataRef) error {
				for _, dataRef := range dataRefs {
					hashes = append(hashes, dataRef.Hash)
				}
				return nil
			})
			if err := uploader.Upload(nil, r); err != nil {
				return err
			}
			return uploader.Close()
		}); err != nil {
			return nil, err
		}
	} else {
		buf := &bytes.Buffer{}
		r := mfr.chunks.NewReader(ctx, mfr.idx.File.DataRefs)
		if err := r.Get(buf); err != nil {
			return nil, err
		}
		hashes = [][]byte{chunk.Hash(buf.Bytes())}
	}
	return computeFileHash(hashes)
}
