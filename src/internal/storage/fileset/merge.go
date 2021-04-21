package fileset

import (
	"context"
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// MergeReader is an abstraction for reading merged filesets.
// A file's content is ordered based on the lexicographical order of
// the tagged content, so the output file content is produced by
// performing a merge of the tagged content.
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
func (mr *MergeReader) Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error {
	if len(deletive) > 0 && deletive[0] {
		return mr.iterateDeletive(ctx, cb)
	}
	return mr.iterate(ctx, cb)
}

func (mr *MergeReader) iterate(ctx context.Context, cb func(File) error) error {
	var ss []stream.Stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs, true),
			priority: len(ss),
			deletive: true,
		})
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs),
			priority: len(ss),
		})
	}
	pq := stream.NewPriorityQueue(ss)
	return pq.Iterate(func(ss []stream.Stream, _ ...string) error {
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		if len(fss) == 1 {
			if fss[0].deletive {
				return nil
			}
			return cb(newFileReader(ctx, mr.chunks, fss[0].file.Index()))
		}
		idxs := mergeFile(fss)
		for _, idx := range sortIndexes(idxs) {
			if err := cb(newMergeFileReader(ctx, mr.chunks, idx)); err != nil {
				return err
			}
		}
		return nil
	})
}

func mergeFile(fss []*fileStream) map[string]*index.Index {
	idxs := make(map[string]*index.Index)
	// TODO: Might as well just swap the priority in the priority queue from lowest to highest during set up.
	// That lines up more with how we lay out the filesets.
	for i := len(fss) - 1; i >= 0; i-- {
		idx := fss[i].file.Index()
		if fss[i].deletive {
			idxs = make(map[string]*index.Index)
			continue
		}
		if _, ok := idxs[idx.File.Tag]; !ok {
			idxs[idx.File.Tag] = &index.Index{
				Path: idx.Path,
				File: &index.File{
					Tag: idx.File.Tag,
				},
			}
		}
		mergeIdx := idxs[idx.File.Tag]
		mergeIdx.File.DataRefs = append(mergeIdx.File.DataRefs, idx.File.DataRefs...)
	}
	return idxs
}

func sortIndexes(idxs map[string]*index.Index) []*index.Index {
	var result []*index.Index
	for _, idx := range idxs {
		result = append(result, idx)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].File.Tag < result[j].File.Tag
	})
	return result
}

func (mr *MergeReader) iterateDeletive(ctx context.Context, cb func(File) error) error {
	var ss []stream.Stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs, true),
			priority: len(ss),
		})
	}
	pq := stream.NewPriorityQueue(ss)
	return pq.Iterate(func(ss []stream.Stream, _ ...string) error {
		idx := ss[0].(*fileStream).file.Index()
		return cb(newFileReader(ctx, mr.chunks, &index.Index{
			Path: idx.Path,
		}))
	})
}

// MergeFileReader is an abstraction for reading a merged file.
type MergeFileReader struct {
	ctx    context.Context
	chunks *chunk.Storage
	idx    *index.Index
}

func newMergeFileReader(ctx context.Context, chunks *chunk.Storage, idx *index.Index) *MergeFileReader {
	return &MergeFileReader{
		ctx:    ctx,
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
func (mfr *MergeFileReader) Content(w io.Writer) error {
	r := mfr.chunks.NewReader(mfr.ctx, mfr.idx.File.DataRefs)
	return r.Get(w)
}

type fileStream struct {
	iterator *Iterator
	file     File
	priority int
	deletive bool
}

func (fs *fileStream) Next() error {
	var err error
	fs.file, err = fs.iterator.Next()
	return err
}

func (fs *fileStream) Key() string {
	return fs.file.Index().Path
}

func (fs *fileStream) Priority() int {
	return fs.priority
}
