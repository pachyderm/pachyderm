package fileset

import (
	"bytes"
	"io"
	"sort"
	"strings"
)

type Buffer struct {
	additive  map[string]map[string]*file
	deletive  map[string]map[string]*file
	fileCount int // the number of files stored in the additive and deletive maps
}

type file struct {
	path     string
	datum    string
	contents []fileContent
}

// contents are either raw bytes to be appended or an existing file to be copied
// Exactly one will be non-nil
type fileContent struct {
	buf  *bytes.Buffer
	copy File
}

func NewBuffer() *Buffer {
	return &Buffer{
		additive: make(map[string]map[string]*file),
		deletive: make(map[string]map[string]*file),
	}
}

func (b *Buffer) add(path, datum string) *file {
	path = Clean(path, false)
	if _, ok := b.additive[path]; !ok {
		b.additive[path] = make(map[string]*file)
	}
	datumFiles := b.additive[path]
	if _, ok := datumFiles[datum]; !ok {
		datumFiles[datum] = &file{
			path:  path,
			datum: datum,
		}
		b.fileCount++
	}
	return datumFiles[datum]
}

func (b *Buffer) Add(path, datum string) io.Writer {
	f := b.add(path, datum)
	if len(f.contents) > 0 && f.contents[len(f.contents)-1].copy == nil {
		return f.contents[len(f.contents)-1].buf
	}
	buf := &bytes.Buffer{}
	f.contents = append(f.contents, fileContent{buf: buf})
	return buf
}

func (b *Buffer) Delete(path, datum string) {
	path = Clean(path, IsDir(path))
	if IsDir(path) {
		// TODO: Linear scan for directory delete is less than ideal.
		// Fine for now since this should be rare and is an in-memory operation.
		for file, datumFiles := range b.additive {
			if strings.HasPrefix(file, path) {
				delete(b.additive, file)
				b.fileCount -= len(datumFiles)
			}
		}
		return
	}
	if _, ok := b.additive[path][datum]; ok {
		delete(b.additive[path], datum)
		b.fileCount--
	}
	if _, ok := b.deletive[path]; !ok {
		b.deletive[path] = make(map[string]*file)
	}
	datumFiles := b.deletive[path]
	datumFiles[datum] = &file{
		path:  path,
		datum: datum,
	}
	b.fileCount++
}

func (b *Buffer) Copy(file File, datum string) {
	f := b.add(file.Index().Path, datum)
	f.contents = append(f.contents, fileContent{copy: file})
}

func (b *Buffer) WalkAdditive(onAdd func(path, datum string, r io.Reader) error, onCopy func(file File, datum string) error) error {
	for _, file := range sortFiles(b.additive) {
		for _, content := range file.contents {
			if content.copy != nil {
				if err := onCopy(content.copy, file.datum); err != nil {
					return err
				}
			} else if err := onAdd(file.path, file.datum, bytes.NewReader(content.buf.Bytes())); err != nil {
				return err
			}
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

// Count gives the number of paths tracked in the buffer, meant as a proxy for metadata memory usage
func (b *Buffer) Count() int {
	return b.fileCount
}
