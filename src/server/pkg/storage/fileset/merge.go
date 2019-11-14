package fileset

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// MergeReader merges a file's content that shows up across
// multiple file set streams.
// A file's content is ordered based on the lexicographical order of
// the tagged content, so the output file content is produced by
// performing a merge of the tagged content.
type MergeReader struct {
	pq *priorityQueue
}

func newMergeReader(rs []*Reader) *MergeReader {
	var fileStreams []stream
	for _, r := range rs {
		fileStreams = append(fileStreams, &fileStream{r: r})
	}
	return &MergeReader{pq: newPriorityQueue(fileStreams)}
}

func (mr *MergeReader) Get(w io.Writer) error {
	// (bryce) need tar EOF blocks.
	return mr.Iterate(func(fmr *FileMergeReader) error {
		return fmr.Get(w)
	})
}

func (mr *MergeReader) Iterate(f func(*FileMergeReader) error) error {
	return mr.pq.iterate(func(ss []stream, next string) error {
		// Convert generic streams to file streams.
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		var frs []*FileReader
		for _, fs := range fss {
			fr, err := fs.r.Next()
			if err != nil {
				return err
			}
			frs = append(frs, fr)
		}
		return f(newFileMergeReader(frs))
	})
}

func (mr *MergeReader) WriteTo(w *Writer) error {
	return mr.pq.iterate(func(ss []stream, next string) error {
		// Convert generic streams to file streams.
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		// Cheap copy.
		if len(fss) == 1 {
			return fss[0].r.Iterate(func(fr *FileReader) error {
				return w.CopyFile(fr)
			}, next)
		}
		// Regular copy.
		var frs []*FileReader
		for _, fs := range fss {
			fr, err := fs.r.Next()
			if err != nil {
				return err
			}
			frs = append(frs, fr)
		}
		return fmr.WriteTo(w)
	})
}

type fileStream struct {
	r   *Reader
	idx *index.Index
}

func (fs *fileStream) next() error {
	var err error
	fs.idx, err = fs.r.Peek()
	return err
}

func (fs *fileStream) key() string {
	return fs.idx.Path
}

type FileMergeReader struct {
	tmr *TagMergeReader
}

func newFileMergeReader(frs []*FileReader) *FileMergeReader {
	return &FileMergeReader{tmr: newTagMergeReader(frs)}
}

func (fmr *FileMergeReader) Get(w io.Writer) error {
	tw := tar.NewWriter(w)
	// Write merged header.
	hdr, err := fmr.tmr.Header()
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	// Write merged content.
	if err := fmr.tmr.Get(tw); err != nil {
		return err
	}
	return tw.Flush()
}

func (fmr *FileMergeReader) TagMergeReader() *TagMergeReader {
	return fmr.tmr
}

func (fmr *FileMergeReader) WriteTo(w *Writer) error {
	tw := tar.NewWriter(w)
	// Write merged header.
	hdr, err := fmr.tmr.Header()
	if err != nil {
		return err
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if err := fmr.tmr.WriteTo(w); err != nil {
		return err
	}
	return tw.Flush()
}

// TagMergeReader merges the tagged content in a file.
// Tags in a file should be unique across file sets being merged,
// so the operation is a simple copy.
type TagMergeReader struct {
	pq *priorityQueue
}

func newTagMergeReader(frs []*FileReader) *TagMergeReader {
	// (bryce) need to handle delete operations.
	var tagStreams []stream
	for _, fr := range frs {
		tagStreams = append(tagStreams, &tagStream{cr: fr.cr})
	}
	return &TagMergeReader{pq: newPriorityQueue(tagStreams)}
}

func (tmr *TagMergeReader) Header() (*tar.Header, error) {
	var size int64
	for _, tr := range tmr.trs {
		hdr, err := tr.Header()
		if err != nil {
			return nil, err
		}
		size += hdr.Size
	}
	hdr, err := tmr.trs[0].Header()
	if err != nil {
		return nil, err
	}
	hdr.Size = size
	return hdr, nil
}

func (tmr *TagMergeReader) Get(w io.Writer) error {
	return tmr.Iterate(func(_ *index.Tag, r io.Reader) error {
		_, err := io.Copy(w, r)
		return err
	})
}

func (tmr *TagMergeReader) WriteTo(w *Writer) error {
	return tmr.pq.iterate(func(ss []stream, next string) error {
		// (bryce) this can actually happen when serializing parts of
		// in memory file set, need to change this.
		if len(ss) > 1 {
			return fmt.Errorf("tags should be distinct within a file")
		}
		// Convert generic stream to tag stream.
		ts := ss[0].(*tagStream)
		// Cheap copy.
		return w.CopyTags(ts.dr, next)
	})
}

func (tmr *TagMergeReader) Iterate(f func(*chunk.Tag, io.Reader) error) error {
	return tmr.pq.iterate(func(ss []stream, next string) error {
		// (bryce) this can actually happen when serializing parts of
		// in memory file set, need to change this.
		if len(ss) > 1 {
			return fmt.Errorf("tags should be distinct within a file")
		}
		// Convert generic stream to tag stream.
		ts := ss[0].(*tagStream)
		return ts.dr.Iterate(f, tmr.pq.peek())
	})
}

type tagStream struct {
	cr  *chunk.Reader
	dr  *chunk.DataReader
	tag *chunk.Tag
}

func (ts *tagStream) next() error {
	if ts.dr == nil || ts.dr.Done() {
		ts.dr, err = cr.Peek()
		if err != nil {
			return err
		}
	}
	var err error
	ts.tag, err = ts.dr.Peek()
	return err
}

func (ts *tagStream) key() string {
	return ts.tag.Id
}

//// shard creates shards (path ranges) from the file set streams being merged.
//// A shard is created when the size of the content for a path range is greater than
//// the passed in shard threshold.
//// For each shard, the callback is called with the path range for the shard.
//func shard(ss []stream, shardThreshold int64, f func(r *index.PathRange) error) {
//	pq := newPriorityQueue(ss)
//	var size int64
//	pathRange := &index.PathRange{}
//	for {
//		ss, err := pq.Next()
//		if err != nil {
//			if err == io.EOF {
//				return nil
//			}
//			return nil, err
//		}
//		// Convert generic streams to file streams.
//		var fileStreams []*fileStream
//		for _, s := range ss {
//			fileStreams = append(fileStreams, s.(*fileStream))
//		}
//		var hdr *index.Header
//		var err error
//		last := true
//		for _, fs := range fileStreams {
//			hdr, err = fs.r.Next()
//			if err != nil {
//				return err
//			}
//			size += hdr.Idx.SizeBytes
//			if _, err = fs.r.Peek(); err != io.EOF {
//				last = false
//			}
//		}
//		if pathRange.Lower == "" {
//			pathRange.Lower = hdr.Hdr.Name
//		}
//		// A shard is created when we have encountered more than shardThreshold content bytes.
//		// The last shard is created when there are no more streams in the queue and the
//		// streams being handled in this call are all io.EOF.
//		if size > shardThreshold || len(next) == 0 && last {
//			pathRange.Upper = hdr.Hdr.Name
//			if err := f(pathRange); err != nil {
//				return err
//			}
//			size = 0
//			pathRange = &index.PathRange{}
//			return nil
//		}
//		return nil
//	}
//}
