package fileset

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
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
	var ss []stream.Stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs.IterateDeletes, opts...),
			deletive: true,
		})
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs.Iterate, opts...),
		})
	}
	pq := stream.NewPriorityQueue(ss, compare)
	return pq.Iterate(func(ss []stream.Stream) error {
		var fss []*fileStream
		for _, s := range ss {
			fs := s.(*fileStream)
			if fs.deletive {
				fss = nil
				continue
			}
			fss = append(fss, fs)
		}
		if len(fss) == 0 {
			return nil
		}
		if len(fss) == 1 {
			return cb(newFileReader(mr.chunks, fss[0].file.Index()))
		}
		var dataRefs []*chunk.DataRef
		for _, fs := range fss {
			idx := fs.file.Index()
			dataRefs = append(dataRefs, idx.File.DataRefs...)
		}
		mergeIdx := fss[0].file.Index()
		mergeIdx.File.DataRefs = dataRefs
		return cb(newMergeFileReader(mr.chunks, mergeIdx))

	})
}

func (mr *MergeReader) IterateDeletes(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	var ss []stream.Stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs.IterateDeletes, opts...),
		})
	}
	pq := stream.NewPriorityQueue(ss, compare)
	return pq.Iterate(func(ss []stream.Stream) error {
		fs := ss[0].(*fileStream)
		return cb(newFileReader(mr.chunks, fs.file.Index()))
	})
}

// TODO: Look at the sizes?
// TODO: Come up with better heuristics for sharding.
func (mr *MergeReader) Shards(ctx context.Context) ([]*index.PathRange, error) {
	shards, err := mr.fileSets[0].Shards(ctx)
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

type fileStream struct {
	iterator *Iterator
	file     File
	deletive bool
}

func (fs *fileStream) Next() error {
	var err error
	fs.file, err = fs.iterator.Next()
	return err
}

func compare(s1, s2 stream.Stream) int {
	idx1 := s1.(*fileStream).file.Index()
	idx2 := s2.(*fileStream).file.Index()
	if idx1.Path == idx2.Path {
		return strings.Compare(idx1.File.Datum, idx2.File.Datum)
	}
	return strings.Compare(idx1.Path, idx2.Path)
}
