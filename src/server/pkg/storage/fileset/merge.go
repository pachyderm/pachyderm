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
func (mr *MergeReader) Iterate(ctx context.Context, cb func(File) error) error {
	var ss []stream
	for _, fs := range mr.fileSets {
		ss = append(ss, &fileStream{
			iterator: NewIterator(ctx, fs),
			priority: len(ss),
		})
	}
	pq := newPriorityQueue(ss)
	return pq.iterate(func(ss []stream, _ ...string) error {
		// Collect files from file streams.
		var files []File
		for _, s := range ss {
			fs := s.(*fileStream)
			files = append(files, fs.file)
		}
		if len(files) == 1 {
			return cb(files[0])
		}
		idx := mergeFiles(files)
		return cb(newMergeFileReader(ctx, mr.chunks, idx))
	})
}

type fileStream struct {
	iterator *Iterator
	file     File
	priority int
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

func mergeFiles(fs []File) *index.Index {
	mergeIdx := &index.Index{
		FileOp: &index.FileOp{},
	}
	mergeIdx.Path = fs[0].Index().Path
	// Handle a full delete.
	if fs[0].Index().FileOp.Op == index.Op_DELETE {
		mergeIdx.FileOp.Op = index.Op_DELETE
		return mergeIdx
	}
	// Collect the indexes for the files being merged.
	var idxs []*index.Index
	for _, f := range fs {
		idx := f.Index()
		if idx.FileOp.Op == index.Op_DELETE {
			mergeIdx.FileOp.Op = index.Op_OVERWRITE
			break
		}
		idxs = append(idxs, idx)
		if idx.FileOp.Op == index.Op_OVERWRITE {
			mergeIdx.FileOp.Op = index.Op_OVERWRITE
			break
		}
	}
	// Collect the header data op from the highest priority file.
	hdo := getHeaderDataOp(idxs[0].FileOp.DataOps)
	// Collect the content data ops, then merge based on the lexicograhical ordering of the tags.
	var dos [][]*index.DataOp
	for _, idx := range idxs {
		dos = append(dos, getContentDataOps(idx.FileOp.DataOps))
		mergeIdx.SizeBytes += idx.SizeBytes
	}
	mergeIdx.FileOp.DataOps = append([]*index.DataOp{hdo}, mergeDataOps(dos)...)
	return mergeIdx
}

func mergeDataOps(dataOps [][]*index.DataOp) []*index.DataOp {
	if len(dataOps) == 0 {
		return nil
	}
	var ss []stream
	for i := len(dataOps) - 1; i >= 0; i-- {
		ss = append(ss, &dataOpStream{
			dataOps:  dataOps[i],
			priority: len(ss),
		})
	}
	pq := newPriorityQueue(ss)
	var mergedDataOps []*index.DataOp
	pq.iterate(func(ss []stream, _ ...string) error {
		for i := 0; i < len(ss); i++ {
			dataOp := ss[i].(*dataOpStream).dataOp
			mergedDataOps = mergeDataOp(mergedDataOps, dataOp)
		}
		return nil
	})
	return mergedDataOps
}

type dataOpStream struct {
	dataOp   *index.DataOp
	dataOps  []*index.DataOp
	priority int
}

func (dos *dataOpStream) next() error {
	if len(dos.dataOps) == 0 {
		return io.EOF
	}
	dos.dataOp = dos.dataOps[0]
	dos.dataOps = dos.dataOps[1:]
	return nil
}

func (dos *dataOpStream) key() string {
	return dos.dataOp.Tag
}

func (dos *dataOpStream) streamPriority() int {
	return dos.priority
}

func mergeDataOp(dataOps []*index.DataOp, dataOp *index.DataOp) []*index.DataOp {
	if len(dataOps) == 0 {
		return []*index.DataOp{dataOp}
	}
	lastDataOp := dataOps[len(dataOps)-1]
	if lastDataOp.Tag == dataOp.Tag {
		if dataOp.Op == index.Op_DELETE || dataOp.Op == index.Op_OVERWRITE {
			lastDataOp.DataRefs = nil
		}
		lastDataOp.DataRefs = append(lastDataOp.DataRefs, dataOp.DataRefs...)
		return dataOps
	}
	return append(dataOps, dataOp)
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
		r := mfr.chunks.NewReader(mfr.ctx, getHeaderDataOp(mfr.idx.FileOp.DataOps).DataRefs)
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
	dataRefs := getDataRefs(getContentDataOps(mfr.idx.FileOp.DataOps))
	r := mfr.chunks.NewReader(mfr.ctx, dataRefs)
	return r.Get(w)
}
