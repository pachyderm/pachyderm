package fileset

import (
	"bytes"
	"io"
	"sort"
	"strings"
)

type Buffer struct {
	additive map[string]map[string]*file
	deletive map[string]map[string]*file
}

type file struct {
	path  string
	datum string
	buf   *bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(map[string]map[string]*file),
		deletive: make(map[string]map[string]*file),
	}
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

func (b *Buffer) WalkAdditive(cb func(path, datum string, r io.Reader) error) error {
	for _, file := range sortFiles(b.additive) {
		if err := cb(file.path, file.datum, bytes.NewReader(file.buf.Bytes())); err != nil {
			return err
		}
	}
	return nil
}

func sortFiles(files map[string]map[string]*file) []*file {
	var result []*file
	for _, datumFiles := range files {
		for _, f := range datumFiles {
			result = append(result, f)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].path == result[j].path {
			return result[i].datum < result[j].datum
		}
		return result[i].path < result[j].path
	})
	return result
}

func (b *Buffer) WalkDeletive(cb func(path, datum string) error) error {
	for _, file := range sortFiles(b.deletive) {
		if err := cb(file.path, file.datum); err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) Empty() bool {
	return len(b.additive) == 0 && len(b.deletive) == 0
}
