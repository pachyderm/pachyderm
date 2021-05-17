package server

import (
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/net/context"
)

// Source iterates over FileInfos generated from a fileset.FileSet
type Source interface {
	// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
	// during iteration, and the File.
	Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error
}

type source struct {
	commitInfo *pfs.CommitInfo
	fileSet    fileset.FileSet
	full       bool
}

// NewSource creates a Source which emits FileInfos with the information from commit, and the entries return from fileSet.
func NewSource(storage *fileset.Storage, commitInfo *pfs.CommitInfo, fs fileset.FileSet, opts ...SourceOption) Source {
	sc := &sourceConfig{}
	for _, opt := range opts {
		opt(sc)
	}
	if sc.full {
		fs = storage.NewIndexResolver(fs)
	}
	fs = fileset.NewDirInserter(fs)
	if sc.filter != nil {
		fs = sc.filter(fs)
	}
	return &source{
		commitInfo: commitInfo,
		fileSet:    fs,
		full:       sc.full,
	}
}

// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
// during iteration, and the File.
func (s *source) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	iter := fileset.NewIterator(ctx, s.fileSet)
	cache := make(map[string]*pfs.FileInfo)
	return s.fileSet.Iterate(ctx, func(f fileset.File) error {
		idx := f.Index()
		fi := &pfs.FileInfo{
			File:      client.NewFile(s.commitInfo.Commit.Branch.Repo.Name, s.commitInfo.Commit.Branch.Name, s.commitInfo.Commit.ID, idx.Path, idx.File.Tag),
			FileType:  pfs.FileType_FILE,
			Committed: s.commitInfo.Finished,
		}
		if fileset.IsDir(idx.Path) {
			fi.FileType = pfs.FileType_DIR
		}
		if s.full {
			cachedFi, ok := checkFileInfoCache(cache, idx)
			if ok {
				fi.SizeBytes = cachedFi.SizeBytes
				fi.Hash = cachedFi.Hash
			} else {
				computedFi, err := computeFileInfo(cache, iter, idx.Path)
				if err != nil {
					return err
				}
				fi.SizeBytes = computedFi.SizeBytes
				fi.Hash = computedFi.Hash
			}
		}
		// TODO: Figure out how to remove directory infos from cache when they are no longer needed.
		return cb(fi, f)
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

func computeFileInfo(cache map[string]*pfs.FileInfo, iter *fileset.Iterator, target string) (*pfs.FileInfo, error) {
	f, err := iter.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.Errorf("stream is done, can't compute hash for %s", target)
		}
		return nil, err
	}
	idx := f.Index()
	if idx.Path != target {
		return nil, errors.Errorf("stream is wrong place to compute hash for %s", target)
	}
	if !fileset.IsDir(idx.Path) {
		return computeRegularFileInfo(idx), nil
	}
	var size uint64
	h := pfs.NewHash()
	for {
		f2, err := iter.Peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		idx2 := f2.Index()
		if !strings.HasPrefix(idx2.Path, target) {
			break
		}
		childFi, err := computeFileInfo(cache, iter, idx2.Path)
		if err != nil {
			return nil, err
		}
		size += childFi.SizeBytes
		h.Write(childFi.Hash)
	}
	fi := &pfs.FileInfo{
		SizeBytes: size,
		Hash:      h.Sum(nil),
	}
	cache[target] = fi
	return fi, nil
}

func computeRegularFileInfo(idx *index.Index) *pfs.FileInfo {
	h := pfs.NewHash()
	for _, dataRef := range idx.File.DataRefs {
		h.Write([]byte(dataRef.Hash))
	}
	return &pfs.FileInfo{
		FileType:  pfs.FileType_FILE,
		SizeBytes: uint64(index.SizeBytes(idx)),
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

type emptySource struct{}

func (emptySource) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	return nil
}
