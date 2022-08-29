package server

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// Source iterates over FileInfos generated from a fileset.FileSet
type Source interface {
	// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
	// during iteration, and the File.
	Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error
}

type source struct {
	commitInfo                  *pfs.CommitInfo
	fileSet                     fileset.FileSet
	dirIndexOpts, fileIndexOpts []index.Option
}

// NewSource creates a Source which emits FileInfos with the information from commit, and the entries return from fileSet.
func NewSource(commitInfo *pfs.CommitInfo, fs fileset.FileSet, opts ...SourceOption) Source {
	sc := &sourceConfig{}
	for _, opt := range opts {
		opt(sc)
	}
	s := &source{
		commitInfo: commitInfo,
		fileSet:    fileset.NewDirInserter(fs, sc.prefix),
		dirIndexOpts: []index.Option{
			index.WithPrefix(sc.prefix),
			index.WithDatum(sc.datum),
		},
		fileIndexOpts: []index.Option{
			index.WithPrefix(sc.prefix),
			index.WithDatum(sc.datum),
		},
	}
	if sc.pathRange != nil {
		s.fileSet = fileset.NewDirInserter(fs, sc.pathRange.Lower)
		// The directory index options have no upper bound because the directory
		// may extend past the upper bound of the path range.
		s.dirIndexOpts = append(s.dirIndexOpts, index.WithRange(&index.PathRange{
			Lower: sc.pathRange.Lower,
		}))
		s.fileIndexOpts = append(s.fileIndexOpts, index.WithRange(&index.PathRange{
			Lower: sc.pathRange.Lower,
			Upper: sc.pathRange.Upper,
		}))
	}
	if sc.filter != nil {
		s.fileSet = sc.filter(s.fileSet)
	}
	return s
}

// Iterate calls cb for each File in the underlying fileset.FileSet, with a FileInfo computed
// during iteration, and the File.
func (s *source) Iterate(ctx context.Context, cb func(*pfs.FileInfo, fileset.File) error) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	iter := fileset.NewIterator(ctx, s.fileSet.Iterate, s.dirIndexOpts...)
	cache := make(map[string]*pfs.FileInfo)
	err := s.fileSet.Iterate(ctx, func(f fileset.File) error {
		idx := f.Index()
		file := s.commitInfo.Commit.NewFile(idx.Path)
		file.Datum = idx.File.Datum
		fi := &pfs.FileInfo{
			File:      file,
			FileType:  pfs.FileType_FILE,
			Committed: s.commitInfo.Finishing,
		}
		if fileset.IsDir(idx.Path) {
			fi.FileType = pfs.FileType_DIR
		}
		cachedFi, ok, err := s.checkFileInfoCache(ctx, cache, f)
		if err != nil {
			return err
		}
		if ok {
			fi.SizeBytes = cachedFi.SizeBytes
			fi.Hash = cachedFi.Hash
		} else {
			computedFi, err := s.computeFileInfo(ctx, cache, iter, idx.Path)
			if err != nil {
				return err
			}
			fi.SizeBytes = computedFi.SizeBytes
			fi.Hash = computedFi.Hash
		}
		// TODO: Figure out how to remove directory infos from cache when they are no longer needed.
		if err := cb(fi, f); err != nil {
			return errors.EnsureStack(err)
		}
		return nil
	}, s.fileIndexOpts...)
	return errors.EnsureStack(err)
}

func (s *source) checkFileInfoCache(ctx context.Context, cache map[string]*pfs.FileInfo, f fileset.File) (*pfs.FileInfo, bool, error) {
	idx := f.Index()
	// Handle a cached directory file info.
	fi, ok := cache[idx.Path]
	if ok {
		return fi, true, nil
	}
	// Handle a regular file info that has already been iterated through
	// when computing the parent directory file info.
	dir, _ := path.Split(idx.Path)
	_, ok = cache[dir]
	if ok {
		fi, err := s.computeRegularFileInfo(ctx, f)
		if err != nil {
			return nil, false, err
		}
		return fi, true, nil
	}
	return nil, false, nil
}

func (s *source) computeFileInfo(ctx context.Context, cache map[string]*pfs.FileInfo, iter *fileset.Iterator, target string) (*pfs.FileInfo, error) {
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
		return s.computeRegularFileInfo(ctx, f)
	}
	var size int64
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
		childFi, err := s.computeFileInfo(ctx, cache, iter, idx2.Path)
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

func (s *source) computeRegularFileInfo(ctx context.Context, f fileset.File) (*pfs.FileInfo, error) {
	fi := &pfs.FileInfo{
		SizeBytes: index.SizeBytes(f.Index()),
	}
	var err error
	fi.Hash, err = f.Hash(ctx)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	return fi, nil
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
		return errors.EnsureStack(err)
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

// checkSingleFile iterates through the source and returns errors for non-files, or multiple files.
// If the source contains a directory, then singleFile errors.
// If the source contains more than one file of any type, then singleFile errors
// the not exist error should be provided by the Source
func checkSingleFile(ctx context.Context, src Source) error {
	var count int
	err := src.Iterate(ctx, func(finfo *pfs.FileInfo, fsFile fileset.File) error {
		if finfo.FileType != pfs.FileType_FILE {
			return errors.EnsureStack(pfsserver.ErrMatchedNonFile)
		}
		if count == 0 {
			count++
			return nil
		}
		return errors.EnsureStack(pfsserver.ErrMatchedMultipleFiles)
	})
	return errors.EnsureStack(err)
}
