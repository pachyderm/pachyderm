package load

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"path"

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
	IncrementPath       bool                 `yaml:"incrementPath,omitempty"`
	RandomDirectorySpec *RandomDirectorySpec `yaml:"directory,omitempty"`
	FuzzSizeSpecs       []*FuzzSizeSpec      `yaml:"fuzzSize,omitempty"`
}

type randomFileSource struct {
	spec      *RandomFileSourceSpec
	dirSource *randomDirectorySource
	next      int64
}

func newRandomFileSource(spec *RandomFileSourceSpec) FileSource {
	var dirSource *randomDirectorySource
	if spec.RandomDirectorySpec != nil {
		dirSource = &randomDirectorySource{
			spec: spec.RandomDirectorySpec,
		}
	}
	return &randomFileSource{
		spec:      spec,
		dirSource: dirSource,
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
	var dir string
	if rfs.dirSource != nil {
		dir = rfs.dirSource.nextPath()
	}
	if rfs.spec.IncrementPath {
		next := rfs.next
		rfs.next += 1
		return path.Join(dir, fmt.Sprintf("%016d", next))
	}
	return path.Join(dir, uuid.NewWithoutDashes())
}

type RandomDirectorySpec struct {
	Depth int   `yaml:"depth,omitempty"`
	Run   int64 `yaml:"run,omitempty"`
}

type randomDirectorySource struct {
	spec *RandomDirectorySpec
	next string
	run  int64
}

func (rds *randomDirectorySource) nextPath() string {
	if rds.next == "" {
		depth := rand.Intn(rds.spec.Depth)
		for i := 0; i < depth; i++ {
			rds.next = path.Join(rds.next, uuid.NewWithoutDashes())
		}
	}
	dir := rds.next
	rds.run++
	if rds.run == rds.spec.Run {
		rds.next = ""
	}
	return dir

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
