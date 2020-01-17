package fileset

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type stream interface {
	next() error
	key() string
}

// mergeFunc is a function that merges one or more file set streams.
// ss is the set of file set streams that are at the same key
// in the merge process.
// next is the next key that will be merged.
type mergeFunc func(ss []stream, next ...string) error

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

// contentMergeFunc merges a file's content that shows up across
// multiple file set streams.
// A file's content is ordered based on the lexicographical order of
// the tagged content, so the output file content is produced by
// performing a merge of the tagged content.
func contentMergeFunc(w *Writer) mergeFunc {
	return func(ss []stream, next ...string) error {
		// Convert generic streams to file streams.
		var fileStreams []*fileStream
		for _, s := range ss {
			fileStreams = append(fileStreams, s.(*fileStream))
		}
		// Fast path for copying files from one stream.
		if len(fileStreams) == 1 {
			if err := w.CopyFiles(fileStreams[0].r, next...); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			return nil
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
		// Write header for file.
		hdr := &index.Header{
			Hdr: &tar.Header{
				Name: fileStreams[0].hdr.Hdr.Name,
				Size: size,
			},
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		// Merge file content.
		return merge(tagStreams, tagMergeFunc(w))
	}
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

// tageMergeFunc merges the tagged content in a file.
// Tags in a file should be unique across file sets being merged,
// so the operation is a simple copy.
func tagMergeFunc(w *Writer) mergeFunc {
	return func(ss []stream, next ...string) error {
		// (bryce) this should be an Internal error type.
		if len(ss) > 1 {
			return fmt.Errorf("tags should be distinct within a file")
		}
		// Convert generic stream to tag stream.
		tagStream := ss[0].(*tagStream)
		// Copy tagged data to writer.
		return w.CopyTags(tagStream.r, next...)
	}
}

// shardMergeFunc creates shards (path ranges) from the file set streams being merged.
// A shard is created when the size of the content for a path range is greater than
// the passed in shard threshold.
// For each shard, the callback is called with the path range for the shard.
func shardMergeFunc(shardThreshold int64, f func(r *index.PathRange) error) mergeFunc {
	var size int64
	pathRange := &index.PathRange{}
	return func(ss []stream, next ...string) error {
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

type mergePriorityQueue struct {
	queue []stream
	size  int
}

func (mq *mergePriorityQueue) key(i int) string {
	return mq.queue[i].key()
}

func (mq *mergePriorityQueue) insert(s stream) error {
	// Get next in stream and insert it.
	if err := s.next(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	mq.queue[mq.size+1] = s
	mq.size++
	// Propagate insert up the queue
	i := mq.size
	for i > 1 {
		if mq.key(i/2) <= mq.key(i) {
			break
		}
		mq.swap(i/2, i)
		i /= 2
	}
	return nil
}

func (mq *mergePriorityQueue) next() []stream {
	ss := []stream{mq.queue[1]}
	mq.fill()
	// Keep popping streams off the queue if they have the same key.
	for mq.queue[1] != nil && mq.key(1) == ss[0].key() {
		ss = append(ss, mq.queue[1])
		mq.fill()
	}
	return ss
}

func (mq *mergePriorityQueue) peek() string {
	return mq.queue[1].key()
}

func (mq *mergePriorityQueue) fill() {
	// Replace first stream with last
	mq.queue[1] = mq.queue[mq.size]
	mq.queue[mq.size] = nil
	mq.size--
	// Propagate last stream down the queue
	i := 1
	var next int
	for {
		left, right := i*2, i*2+1
		if left > mq.size {
			break
		} else if right > mq.size || mq.key(left) <= mq.key(right) {
			next = left
		} else {
			next = right
		}
		if mq.key(i) <= mq.key(next) {
			break
		}
		mq.swap(i, next)
		i = next
	}
}

func (mq *mergePriorityQueue) swap(i, j int) {
	mq.queue[i], mq.queue[j] = mq.queue[j], mq.queue[i]
}

func merge(ss []stream, f mergeFunc) error {
	if len(ss) == 0 {
		return nil
	}
	mq := &mergePriorityQueue{queue: make([]stream, len(ss)+1)}
	// Insert streams.
	for _, s := range ss {
		if err := mq.insert(s); err != nil {
			return err
		}
	}
	for mq.queue[1] != nil {
		// Get next streams and merge them.
		ss := mq.next()
		var next []string
		if mq.queue[1] != nil {
			next = append(next, mq.peek())
		}
		if err := f(ss, next...); err != nil {
			return err
		}
		// Re-insert streams
		for _, s := range ss {
			if err := mq.insert(s); err != nil {
				return err
			}
		}
	}
	return nil
}
