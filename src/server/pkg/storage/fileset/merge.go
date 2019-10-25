package fileset

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type fileStream struct {
	r   *Reader
	hdr *index.Header
}

func (fs *fileStream) next() error {
	var err error
	fs.hdr, err = fs.r.Peek()
	return err
}

func (fs *fileStream) key() string {
	return fs.hdr.Hdr.Name
}

type copyFunc func(r *Reader, bound string) error

// MergeReader merges a file's content that shows up across
// multiple file set streams.
// A file's content is ordered based on the lexicographical order of
// the tagged content, so the output file content is produced by
// performing a merge of the tagged content.
type MergeReader struct {
	pq    *priorityQueue
	tmr   *TagMergeReader
	copyF copyFunc
}

func NewMergeReader(rs []*Reader) *MergeReader {
	var fileStreams []stream
	for _, r := range rs {
		fileStreams = append(fileStreams, &fileStream{r: r})
	}
	return &MergeReader{pq: newPriorityQueue(fileStreams)}
}

func (mr *MergeReader) Next() (*index.Header, error) {
	ss, err := mr.pq.Next()
	if err != nil {
		return nil, err
	}
	// Fast path for copying files from one stream.
	for mr.copyF != nil && len(ss) == 1 {
		if err := mr.copyF(ss[0], mr.pq.peek()); err != nil {
			return nil, err
		}
		ss, err = m.Next()
		if err != nil {
			return nil, err
		}
	}
	// Convert generic streams to file streams.
	var fileStreams []*fileStream
	for _, s := range ss {
		fileStreams = append(fileStreams, s.(*fileStream))
	}
	// Setup tag streams for tag merge.
	var tagStreams []stream
	var size int64
	for _, fs := range fileStreams {
		hdr, err := fs.r.Next()
		if err != nil {
			return err
		}
		// (bryce) need to handle delete operations.
		tagStreams = append(tagStreams, &tagStream{r: fs.r})
		size += hdr.Idx.SizeBytes
	}
	mr.tmr = NewTagMergeReader(tagStreams)
	return &index.Header{
		Hdr: &tar.Header{
			Name: fileStreams[0].hdr.Hdr.Name,
			Size: size,
		},
	}, nil
}

func (mr *MergeReader) Read(data []byte) (int, error) {
	return mr.tmr.Read(data)
}

func (mr *MergeReader) TagMergeReader() (*TagMergeReader, error) {
	return mr.tmr
}

type tagStream struct {
	r   *Reader
	tag *index.Tag
}

func (ts *tagStream) next() error {
	var err error
	ts.tag, err = ts.r.PeekTag()
	return err
}

func (ts *tagStream) key() string {
	return ts.tag.Id
}

// TagMergeReader merges the tagged content in a file.
// Tags in a file should be unique across file sets being merged,
// so the operation is a simple copy.
type TagMergeReader struct {
	pq *priorityQueue
	tr TagReader
}

func NewTagMergeReader(ss []stream) *TagMergeReader {
	var tagStreams []stream
	for _, r := range rs {
		tagStreams = append(tagStreams, &tagStream{r: r})
	}
	return &TagMergeReader{pq: newPriorityQueue(tagStreams)}
}

func (tmr *TagMergeReader) Next() ([]*index.Tag, error) {
	if err := nextTagReader(); err != nil {
		return nil, err
	}
	return tmr.tr.Tags(), nil
}

func (tmr *TagMergeReader) Read(data []byte) (int, error) {
	var totalBytesRead int
	for len(data) > 0 {
		bytesRead, err := tmr.tr.Read(data)
		data = data[bytesRead:]
		totalBytesRead += bytesRead
		if err != nil {
			if err != io.EOF {
				return totalBytesRead, err
			}
			if err := r.nextTagReader(); err != nil {
				return totalBytesRead, err
			}
		}
	}
	return totalBytesRead, nil
}

func (tmr *TagMergeReader) nextTagReader() error {
	ss, err := tmr.pq.Next()
	if err != nil {
		return err
	}
	// (bryce) this should be an Internal error type.
	if len(ss) > 1 {
		return fmt.Errorf("tags should be distinct within a file")
	}
	// Convert generic stream to tag stream.
	tagStream := ss[0].(*tagStream)
	// Setup next tag reader and append tags.
	tmr.tr = tagStream.r.NewTagReader(tmr.pq.peek())
	return nil
}

func compact(w *Writer, rs []*Reader) error {
	mr := NewMergeReader(rs)
	for {
		hdr, err := mr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		tmr := mr.TagMergeReader()
		for {
			tr, err := tmr.Next()
			if err != nil {
				if err == io.EOF {
					continue
				}
				return err
			}
			if err := w.CopyTags(tmr); err != nil {
				return err
			}
		}
	}
}

// shard creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the content for a path range is greater than
// the passed in shard threshold.
// For each shard, the callback is called with the path range for the shard.
func shard(ss []stream, shardThreshold int64, f func(r *index.PathRange) error) {
	pq := newPriorityQueue(ss)
	var size int64
	pathRange := &index.PathRange{}
	for {
		ss, err := pq.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return nil, err
		}
		// Convert generic streams to file streams.
		var fileStreams []*fileStream
		for _, s := range ss {
			fileStreams = append(fileStreams, s.(*fileStream))
		}
		var hdr *index.Header
		var err error
		last := true
		for _, fs := range fileStreams {
			hdr, err = fs.r.Next()
			if err != nil {
				return err
			}
			size += hdr.Idx.SizeBytes
			if _, err = fs.r.Peek(); err != io.EOF {
				last = false
			}
		}
		if pathRange.Lower == "" {
			pathRange.Lower = hdr.Hdr.Name
		}
		// A shard is created when we have encountered more than shardThreshold content bytes.
		// The last shard is created when there are no more streams in the queue and the
		// streams being handled in this call are all io.EOF.
		if size > shardThreshold || len(next) == 0 && last {
			pathRange.Upper = hdr.Hdr.Name
			if err := f(pathRange); err != nil {
				return err
			}
			size = 0
			pathRange = &index.PathRange{}
			return nil
		}
		return nil
	}
}
