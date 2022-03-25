package fileset

import (
	"bytes"
	"context"
	"io"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type fileMap map[string]map[string]*file

type Buffer struct {
	// keys are file paths
	additive fileMap
	// keys are file paths or directories
	deletive fileMap

	// paths are directories
	appendCopies []*file
	truncCopies  []*file
}

type file struct {
	path  string
	datum string
	buf   *bytes.Buffer
	srcFS func(context.Context) (FileSet, error)
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(fileMap),
		deletive: make(fileMap),
	}
}

func (b *Buffer) CanCopyToDir(newPath string) bool {
	newPath = Clean(newPath, true)
	asFile := Clean(newPath, false)
	for _, m := range []fileMap{b.additive, b.deletive} {
		for path := range m {
			if strings.HasPrefix(path, newPath) || path == asFile {
				return false
			}
		}
	}
	for _, copies := range [][]*file{b.appendCopies, b.truncCopies} {
		for _, c := range copies {
			if strings.HasPrefix(newPath, c.path) || strings.HasPrefix(c.path, newPath) {
				return false
			}
		}
	}
	return true
}

func (b *Buffer) CanWriteToPath(newPath string) bool {
	newPath = Clean(newPath, true)
	for _, copies := range [][]*file{b.appendCopies, b.truncCopies} {
		for _, c := range copies {
			if strings.HasPrefix(newPath, c.path) {
				return false
			}
		}
	}
	return true
}

func (b *Buffer) Add(path, datum string) io.Writer {
	path = Clean(path, false)
	if _, ok := b.additive[path]; !ok {
		b.additive[path] = make(map[string]*file)
	}
	datumFiles := b.additive[path]
	if _, ok := datumFiles[datum]; !ok {
		datumFiles[datum] = &file{
			path:  path,
			datum: datum,
			buf:   &bytes.Buffer{},
		}
	}
	f := datumFiles[datum]
	return f.buf
}

func (b *Buffer) Delete(path, datum string) {
	path = Clean(path, IsDir(path))
	if IsDir(path) {
		// TODO: Linear scan for directory delete is less than ideal.
		// Fine for now since this should be rare and is an in-memory operation.
		for file := range b.additive {
			if strings.HasPrefix(file, path) {
				delete(b.additive, file)
			}
		}
		return
	}
	if datumFiles, ok := b.additive[path]; ok {
		delete(datumFiles, datum)
	}
	if _, ok := b.deletive[path]; !ok {
		b.deletive[path] = make(map[string]*file)
	}
	datumFiles := b.deletive[path]
	datumFiles[datum] = &file{
		path:  path,
		datum: datum,
	}
}

func (b *Buffer) Copy(path, datum string, withAppend bool, srcFS func(context.Context) (FileSet, error)) {
	copyFile := &file{
		path:  Clean(path, true),
		datum: datum,
		srcFS: srcFS,
	}
	if withAppend {
		b.appendCopies = append(b.appendCopies, copyFile)
	} else {
		b.truncCopies = append(b.truncCopies, copyFile)
	}
}

func (b *Buffer) WalkAdditive(ctx context.Context, onAdd func(path, datum string, r io.Reader) error, onCopy func(file File, datum string) error) error {
	for _, file := range sortFiles(b.additive, b.appendCopies, b.truncCopies) {
		if file.srcFS != nil {
			fs, err := file.srcFS(ctx)
			if err != nil {
				return err
			}
			if err := fs.Iterate(ctx, func(f File) error {
				return onCopy(f, file.datum)
			}); err != nil {
				return errors.EnsureStack(err)
			}
		} else if err := onAdd(file.path, file.datum, bytes.NewReader(file.buf.Bytes())); err != nil {
			return err
		}
	}
	return nil
}

func sortFiles(files map[string]map[string]*file, extras ...[]*file) []*file {
	var result []*file
	for _, datumFiles := range files {
		for _, f := range datumFiles {
			result = append(result, f)
		}
	}
	for _, fs := range extras {
		result = append(result, fs...)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].path == result[j].path {
			return result[i].datum < result[j].datum
		}
		return result[i].path < result[j].path
	})
	return result
}

func (b *Buffer) WalkDeletive(ctx context.Context, cb func(path, datum string) error) error {
	for _, file := range sortFiles(b.deletive, b.truncCopies) {
		if file.srcFS != nil {
			fs, err := file.srcFS(ctx)
			if err != nil {
				return err
			}
			if err := fs.Iterate(ctx, func(f File) error {
				return cb(f.Index().Path, file.datum)
			}); err != nil {
				return errors.EnsureStack(err)
			}
		} else if err := cb(file.path, file.datum); err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) Empty() bool {
	return len(b.additive) == 0 && len(b.deletive) == 0 && len(b.appendCopies) == 0 && len(b.truncCopies) == 0
}
