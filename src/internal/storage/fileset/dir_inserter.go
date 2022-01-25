package fileset

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

type dirInserter struct {
	x FileSet
}

// NewDirInserter creates a file set that inserts directory entries.
func NewDirInserter(x FileSet) FileSet {
	return &dirInserter{x: x}
}

// Iterate calls cb once for every file in lexicographical order by path
func (s *dirInserter) Iterate(ctx context.Context, cb func(File) error, _ ...bool) error {
	lastPath := ""
	var emit func(p string, f File) error
	emit = func(p string, f File) error {
		parent := Clean(parentOf(p), true)
		if p == "/" || lastPath >= parent {
			if err := cb(f); err != nil {
				return err
			}
			lastPath = p
			return nil
		}
		// need to create entry for parent
		df := dirFile{
			path: parent,
		}
		if err := emit(parent, df); err != nil {
			return err
		}
		return emit(p, f)
	}
	return errors.EnsureStack(s.x.Iterate(ctx, func(f File) error {
		return emit(f.Index().Path, f)
	}))
}

func parentOf(x string) string {
	x = strings.TrimRight(x, "/")
	y := path.Dir(x)
	if y == x {
		return "/"
	}
	if !IsDir(y) {
		y += "/"
	}
	return y
}

type dirFile struct {
	path string
}

func (d dirFile) Index() *index.Index {
	return &index.Index{
		Path: d.path,
		File: &index.File{},
	}
}

func (d dirFile) Content(_ context.Context, _ io.Writer, _ ...chunk.ReaderOption) error {
	return nil
}

func (d dirFile) Hash(_ context.Context) ([]byte, error) {
	// TODO: It may make sense to move the generation of directory metadata (size / hash) into the directory inserter.
	panic("we should not be using the Hash function for dirFile, this is a bug")
}
