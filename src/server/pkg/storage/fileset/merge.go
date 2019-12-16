package fileset

import (
	"io"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// MergeReader merges a file's content that shows up across
// multiple fileset streams.
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

func (mr *MergeReader) Iterate(f func(*FileMergeReader) error) error {
	return mr.pq.iterate(func(ss []stream, _ ...string) error {
		// Convert generic streams to file streams.
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		// Create file merge reader and execute callback.
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
	return mr.pq.iterate(func(ss []stream, next ...string) error {
		// Convert generic streams to file streams.
		var fss []*fileStream
		for _, s := range ss {
			fss = append(fss, s.(*fileStream))
		}
		// Copy single file stream.
		if len(fss) == 1 {
			return fss[0].r.Iterate(func(fr *FileReader) error {
				return w.CopyFile(fr)
			}, next...)
		}
		// Create file merge reader and call its WriteTo function.
		var frs []*FileReader
		for _, fs := range fss {
			fr, err := fs.r.Next()
			if err != nil {
				return err
			}
			frs = append(frs, fr)
		}
		return newFileMergeReader(frs).WriteTo(w)
	})
}

func (mr *MergeReader) Get(w io.Writer) error {
	// Write a tar entry for each file merge reader.
	if err := mr.Iterate(func(fmr *FileMergeReader) error {
		return fmr.Get(w)
	}); err != nil {
		return err
	}
	// Close a tar writer to create tar EOF padding.
	return tar.NewWriter(w).Close()

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
	frs  []*FileReader
	hdr  *tar.Header
	tsmr *TagSetMergeReader
}

func newFileMergeReader(frs []*FileReader) *FileMergeReader {
	// (bryce) need to handle delete operations.
	return &FileMergeReader{
		frs:  frs,
		tsmr: newTagSetMergeReader(frs),
	}
}

func (fmr *FileMergeReader) Index() *index.Index {
	// (bryce) need to merge and compute new hashes when client
	// wants the stable hash for the content. Only returning
	// returning index with path and size for now.
	idx := &index.Index{}
	idx.Path = fmr.frs[0].Index().Path
	for _, fr := range fmr.frs {
		idx.SizeBytes += fr.Index().SizeBytes
	}
	return idx
}

func (fmr *FileMergeReader) Header() (*tar.Header, error) {
	if fmr.hdr == nil {
		// Compute the size of the headers being merged.
		var size int64
		for _, fr := range fmr.frs {
			hdr, err := fr.Header()
			if err != nil {
				return nil, err
			}
			size += hdr.Size
		}
		// Use the header from the highest priority file
		// reader and update the size.
		// (bryce) might want to copy the header?
		var err error
		fmr.hdr, err = fmr.frs[len(fmr.frs)-1].Header()
		if err != nil {
			return nil, err
		}
		fmr.hdr.Size = size
	}
	return fmr.hdr, nil
}

func (fmr *FileMergeReader) WriteTo(w *Writer) error {
	hdr, err := fmr.Header()
	if err != nil {
		return err
	}
	// Write merged header.
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	// Write merged content.
	return fmr.tsmr.WriteTo(w)
}

// (bryce) it might be cleaner to check if w is of type *Writer then use WriteTo rather than Get.
func (fmr *FileMergeReader) Get(w io.Writer) error {
	hdr, err := fmr.Header()
	if err != nil {
		return err
	}
	// Write merged header.
	tw := tar.NewWriter(w)
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	// Write merged content.
	if err := fmr.tsmr.Get(tw); err != nil {
		return err
	}
	return tw.Flush()
}

func (fmr *FileMergeReader) TagSetMergeReader() (*TagSetMergeReader, error) {
	if _, err := fmr.Header(); err != nil {
		return nil, err
	}
	return fmr.tsmr, nil
}

type TagSetMergeReader struct {
	frs []*FileReader
	hdr *tar.Header
	pq  *priorityQueue
}

func newTagSetMergeReader(frs []*FileReader) *TagSetMergeReader {
	var tagStreams []stream
	for _, fr := range frs {
		tagStreams = append(tagStreams, &tagStream{fr: fr})
	}
	return &TagSetMergeReader{
		frs: frs,
		pq:  newPriorityQueue(tagStreams),
	}
}

func (tsmr *TagSetMergeReader) Iterate(f func(*TagMergeReader) error) error {
	return tsmr.pq.iterate(func(ss []stream, _ ...string) error {
		// Convert generic stream to tag stream.
		var tss []*tagStream
		for _, s := range ss {
			tss = append(tss, s.(*tagStream))
		}
		// Check if we are at the padding tag, which indicates we are done.
		tag, err := tss[0].fr.PeekTag()
		if err != nil {
			return err
		}
		if tag.Id == paddingTag {
			for _, ts := range tss {
				if err := ts.fr.NextTagReader().Get(ioutil.Discard); err != nil {
					return err
				}
			}
			return nil
		}
		// Create tag merge reader and execute callback.
		var trs []*chunk.TagReader
		for _, ts := range tss {
			trs = append(trs, ts.fr.NextTagReader())
		}
		return f(newTagMergeReader(trs))
	})
}

func (tsmr *TagSetMergeReader) WriteTo(w *Writer) error {
	return tsmr.pq.iterate(func(ss []stream, next ...string) error {
		// Convert generic stream to tag stream.
		var tss []*tagStream
		for _, s := range ss {
			tss = append(tss, s.(*tagStream))
		}
		// Check if we are at the padding tag, which indicates we are done.
		tag, err := tss[0].fr.PeekTag()
		if err != nil {
			return err
		}
		if tag.Id == paddingTag {
			for _, ts := range tss {
				if err := ts.fr.NextTagReader().Get(ioutil.Discard); err != nil {
					return err
				}
			}
			return nil
		}
		// Copy single tag stream.
		if len(ss) == 1 {
			if len(next) == 0 {
				next = []string{paddingTag}
			}
			return tss[0].fr.Iterate(func(dr *chunk.DataReader) error {
				return w.CopyTags(dr.LimitReader(next...))
			}, next...)
		}
		// Create tag merge reader and call its WriteTo function.
		var trs []*chunk.TagReader
		for _, ts := range tss {
			trs = append(trs, ts.fr.NextTagReader())
		}
		return newTagMergeReader(trs).WriteTo(w)
	})
}

func (tsmr *TagSetMergeReader) Get(w io.Writer) error {
	return tsmr.Iterate(func(tmr *TagMergeReader) error {
		return tmr.Get(w)
	})
}

type tagStream struct {
	fr  *FileReader
	tag *chunk.Tag
}

func (ts *tagStream) next() error {
	var err error
	ts.tag, err = ts.fr.PeekTag()
	return err
}

func (ts *tagStream) key() string {
	return ts.tag.Id
}

type TagMergeReader struct {
	trs []*chunk.TagReader
}

func newTagMergeReader(trs []*chunk.TagReader) *TagMergeReader {
	return &TagMergeReader{trs: trs}
}

func (tmr *TagMergeReader) Iterate(f func(*chunk.DataReader) error) error {
	for _, tr := range tmr.trs {
		if err := tr.Iterate(f); err != nil {
			return err
		}
	}
	return nil
}

func (tmr *TagMergeReader) WriteTo(w *Writer) error {
	return tmr.Iterate(func(dr *chunk.DataReader) error {
		return w.CopyTags(dr)
	})
}

func (tmr *TagMergeReader) Get(w io.Writer) error {
	return tmr.Iterate(func(dr *chunk.DataReader) error {
		return dr.Iterate(func(_ *chunk.Tag, r io.Reader) error {
			_, err := io.Copy(w, r)
			return err
		})
	})
}

// ShardFunc is a callback that returns a PathRange for each shard.
type ShardFunc func(*index.PathRange) error

// shard creates shards (path ranges) from the fileset streams being merged.
// A shard is created when the size of the content for a path range is greater than
// the passed in shard threshold.
// For each shard, the callback is called with the path range for the shard.
func shard(mr *MergeReader, shardThreshold int64, f ShardFunc) error {
	var size int64
	pathRange := &index.PathRange{}
	if err := mr.Iterate(func(fmr *FileMergeReader) error {
		if pathRange.Lower == "" {
			pathRange.Lower = fmr.Index().Path
		}
		size += fmr.Index().SizeBytes
		// A shard is created when we have encountered more than shardThreshold content bytes.
		if size >= shardThreshold {
			pathRange.Upper = fmr.Index().Path
			if err := f(pathRange); err != nil {
				return err
			}
			size = 0
			pathRange = &index.PathRange{}
		}
		return nil
	}); err != nil {
		return err
	}
	// Edge case where the last shard is created at the last file.
	if pathRange.Lower == "" {
		return nil
	}
	return f(pathRange)
}
