package fileset

import (
	"context"
	"io"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
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
		fileStreams = append(fileStreams, &fileStream{
			r:        r,
			priority: len(fileStreams),
		})
	}
	return &MergeReader{pq: newPriorityQueue(fileStreams)}
}

func (mr *MergeReader) iterate(cb func(*FileMergeReader) error) error {
	return mr.pq.iterate(func(ss []stream, _ ...string) error {
		return mr.mergeFile(ss, cb)
	})
}

func (mr *MergeReader) mergeFile(ss []stream, cb func(*FileMergeReader) error, deleter ...func(string, ...string)) (retErr error) {
	// Convert generic streams to file streams.
	var fss []*fileStream
	for _, s := range ss {
		fss = append(fss, s.(*fileStream))
	}
	// Progress each stream and collect each file reader.
	var i int
	var frs []*FileReader
	for ; i < len(fss); i++ {
		fr, err := fss[i].r.Next()
		if err != nil {
			return err
		}
		frs = append(frs, fr)
		// Filter out the file readers with lower priority than the highest priority file reader that contains a full file deletion.
		if isFullDelete(fr.Index()) {
			break
		}
	}
	// Progress each lower priority stream past the deleted file(s).
	for i += 1; i < len(fss); i++ {
		fr, err := fss[i].r.Next()
		if err != nil {
			return err
		}
		// If the full delete is for a directory, progress all lower priority streams past the directory.
		if IsDir(fr.Index().Path) {
			if err := fss[i].r.iterate(func(*FileReader) error { return nil }, DirUpperBound(fr.Index().Path)); err != nil {
				return err
			}
		}
	}
	contentTags, deleteTags := computeTags(frs)
	if len(contentTags) == 0 {
		if len(deleter) > 0 && len(deleteTags) > 0 {
			deleter[0](frs[0].Index().Path, deleteTags...)
		}
		return nil
	}
	// Create file merge reader and execute callback.
	return cb(newFileMergeReader(frs, sizeOfTags(contentTags)))
}

func isFullDelete(idx *index.Index) bool {
	return len(idx.DataOp.DeleteTags) > 0 && idx.DataOp.DeleteTags[0].Id == headerTag
}

func computeTags(frs []*FileReader) (map[string]int64, []string) {
	contentTags := make(map[string]int64)
	deleteTags := make(map[string]struct{})
	for i := len(frs) - 1; i >= 0; i-- {
		dataOp := frs[i].Index().DataOp
		for _, tag := range dataOp.DeleteTags {
			delete(contentTags, tag.Id)
			deleteTags[tag.Id] = struct{}{}
		}
		for _, dataRef := range dataOp.DataRefs {
			for _, tag := range dataRef.Tags {
				if tag.Id == headerTag || tag.Id == paddingTag {
					continue
				}
				contentTags[tag.Id] += tag.SizeBytes
			}
		}
	}
	return contentTags, getSortedKeys(deleteTags)
}

func sizeOfTags(tags map[string]int64) int64 {
	var size int64
	for _, tagSize := range tags {
		size += tagSize
	}
	return size
}

// Next gets the next file merge reader in the merge reader.
func (mr *MergeReader) Next() (*FileMergeReader, error) {
	for {
		var nextFmr *FileMergeReader
		ss, err := mr.pq.next()
		if err != nil {
			return nil, err
		}
		if err := mr.mergeFile(ss, func(fmr *FileMergeReader) error {
			nextFmr = fmr
			return nil
		}); err != nil {
			return nil, err
		}
		if nextFmr != nil {
			return nextFmr, nil
		}
	}
}

// WriteTo writes the merged fileset to the passed in fileset writer.
func (mr *MergeReader) WriteTo(w *Writer) error {
	return mr.pq.iterate(func(ss []stream, next ...string) error {
		// Copy single file stream.
		if len(ss) == 1 {
			return ss[0].(*fileStream).r.iterate(func(fr *FileReader) error {
				return w.CopyFile(fr)
			}, next...)
		}
		deleter := func(path string, tags ...string) {
			w.DeleteFile(path, tags...)
		}
		return mr.mergeFile(ss, func(fmr *FileMergeReader) error {
			return fmr.WriteTo(w)
		}, deleter)
	})
}

// Get writes the merged fileset.
func (mr *MergeReader) Get(w io.Writer) error {
	// Write a tar entry for each file merge reader.
	if err := mr.iterate(func(fmr *FileMergeReader) error {
		return fmr.Get(w)
	}); err != nil {
		return err
	}
	// Close a tar writer to create tar EOF padding.
	return tar.NewWriter(w).Close()
}

type fileStream struct {
	r        *Reader
	idx      *index.Index
	priority int
}

func (fs *fileStream) next() error {
	var err error
	fs.idx, err = fs.r.Peek()
	return err
}

func (fs *fileStream) key() string {
	return fs.idx.Path
}

func (fs *fileStream) streamPriority() int {
	return fs.priority
}

// FileMergeReader is an abstraction for reading a merged file.
type FileMergeReader struct {
	frs     []*FileReader
	hdr     *tar.Header
	tsmr    *TagSetMergeReader
	size    int64
	fullIdx *index.Index
}

func newFileMergeReader(frs []*FileReader, size int64) *FileMergeReader {
	return &FileMergeReader{
		frs:  frs,
		tsmr: newTagSetMergeReader(frs),
		size: size,
	}
}

// Index returns the index for the merged file.
func (fmr *FileMergeReader) Index() *index.Index {
	// If the full index has been computed, then return it.
	if fmr.fullIdx != nil {
		return fmr.fullIdx
	}
	idx := &index.Index{}
	idx.Path = fmr.frs[0].Index().Path
	idx.SizeBytes = fmr.size
	return idx
}

// Header returns the tar header for the merged file.
func (fmr *FileMergeReader) Header() (*tar.Header, error) {
	if fmr.hdr == nil {
		// TODO Validate the headers being merged?
		for _, fr := range fmr.frs {
			if len(fr.Index().DataOp.DataRefs) > 0 {
				_, err := fr.Header()
				if err != nil {
					return nil, err
				}
			}
		}
		// Use the header from the highest priority file
		// reader and update the size.
		// TODO Deep copy the header?
		for _, fr := range fmr.frs {
			if len(fr.Index().DataOp.DataRefs) > 0 {
				var err error
				fmr.hdr, err = fr.Header()
				if err != nil {
					return nil, err
				}
				fmr.hdr.Size = fmr.size
				break
			}
		}
	}
	return fmr.hdr, nil
}

// WriteTo writes the merged file to the passed in fileset writer.
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

// Get writes the merged file.
// TODO It might be cleaner to check if w is of type *Writer then use WriteTo rather than Get.
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

// Content writes the content of the current file excluding the header to w
func (fmr *FileMergeReader) Content(w io.Writer) error {
	return fmr.tsmr.Get(w)
}

// TagSetMergeReader returns the tagset merge reader for the file.
// This is how you would get just the data in the file (excludes the tar
// header and padding).
func (fmr *FileMergeReader) TagSetMergeReader() (*TagSetMergeReader, error) {
	if _, err := fmr.Header(); err != nil {
		return nil, err
	}
	return fmr.tsmr, nil
}

// TagSetMergeReader is an abstraction for reading the merged tagged data in a merged file.
type TagSetMergeReader struct {
	frs []*FileReader
	pq  *priorityQueue
}

func newTagSetMergeReader(frs []*FileReader) *TagSetMergeReader {
	var tagStreams []stream
	for i := len(frs) - 1; i >= 0; i-- {
		tagStreams = append(tagStreams, &deleteTagStream{
			tags:     frs[i].Index().DataOp.DeleteTags,
			priority: len(tagStreams),
		})
		tagStreams = append(tagStreams, &tagStream{
			fr:       frs[i],
			priority: len(tagStreams),
		})
	}
	return &TagSetMergeReader{
		frs: frs,
		pq:  newPriorityQueue(tagStreams),
	}
}

// Iterate iterates over the tag merge readers in the merged tagset.
func (tsmr *TagSetMergeReader) Iterate(cb func(*TagMergeReader) error) error {
	return tsmr.pq.iterate(func(ss []stream, _ ...string) error {
		return tsmr.mergeTag(ss, cb)
	})
}

func (tsmr *TagSetMergeReader) mergeTag(ss []stream, cb func(*TagMergeReader) error, deleter ...func(string)) (retErr error) {
	// Convert generic streams to tag streams.
	var i int
	var tss []*tagStream
	for ; i < len(ss); i++ {
		if dts, ok := ss[i].(*deleteTagStream); ok {
			if len(deleter) > 0 {
				deleter[0](dts.tag.Id)
			}
			break
		}
		tss = append(tss, ss[i].(*tagStream))
	}
	for i += 1; i < len(ss); i++ {
		if _, ok := ss[i].(*deleteTagStream); ok {
			continue
		}
		// TODO: This could be more performant.
		if err := ss[i].(*tagStream).fr.NextTagReader().Get(ioutil.Discard); err != nil {
			return err
		}
	}
	if len(tss) == 0 {
		return nil
	}
	var trs []*chunk.TagReader
	for _, ts := range tss {
		trs = append(trs, ts.fr.NextTagReader())
	}
	return cb(newTagMergeReader(trs))
}

// WriteTo writes the merged tagset to the passed in fileset writer.
func (tsmr *TagSetMergeReader) WriteTo(w *Writer) error {
	return tsmr.pq.iterate(func(ss []stream, next ...string) (retErr error) {
		deleter := func(tag string) {
			w.DeleteTag(tag)
		}
		// Copy single tag stream.
		if len(ss) == 1 {
			if dts, ok := ss[0].(*deleteTagStream); ok {
				deleter(dts.tag.Id)
				return nil
			}
			if len(next) == 0 {
				next = []string{paddingTag}
			}
			return ss[0].(*tagStream).fr.Iterate(func(dr *chunk.DataReader) error {
				return w.CopyTags(dr.BoundReader(next...))
			}, next...)
		}
		return tsmr.mergeTag(ss, func(tmr *TagMergeReader) error {
			return tmr.WriteTo(w)
		}, deleter)
	})
}

// Get writes the merged tagset.
func (tsmr *TagSetMergeReader) Get(w io.Writer) error {
	return tsmr.Iterate(func(tmr *TagMergeReader) error {
		return tmr.Get(w)
	})
}

type tagStream struct {
	fr       *FileReader
	tag      *chunk.Tag
	priority int
}

func (ts *tagStream) next() error {
	var err error
	ts.tag, err = ts.fr.PeekTag()
	if ts.tag != nil && ts.tag.Id == paddingTag {
		return io.EOF
	}
	return err
}

func (ts *tagStream) key() string {
	return ts.tag.Id
}

func (ts *tagStream) streamPriority() int {
	return ts.priority
}

type deleteTagStream struct {
	tags     []*chunk.Tag
	tag      *chunk.Tag
	priority int
}

func (dts *deleteTagStream) next() error {
	if len(dts.tags) == 0 {
		return io.EOF
	}
	dts.tag = dts.tags[0]
	dts.tags = dts.tags[1:]
	return nil
}

func (dts *deleteTagStream) key() string {
	return dts.tag.Id
}

func (dts *deleteTagStream) streamPriority() int {
	return dts.priority
}

// TagMergeReader is an abstraction for reading a merged tag.
// This abstraction is necessary because a tag in a file can appear
// across multiple filesets.
type TagMergeReader struct {
	trs []*chunk.TagReader
}

func newTagMergeReader(trs []*chunk.TagReader) *TagMergeReader {
	return &TagMergeReader{trs: trs}
}

// Iterate iterates over the data readers for the tagged data being merged.
func (tmr *TagMergeReader) Iterate(cb func(*chunk.DataReader) error) error {
	for _, tr := range tmr.trs {
		if err := tr.Iterate(cb); err != nil {
			return err
		}
	}
	return nil
}

// WriteTo writes the merged tagged data to the passed in fileset writer.
func (tmr *TagMergeReader) WriteTo(w *Writer) error {
	return tmr.Iterate(func(dr *chunk.DataReader) error {
		return w.CopyTags(dr)
	})
}

// Get writes the merged tagged data.
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
	if err := mr.iterate(func(fmr *FileMergeReader) error {
		// A shard is created when we have encountered more than shardThreshold content bytes.
		if size >= shardThreshold {
			pathRange.Upper = fmr.Index().Path
			if err := f(pathRange); err != nil {
				return err
			}
			size = 0
			pathRange = &index.PathRange{
				Lower: fmr.Index().Path,
			}
		}
		size += fmr.Index().SizeBytes
		return nil
	}); err != nil {
		return err
	}
	return f(pathRange)
}

type mergeSource struct {
	getReader func() (*MergeReader, error)
	s         *Storage
}

func (ms *mergeSource) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
	mr, err := ms.getReader()
	if err != nil {
		return err
	}
	for {
		f, err := mr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := cb(f); err != nil {
			return err
		}
	}
	return nil
}
