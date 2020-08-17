package server

import (
	"hash"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"golang.org/x/net/context"
)

// FileReader is a PFS wrapper for a fileset.MergeReader.
// The primary purpose of this abstraction is to convert from index.Index to
// pfs.FileInfo and to convert a set of index hashes to a file hash.
type FileReader struct {
	file      *pfs.File
	idx       *index.Index
	fmr       *fileset.FileMergeReader
	mr        *fileset.MergeReader
	fileCount int
	hash      hash.Hash
}

func newFileReader(file *pfs.File, idx *index.Index, fmr *fileset.FileMergeReader, mr *fileset.MergeReader) *FileReader {
	h := pfs.NewHash()
	for _, dataRef := range idx.DataOp.DataRefs {
		// TODO Pull from chunk hash.
		h.Write([]byte(dataRef.Hash))
	}
	return &FileReader{
		file: file,
		idx:  idx,
		fmr:  fmr,
		mr:   mr,
		hash: h,
	}
}

func (fr *FileReader) updateFileInfo(idx *index.Index) {
	fr.fileCount++
	for _, dataRef := range idx.DataOp.DataRefs {
		fr.hash.Write([]byte(dataRef.Hash))
	}
}

// Info returns the info for the file.
func (fr *FileReader) Info() *pfs.FileInfo {
	return &pfs.FileInfo{
		File: fr.file,
		Hash: fr.hash.Sum(nil),
	}
}

// Get writes a tar stream that contains the file.
func (fr *FileReader) Get(w io.Writer, noPadding ...bool) error {
	if err := fr.fmr.Get(w); err != nil {
		return err
	}
	for fr.fileCount > 0 {
		fmr, err := fr.mr.Next()
		if err != nil {
			return err
		}
		if err := fmr.Get(w); err != nil {
			return err
		}
		fr.fileCount--
	}
	if len(noPadding) > 0 && noPadding[0] {
		return nil
	}
	// Close a tar writer to create tar EOF padding.
	return tar.NewWriter(w).Close()
}

func (fr *FileReader) drain() error {
	for fr.fileCount > 0 {
		if _, err := fr.mr.Next(); err != nil {
			return err
		}
		fr.fileCount--
	}
	return nil
}

// Source iterates over FileInfos generated from a fileset.Source
type Source interface {
	// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
	// during iteration, and the File.
	Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error
}

type source struct {
	commit        *pfs.Commit
	getReader     func() fileset.FileSet
	computeHashes bool
}

// NewSource creates a Source which emits FileInfos with the information from commit, and the entries from readers
// returned by getReader.  If getReader returns different Readers all bets are off.
func NewSource(commit *pfs.Commit, computeHashes bool, getReader func() fileset.FileSet) Source {
	return &source{
		commit:        commit,
		getReader:     getReader,
		computeHashes: computeHashes,
	}
}

// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
// during iteration, and the File.
func (s *source) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	fs1 := s.getReader()
	fs2 := s.getReader()
	s2 := newStream(ctx, fs2)
	cache := make(map[string][]byte)
	return fs1.Iterate(ctx, func(fr fileset.File) error {
		idx := fr.Index()
		fi := &pfs.FileInfo{
			File: client.NewFile(s.commit.Repo.Name, s.commit.ID, idx.Path),
		}
		if s.computeHashes {
			var err error
			var hashBytes []byte
			if indexIsDir(idx) {
				hashBytes, err = computeHash(cache, s2, idx.Path)
				if err != nil {
					return err
				}
			} else {
				hashBytes = computeFileHash(idx)
			}
			fi.Hash = hashBytes
		}
		if err := cb(fi, fr); err != nil {
			return err
		}
		delete(cache, idx.Path)
		return nil
	})
}

func computeHash(cache map[string][]byte, s *stream, target string) ([]byte, error) {
	if hashBytes, exists := cache[target]; exists {
		return hashBytes, nil
	}
	// consume the target from the stream
	fr, err := s.Next()
	if err != nil {
		if err == io.EOF {
			return nil, errors.Errorf("stream is done, can't compute hash for %s", target)
		}
		return nil, err
	}
	idx := fr.Index()
	if idx.Path != target {
		return nil, errors.Errorf("stream is wrong place to compute hash for %s", target)
	}
	// for file
	if !indexIsDir(idx) {
		return computeFileHash(idx), nil
	}
	// for directory
	h := pfs.NewHash()
	for {
		f2, err := s.Peek()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		idx2 := f2.Index()
		if !strings.HasPrefix(idx2.Path, target) {
			break
		}
		childHash, err := computeHash(cache, s, idx2.Path)
		if err != nil {
			return nil, err
		}
		h.Write(childHash)
	}
	hashBytes := h.Sum(nil)
	cache[target] = hashBytes
	return hashBytes, nil
}

func computeFileHash(idx *index.Index) []byte {
	h := pfs.NewHash()
	if idx.DataOp != nil {
		for _, dataRef := range idx.DataOp.DataRefs {
			if dataRef.Hash == "" {
				h.Write([]byte(dataRef.ChunkInfo.Chunk.Hash))
			} else {
				h.Write([]byte(dataRef.Hash))
			}
		}
	}
	return h.Sum(nil)
}

type errOnEmpty struct {
	source Source
	err    error
}

// NewErrOnEmpty causes iterate to return a not found error if there are no items to iterate over
func NewErrOnEmpty(s Source, err error) Source {
	return &errOnEmpty{source: s, err: err}
}

// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
// during iteration, and the File.
func (s *errOnEmpty) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	empty := true
	if err := s.source.Iterate(ctx, func(fi *pfs.FileInfo, f fileset.File) error {
		empty = false
		return cb(fi, f)
	}); err != nil {
		return err
	}
	if empty {
		return s.err
	}
	return nil
}

type stream struct {
	peek     fileset.File
	fileChan chan fileset.File
	errChan  chan error
}

func newStream(ctx context.Context, source fileset.FileSet) *stream {
	fileChan := make(chan fileset.File)
	errChan := make(chan error, 1)
	go func() {
		if err := source.Iterate(ctx, func(file fileset.File) error {
			fileChan <- file
			return nil
		}); err != nil {
			errChan <- err
			return
		}
		close(fileChan)
	}()
	return &stream{
		fileChan: fileChan,
		errChan:  errChan,
	}
}

func (s *stream) Peek() (fileset.File, error) {
	if s.peek != nil {
		return s.peek, nil
	}
	var err error
	s.peek, err = s.Next()
	return s.peek, err
}

func (s *stream) Next() (fileset.File, error) {
	if s.peek != nil {
		tmp := s.peek
		s.peek = nil
		return tmp, nil
	}
	select {
	case file, more := <-s.fileChan:
		if !more {
			return nil, io.EOF
		}
		return file, nil
	case err := <-s.errChan:
		return nil, err
	}
}

func indexIsDir(idx *index.Index) bool {
	return strings.HasSuffix(idx.Path, "/")
}

type emptySource struct{}

func (emptySource) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	return nil
}
