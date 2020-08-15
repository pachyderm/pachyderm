package fileset

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type dirInserter struct {
	x FileSource
}

// NewDirInserter returns a FileSource which will include all directories on the path
// from the root to a leaf (regular file).
func NewDirInserter(x FileSource) FileSource {
	return &dirInserter{x: x}
}

// Iterate calls cb once for every file in lexicographical order by path
func (s *dirInserter) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
	lastPath := ""
	var emit func(p string, f File) error
	emit = func(p string, f File) error {
		parent := CleanTarPath(parentOf(p), true)
		if p == "/" || lastPath >= parent {
			if err := cb(f); err != nil {
				return err
			}
			lastPath = p
			return nil
		}
		// need to create entry for parent
		df := dirFile{
			hdr: &tar.Header{
				Name: parent,
			},
		}
		if err := emit(parent, df); err != nil {
			return err
		}
		return emit(p, f)
	}
	return s.x.Iterate(ctx, func(f File) error {
		hdr, err := f.Header()
		if err != nil {
			return err
		}
		return emit(hdr.Name, f)
	}, stopBefore...)
}

func parentOf(x string) string {
	x = strings.TrimRight(x, "/")
	y := path.Dir(x)
	if y == x {
		return "/"
	}
	if !strings.HasSuffix(y, "/") {
		y += "/"
	}
	return y
}

type dirFile struct {
	hdr *tar.Header
}

func (d dirFile) Index() *index.Index {
	return nil
}

func (d dirFile) Header() (*tar.Header, error) {
	return d.hdr, nil
}

func (d dirFile) Content(w io.Writer) error {
	return nil
}
