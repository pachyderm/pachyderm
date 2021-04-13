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
	path  string
	parts map[string]*part
}

type part struct {
	tag string
	buf *bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(map[string]*file),
		deletive: make(map[string]*file),
	}
}

func (b *Buffer) Add(p, tag string) io.Writer {
	p = Clean(p, false)
	if _, ok := b.additive[p]; !ok {
		b.additive[p] = &file{
			path:  p,
			parts: make(map[string]*part),
		}
	}
	buf := &bytes.Buffer{}
	b.additive[p].parts[tag] = &part{
		tag: tag,
		buf: buf,
	}
	return buf
}

func (b *Buffer) Delete(p string, tag ...string) {
	p = Clean(p, IsDir(p))
	if IsDir(p) {
		// TODO: Linear scan for directory delete is less than ideal.
		// Fine for now since this should be rare and is an in-memory operation.
		for file := range b.additive {
			if strings.HasPrefix(file, p) {
				delete(b.additive, file)
			}
		}
	}
	if len(tag) == 0 {
		if _, ok := b.additive[p]; ok {
			delete(b.additive, p)
		}
		b.deletive[p] = &file{
			path: p,
		}
		return
	}
	if file, ok := b.additive[p]; ok {
		delete(file.parts, tag[0])
	}
	if _, ok := b.deletive[p]; !ok {
		b.deletive[p] = &file{
			path:  p,
			parts: make(map[string]*part),
		}
	}
	b.deletive[p].parts[tag[0]] = &part{tag: tag[0]}
}

func (b *Buffer) WalkAdditive(cb func(string, string, io.Reader) error) error {
	for _, file := range sortFiles(b.additive) {
		for _, part := range sortParts(file.parts) {
			if err := cb(file.path, part.tag, part.buf); err != nil {
				return err
			}
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

func sortParts(parts map[string]*part) []*part {
	var result []*part
	for _, p := range parts {
		result = append(result, p)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].tag < result[j].tag
	})
	return result
}

func (b *Buffer) WalkDeletive(cb func(string, ...string) error) error {
	for _, file := range sortFiles(b.deletive) {
		if len(file.parts) == 0 {
			if err := cb(file.path); err != nil {
				return err
			}
			continue
		}
		for _, part := range sortParts(file.parts) {
			if err := cb(file.path, part.tag); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Buffer) Empty() bool {
	return len(b.additive) == 0 && len(b.deletive) == 0
}
