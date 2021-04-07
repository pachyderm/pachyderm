package load

import (
	"bytes"
	"io"
	"math/rand"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

type MemFile struct {
	path    string
	content []byte
}

func NewMemFile(path string, data []byte) *MemFile {
	return &MemFile{
		path:    path,
		content: data,
	}
}

func (mf *MemFile) Path() string {
	return mf.path
}

func (mf *MemFile) Reader() io.Reader {
	return bytes.NewReader(mf.content)
}

type FilesSpec struct {
	Count         int             `yaml:"count,omitempty"`
	FuzzFileSpecs []*FuzzFileSpec `yaml:"fuzzFile,omitempty"`
}

func Files(spec *FilesSpec) ([]*MemFile, error) {
	var files []*MemFile
	for i := 0; i < spec.Count; i++ {
		file, err := FuzzFile(spec.FuzzFileSpecs)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// TODO: Add different types of files.
type FileSpec struct {
	RandomFileSpec *RandomFileSpec `yaml:"randomFile,omitempty"`
}

func File(spec *FileSpec) (*MemFile, error) {
	return RandomFile(spec.RandomFileSpec)
}

type RandomFileSpec struct {
	FuzzSizeSpecs []*FuzzSizeSpec `yaml:"fuzzSize,omitempty"`
}

func RandomFile(spec *RandomFileSpec) (*MemFile, error) {
	name := uuid.NewWithoutDashes()
	sizeSpec := FuzzSize(spec.FuzzSizeSpecs)
	min, max := sizeSpec.Min, sizeSpec.Max
	size := min
	if max > min {
		size += rand.Intn(max - min)
	}
	return NewMemFile("/"+name, chunk.RandSeq(size)), nil
}

type SizeSpec struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}
