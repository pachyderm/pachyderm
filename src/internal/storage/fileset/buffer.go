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
	path string
	tag  string
	buf  *bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(map[string]map[string]*file),
		deletive: make(map[string]map[string]*file),
	}
}

func (b *Buffer) Add(path, tag string) io.Writer {
	path = Clean(path, false)
	if _, ok := b.additive[path]; !ok {
		b.additive[path] = make(map[string]*file)
	}
	taggedFiles := b.additive[path]
	if _, ok := taggedFiles[tag]; !ok {
		taggedFiles[tag] = &file{
			path: path,
			tag:  tag,
			buf:  &bytes.Buffer{},
		}
	}
	f := taggedFiles[tag]
	return f.buf
}

func (b *Buffer) Delete(path, tag string) {
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
	if taggedFiles, ok := b.additive[path]; ok {
		delete(taggedFiles, tag)
	}
	if _, ok := b.deletive[path]; !ok {
		b.deletive[path] = make(map[string]*file)
	}
	taggedFiles := b.deletive[path]
	taggedFiles[tag] = &file{
		path: path,
		tag:  tag,
	}
}

func (b *Buffer) WalkAdditive(cb func(path, tag string, r io.Reader) error) error {
	for _, file := range sortFiles(b.additive) {
		if err := cb(file.path, file.tag, bytes.NewReader(file.buf.Bytes())); err != nil {
			return err
		}
	}
	return nil
}

func sortFiles(files map[string]map[string]*file) []*file {
	var result []*file
	for _, taggedFiles := range files {
		for _, f := range taggedFiles {
			result = append(result, f)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].path == result[j].path {
			return result[i].tag < result[j].tag
		}
		return result[i].path < result[j].path
	})
	return result
}

func (b *Buffer) WalkDeletive(cb func(path, tag string) error) error {
	for _, file := range sortFiles(b.deletive) {
		if err := cb(file.path, file.tag); err != nil {
			return err
		}
	}
	return nil
}

func (b *Buffer) Empty() bool {
	return len(b.additive) == 0 && len(b.deletive) == 0
}
