package load

import (
	"bytes"
	"fmt"
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

type FileSourceSpec struct {
	Name                 string                `yaml:"name,omitempty"`
	RandomFileSourceSpec *RandomFileSourceSpec `yaml:"random,omitempty"`
}

type FileSource interface {
	Next() *MemFile
}

func NewFileSource(spec *FileSourceSpec) FileSource {
	return newRandomFileSource(spec.RandomFileSourceSpec)
}

type RandomFileSourceSpec struct {
	IncrementPath bool            `yaml:"incrementPath,omitempty"`
	FuzzSizeSpecs []*FuzzSizeSpec `yaml:"fuzzSize,omitempty"`
}

type randomFileSource struct {
	spec *RandomFileSourceSpec
	next int64
}

func newRandomFileSource(spec *RandomFileSourceSpec) FileSource {
	return &randomFileSource{
		spec: spec,
	}
}

func (rfs *randomFileSource) Next() *MemFile {
	sizeSpec := FuzzSize(rfs.spec.FuzzSizeSpecs)
	min, max := sizeSpec.Min, sizeSpec.Max
	size := min
	if max > min {
		size += rand.Intn(max - min)
	}
	return NewMemFile(rfs.nextPath(), chunk.RandSeq(size))
}

func (rfs *randomFileSource) nextPath() string {
	if rfs.spec.IncrementPath {
		next := rfs.next
		rfs.next += 1
		return fmt.Sprintf("%016d", next)
	}
	return uuid.NewWithoutDashes()
}

type FilesSpec struct {
	Count         int             `yaml:"count,omitempty"`
	FuzzFileSpecs []*FuzzFileSpec `yaml:"fuzzFile,omitempty"`
}

func Files(env *Env, spec *FilesSpec) ([]*MemFile, error) {
	var files []*MemFile
	for i := 0; i < spec.Count; i++ {
		file, err := FuzzFile(env, spec.FuzzFileSpecs)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

// TODO: Add different types of files.
type FileSpec struct {
	Source string `yaml:"source,omitempty"`
}

func File(env *Env, spec *FileSpec) (*MemFile, error) {
	return env.FileSource(spec.Source).Next(), nil
}

type SizeSpec struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}
