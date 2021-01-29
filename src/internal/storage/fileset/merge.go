package fileset

import (
	"context"
	"io"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
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
	var skipPrefix string
	var skipPriority int
	return pq.iterate(func(ss []stream, _ ...string) error {
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		if skipPrefix != "" && strings.HasPrefix(fss[0].key(), skipPrefix) {
			for i, fs := range fss {
				if fs.streamPriority() < skipPriority {
					fss = fss[:i]
					break
				}
			}
		}
		if len(fss) == 0 {
			return nil
		}
		if IsDir(fss[0].key()) {
			skipPrefix = fss[0].key()
			skipPriority = fss[0].streamPriority()
			return nil
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
		Path: fss[0].file.Index().Path,
		File: &index.File{},
	}
	var ps []*partStream
	for _, fs := range fss {
		idx := fs.file.Index()
		if fs.deletive && idx.File.Parts == nil {
			break
		}
		ps = append(ps, &partStream{
			parts:    idx.File.Parts,
			deletive: fs.deletive,
		})
	}
	// Merge the parts based on the lexicograhical ordering of the tags.
	mergeIdx.File.Parts = mergeParts(ps)
	return mergeIdx
}

func mergeParts(pss []*partStream) []*index.Part {
	if len(pss) == 0 {
		return nil
	}
	var ss []stream
	for i := len(pss) - 1; i >= 0; i-- {
		pss[i].priority = len(ss)
		ss = append(ss, pss[i])
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
		lastPart.SizeBytes += part.SizeBytes
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
		Path: idxs[0].Path,
		File: &index.File{},
	}
	var ps []*partStream
	for _, idx := range idxs {
		// Handle full delete.
		if idx.File.Parts == nil {
			return mergeIdx
		}
		ps = append(ps, &partStream{
			parts: idx.File.Parts,
		})
	}
	// Merge the parts based on the lexicograhical ordering of the tags.
	mergeIdx.File.Parts = mergeParts(ps)
	return mergeIdx
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
func (mfr *MergeFileReader) Index() *index.Index {
	return proto.Clone(mfr.idx).(*index.Index)
}

// Content returns the content of the merged file.
func (mfr *MergeFileReader) Content(w io.Writer) error {
	dataRefs := getDataRefs(mfr.idx.File.Parts)
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
