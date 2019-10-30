package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Reader reads the serialized format of a file set.
type Reader struct {
	ctx context.Context
	ir  *index.Reader
	cr  *chunk.Reader
}

func newReader(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string, opts ...index.Option) *Reader {
	cr := chunks.NewReader(ctx)
	return &Reader{
		ctx: ctx,
		ir:  index.NewReader(ctx, objC, chunks, path, opts...),
		cr:  cr,
	}
}

func (r *Reader) Peek() (*index.Index, error) {
	return r.ir.Peek()
}

func (r *Reader) Read(data []byte) (int, error) {
	// (bryce) no-op for now, will read the serialized tar stream.
	return 0, nil
}

func (r *Reader) Iterate(f func(*FileReader) error, pathBound ...string) error {
	return r.ir.Iterate(func(idx *index.Index) error {
		r.cr.NextRange(idx.DataOp.DataRefs)
		return f(&FileReader{
			idx: idx,
			cr:  r.cr,
		})
	}, pathBound...)
}

type FileReader struct {
	idx *index.Index
	cr  *chunk.Reader
}

func newFileReader(idx *index.Index, cr *chunk.Reader) *FileReader {
	return &FileReader{
		idx: idx,
		cr:  cr,
	}
}

func (fr *FileReader) Index() *index.Index {
	return fr.idx
}

func (fr *FileReader) Read(data []byte) (int, error) {
	return fr.cr.Read(data)
}

func (fr *FileReader) TagReader() *TagReader {
	return newTagReader(fr.idx.DataOp.Tags, fr.cr)
}

type TagReader struct {
	tags []*index.Tag
	cr   *chunk.Reader
	hdr  *tar.Header
}

func newTagReader(tags []*index.Tag, cr *chunk.Reader) *TagReader {
	return &TagReader{
		tags: tags,
		cr:   cr,
	}
}

func (tr *TagReader) Header() (*tar.Header, error) {
	if tr.hdr == nil {
		var err error
		tr.hdr, err = tar.NewReader(tr.cr).Next()
		if err != nil {
			return nil, err
		}
		tr.tags = tr.tags[1:]
	}
	return tr.hdr, nil
}

func (tr *TagReader) Peek() (*index.Tag, error) {
	if _, err := tr.Header(); err != nil {
		return nil, err
	}
	if len(tr.tags) == 0 {
		return nil, io.EOF
	}
	return tr.tags[0], nil
}

func (tr *TagReader) Iterate(f func(*index.Tag, io.Reader) error, tagBound ...string) error {
	for {
		tag, err := tr.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !index.BeforeBound(tag.Id, tagBound...) {
			return nil
		}
		if err := f(tag, io.LimitReader(tr.cr, tag.SizeBytes)); err != nil {
			return err
		}
		tr.tags = tr.tags[1:]
	}
}

// (bryce) commented out for now to checkpoint changes to reader/writer
//type copyFiles struct {
//	indexCopyF func() (*index.Copy, error)
//	files      []*copyTags
//}
//
//func (r *Reader) readCopyFiles(pathBound ...string) func() (*copyFiles, error) {
//	var done bool
//	return func() (*copyFiles, error) {
//		if done {
//			return nil, io.EOF
//		}
//		c := &copyFiles{}
//		// Setup index copying callback at the first split point where
//		// the next chunk will be copied.
//		var hdr *index.Header
//		r.cr.OnSplit(func() {
//			if hdr != nil && index.BeforeBound(hdr.Idx.LastPathChunk, pathBound...) {
//				c.indexCopyF = r.ir.ReadCopyFunc(pathBound...)
//			}
//		})
//		for {
//			var err error
//			hdr, err = r.Peek()
//			if err != nil {
//				if err == io.EOF {
//					done = true
//					break
//				}
//				return nil, err
//			}
//			// If the header is past the path bound, then we are done.
//			if !index.BeforeBound(hdr.Hdr.Name, pathBound...) {
//				done = true
//				break
//			}
//			if _, err := r.Next(); err != nil {
//				return nil, err
//			}
//			file, err := r.readCopyTags()
//			if err != nil {
//				return nil, err
//			}
//			c.files = append(c.files, file)
//			// Return copy information for content level, and callback
//			// for index level copying.
//			if c.indexCopyF != nil {
//				break
//			}
//		}
//		return c, nil
//	}
//}
//
//type copyTags struct {
//	hdr     *index.Header
//	content *chunk.Copy
//	tags    []*index.Tag
//}
//
//func (r *Reader) readCopyTags(tagBound ...string) (*copyTags, error) {
//	// Lazily setup reader for underlying file.
//	if err := r.setupReader(); err != nil {
//		return nil, err
//	}
//	// Determine the tags and number of bytes to copy.
//	var idx int
//	var numBytes int64
//	for i, tag := range r.tags {
//		if !index.BeforeBound(tag.Id, tagBound...) {
//			break
//		}
//		idx = i + 1
//		numBytes += tag.SizeBytes
//
//	}
//	// Setup copy struct.
//	c := &copyTags{hdr: r.hdr}
//	var err error
//	c.content, err = r.cr.ReadCopy(numBytes)
//	if err != nil {
//		return nil, err
//	}
//	c.tags = r.tags[:idx]
//	// Update reader state.
//	r.tags = r.tags[idx:]
//	if err := r.tr.Skip(numBytes); err != nil {
//		return nil, err
//	}
//	return c, nil
//}
