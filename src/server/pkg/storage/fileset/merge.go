package fileset

import (
	"fmt"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type stream interface {
	next() error
	key() string
}

type mergeFunc func([]stream, ...string) error

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

func idxMergeFunc(w *Writer) mergeFunc {
	return func(ss []stream, next ...string) error {
		// (bryce) this will implement an index merge, which will be used by the distributed merge process.
		return nil
	}
}

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

type mergePQ struct {
	q    []stream
	f    mergeFunc
	size int
}

func (mq *mergePQ) key(i int) string {
	return mq.q[i].key()
}

func (mq *mergePQ) insert(s stream) error {
	// Get next in stream and insert it.
	if err := s.next(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	mq.q[mq.size+1] = s
	mq.size++
	// Propagate insert up the queue
	i := mq.size
	for i > 1 {
		if strings.Compare(mq.key(i/2), mq.key(i)) <= 0 {
			break
		}
		mq.swap(i/2, i)
		i /= 2
	}
	return nil
}

func (mq *mergePQ) next() []stream {
	ss := []stream{mq.q[1]}
	mq.fill()
	// Keep popping streams off the queue if they have the same key.
	for mq.q[1] != nil && strings.Compare(mq.key(1), ss[0].key()) == 0 {
		ss = append(ss, mq.q[1])
		mq.fill()
	}
	return ss
}

func (mq *mergePQ) peek() string {
	return mq.q[1].key()
}

func (mq *mergePQ) fill() {
	// Replace first stream with last
	mq.q[1] = mq.q[mq.size]
	mq.q[mq.size] = nil
	mq.size--
	// Propagate last stream down the queue
	i := 1
	var next int
	for {
		l, r := i*2, i*2+1
		if l > mq.size {
			break
		} else if r > mq.size || strings.Compare(mq.key(l), mq.key(r)) <= 0 {
			next = l
		} else {
			next = r
		}
		if strings.Compare(mq.key(i), mq.key(next)) <= 0 {
			break
		}
		mq.swap(i, next)
		i = next
	}
}

func (mq *mergePQ) swap(i, j int) {
	mq.q[i], mq.q[j] = mq.q[j], mq.q[i]
}

func merge(ss []stream, f mergeFunc) error {
	if len(ss) == 0 {
		return nil
	}
	mq := &mergePQ{q: make([]stream, len(ss)+1)}
	// Insert streams.
	for _, s := range ss {
		if err := mq.insert(s); err != nil {
			return err
		}
	}
	for mq.q[1] != nil {
		// Get next streams and merge them.
		ss := mq.next()
		var next []string
		if mq.q[1] != nil {
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
