package server

import (
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"golang.org/x/net/context"
)

// Source iterates over FileInfos generated from a fileset.Source
type Source interface {
	// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
	// during iteration, and the File.
	Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error
}

type source struct {
	commit    *pfs.Commit
	full      bool
	getReader func() fileset.FileSet
}

// NewSource creates a Source which emits FileInfos with the information from commit, and the entries from readers
// returned by getReader.  If getReader returns different Readers all bets are off.
func NewSource(commit *pfs.Commit, full bool, getReader func() fileset.FileSet) Source {
	return &source{
		commit:    commit,
		full:      full,
		getReader: getReader,
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
	cache := make(map[string]*pfs.FileInfo)
	return fs1.Iterate(ctx, func(fr fileset.File) error {
		idx := fr.Index()
		fi := &pfs.FileInfo{
			File: client.NewFile(s.commit.Repo.Name, s.commit.ID, idx.Path),
		}
		if s.full {
			cachedFi, ok := checkFileInfoCache(cache, idx)
			if ok {
				fi.FileType = cachedFi.FileType
				fi.SizeBytes = cachedFi.SizeBytes
				fi.Hash = cachedFi.Hash
			} else {
				computedFi, err := computeFileInfo(cache, s2, idx.Path)
				if err != nil {
					return err
				}
				fi.FileType = computedFi.FileType
				fi.SizeBytes = computedFi.SizeBytes
				fi.Hash = computedFi.Hash
			}

		}
		if err := cb(fi, fr); err != nil {
			return err
		}
		delete(cache, idx.Path)
		return nil
	})
}

func checkFileInfoCache(cache map[string]*pfs.FileInfo, idx *index.Index) (*pfs.FileInfo, bool) {
	// Handle a cached directory file info.
	fi, ok := cache[idx.Path]
	if ok {
		return fi, true
	}
	// Handle a regular file info that has already been iterated through
	// when computing the parent directory file info.
	dir, _ := path.Split(idx.Path)
	_, ok = cache[dir]
	if ok {
		return computeRegularFileInfo(idx), true
	}
	return nil, false
}

func computeFileInfo(cache map[string]*pfs.FileInfo, s *stream, target string) (*pfs.FileInfo, error) {
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
	if !indexIsDir(idx) {
		return computeRegularFileInfo(idx), nil
	}
	var size uint64
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
		childFi, err := computeFileInfo(cache, s, idx2.Path)
		if err != nil {
			return nil, err
		}
		size += childFi.SizeBytes
		h.Write(childFi.Hash)
	}
	fi := &pfs.FileInfo{
		FileType:  pfs.FileType_DIR,
		SizeBytes: size,
		Hash:      h.Sum(nil),
	}
	cache[target] = fi
	return fi, nil
}

func computeRegularFileInfo(idx *index.Index) *pfs.FileInfo {
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
	return &pfs.FileInfo{
		FileType:  pfs.FileType_FILE,
		SizeBytes: uint64(idx.SizeBytes),
		Hash:      h.Sum(nil),
	}
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
