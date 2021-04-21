package fileset

import (
	"bytes"
	"io"
	"sort"
	"strings"
)

type Buffer struct {
	additive map[string]*file
	deletive map[string]*file
}

type file struct {
	path string
	tag  string
	buf  *bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(map[string]*file),
		deletive: make(map[string]*file),
	}
}

func (b *Buffer) Add(path string, tag ...string) io.Writer {
	path = Clean(path, false)
	if _, ok := b.additive[path]; !ok {
		b.additive[path] = &file{
			path: path,
			buf:  &bytes.Buffer{},
		}
	}
	f := b.additive[path]
	if len(tag) > 0 {
		f.tag = tag[0]
	}
	return f.buf
}

func (b *Buffer) Delete(path string) {
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
	delete(b.additive, path)
	b.deletive[path] = &file{path: path}
}

func (b *Buffer) WalkAdditive(cb func(string, io.Reader, ...string) error) error {
	for _, file := range sortFiles(b.additive) {
		if err := cb(file.path, bytes.NewReader(file.buf.Bytes()), file.tag); err != nil {
			return err
		}
	}
	return nil
}

func sortFiles(files map[string]*file) []*file {
	var result []*file
	for _, f := range files {
		result = append(result, f)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].path < result[j].path
	})
	return result
}

func (b *Buffer) WalkDeletive(cb func(string) error) error {
	for _, file := range sortFiles(b.deletive) {
		if err := cb(file.path); err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) Empty() bool {
	return len(b.additive) == 0 && len(b.deletive) == 0
}
