package fileset

import (
	"bytes"
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
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
	return mr.iterateAdditive(ctx, cb)
}

func (mr *MergeReader) iterateAdditive(ctx context.Context, cb func(File) error) error {
	var ss []stream
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
	pq := newPriorityQueue(ss)
	return pq.iterate(func(ss []stream, _ ...string) error {
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
		idx := mergeFile(fss)
		// Handle a full delete.
		if len(idx.File.Parts) == 0 {
			return nil
		}
		return cb(newMergeFileReader(ctx, mr.chunks, idx))
	})
}

func mergeFile(fss []*fileStream) *index.Index {
	mergeIdx := &index.Index{
		File: &index.File{},
	}
	mergeIdx.Path = fss[0].file.Index().Path
	var headerPart *index.Part
	var ps []*partStream
	for _, fs := range fss {
		idx := fs.file.Index()
		if fs.deletive && idx.File.Parts == nil {
			break
		}
		// Collect the header part from the lowest priority index.
		if !fs.deletive {
			headerPart = getHeaderPart(idx.File.Parts)
		}
		ps = append(ps, &partStream{
			parts:    getContentParts(idx.File.Parts),
			deletive: fs.deletive,
		})
		mergeIdx.SizeBytes += idx.SizeBytes
	}
	// Merge the parts based on the lexicograhical ordering of the tags.
	contentParts := mergeParts(ps)
	if len(contentParts) == 0 {
		return mergeIdx
	}
	mergeIdx.File.Parts = append([]*index.Part{headerPart}, contentParts...)
	return mergeIdx
}

func mergeParts(ps []*partStream) []*index.Part {
	if len(ps) == 0 {
		return nil
	}
	var ss []stream
	for i := len(ps) - 1; i >= 0; i-- {
		ps[i].priority = len(ss)
		ss = append(ss, ps[i])
	}
	pq := newPriorityQueue(ss)
	var mergedParts []*index.Part
	pq.iterate(func(ss []stream, _ ...string) error {
		for i := 0; i < len(ss); i++ {
			ps := ss[i].(*partStream)
			if ps.deletive {
				return nil
			}
			mergedParts = mergePart(mergedParts, ps.part)
		}
		return nil
	})
	return mergedParts
}

func mergePart(parts []*index.Part, part *index.Part) []*index.Part {
	if len(parts) == 0 {
		return []*index.Part{part}
	}
	lastPart := parts[len(parts)-1]
	if lastPart.Tag == part.Tag {
		if part.DataRefs != nil {
			lastPart.DataRefs = append(lastPart.DataRefs, part.DataRefs...)
		}
		return parts
	}
	return append(parts, part)
}

func (mr *MergeReader) iterateDeletive(ctx context.Context, cb func(File) error) error {
	var ss []stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs, true),
			priority: len(ss),
		})
	}
	pq := newPriorityQueue(ss)
	return pq.iterate(func(ss []stream, _ ...string) error {
		var idxs []*index.Index
		for _, s := range ss {
			idxs = append(idxs, s.(*fileStream).file.Index())
		}
		idx := mergeDeletes(idxs)
		return cb(newFileReader(ctx, mr.chunks, idx))
	})
}

func mergeDeletes(idxs []*index.Index) *index.Index {
	mergeIdx := &index.Index{
		File: &index.File{},
	}
	mergeIdx.Path = idxs[0].Path
	var ps []*partStream
	for _, idx := range idxs {
		if idx.File.Parts == nil {
			break
		}
		ps = append(ps, &partStream{
			parts: getContentParts(idx.File.Parts),
		})
	}
	// Merge the parts based on the lexicograhical ordering of the tags.
	contentParts := mergeParts(ps)
	if len(contentParts) == 0 {
		return mergeIdx
	}
	mergeIdx.File.Parts = append(mergeIdx.File.Parts, contentParts...)
	return mergeIdx
}

// MergeFileReader is an abstraction for reading a merged file.
type MergeFileReader struct {
	ctx    context.Context
	chunks *chunk.Storage
	idx    *index.Index
	hdr    *tar.Header
}

func newMergeFileReader(ctx context.Context, chunks *chunk.Storage, idx *index.Index) *MergeFileReader {
	return &MergeFileReader{
		ctx:    ctx,
		chunks: chunks,
		idx:    idx,
	}
}

// Index returns the index for the merged file.
func (mfr *MergeFileReader) Index() *index.Index {
	return proto.Clone(mfr.idx).(*index.Index)
}

// Header returns the tar header for the merged file.
func (mfr *MergeFileReader) Header() (*tar.Header, error) {
	if mfr.hdr == nil {
		buf := &bytes.Buffer{}
		r := mfr.chunks.NewReader(mfr.ctx, getHeaderPart(mfr.idx.File.Parts).DataRefs)
		if err := r.Get(buf); err != nil {
			return nil, err
		}
		tr := tar.NewReader(buf)
		var err error
		mfr.hdr, err = tr.Next()
		if err != nil {
			return nil, err
		}
		mfr.hdr.Size = mfr.idx.SizeBytes
	}
	return mfr.hdr, nil
}

// Content returns the content of the merged file.
func (mfr *MergeFileReader) Content(w io.Writer) error {
	dataRefs := getDataRefs(getContentParts(mfr.idx.File.Parts))
	r := mfr.chunks.NewReader(mfr.ctx, dataRefs)
	return r.Get(w)
}

type fileStream struct {
	iterator *Iterator
	file     File
	priority int
	deletive bool
}

func (fs *fileStream) next() error {
	var err error
	fs.file, err = fs.iterator.Next()
	return err
}

func (fs *fileStream) key() string {
	return fs.file.Index().Path
}

func (fs *fileStream) streamPriority() int {
	return fs.priority
}

type partStream struct {
	parts    []*index.Part
	part     *index.Part
	priority int
	deletive bool
}

func (ps *partStream) next() error {
	if len(ps.parts) == 0 {
		return io.EOF
	}
	ps.part = ps.parts[0]
	ps.parts = ps.parts[1:]
	return nil
}

func (ps *partStream) key() string {
	return ps.part.Tag
}

func (ps *partStream) streamPriority() int {
	return ps.priority
}
