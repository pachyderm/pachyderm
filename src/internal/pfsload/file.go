package pfsload

import (
	"fmt"
	"io"
	"math/rand"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/randutil"
)

const pathSize = 32

type RandomFile struct {
	path string
	r    io.Reader
}

func NewRandomFile(path string, r io.Reader) *RandomFile {
	return &RandomFile{
		path: path,
		r:    r,
	}
}

func (f *RandomFile) Path() string {
	return f.path
}

func (f *RandomFile) Read(data []byte) (int, error) {
	res, err := f.r.Read(data)
	return res, errors.EnsureStack(err)
}

type FileSource interface {
	Next() (*RandomFile, error)
}

func NewFileSource(spec *FileSourceSpec, random *rand.Rand) FileSource {
	return newRandomFileSource(spec.Random, random)
}

type randomFileSource struct {
	spec      *RandomFileSourceSpec
	random    *rand.Rand
	dirSource *randomDirectorySource
	next      int64
}

func newRandomFileSource(spec *RandomFileSourceSpec, random *rand.Rand) FileSource {
	var dirSource *randomDirectorySource
	if spec.Directory != nil {
		dirSource = &randomDirectorySource{
			spec:   spec.Directory,
			random: random,
		}
	}
	return &randomFileSource{
		spec:      spec,
		random:    random,
		dirSource: dirSource,
	}
}

func (rfs *randomFileSource) Next() (*RandomFile, error) {
	sizeSpec, err := FuzzSize(rfs.spec.Sizes, rfs.random)
	if err != nil {
		return nil, err
	}
	min, max := sizeSpec.MinSize, sizeSpec.MaxSize
	size := min
	if max > min {
		size += rfs.random.Int63n(max - min)
	}
	return NewRandomFile(rfs.nextPath(), randutil.NewBytesReader(rfs.random, int64(size))), nil
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
	return path.Join(dir, string(randutil.Bytes(rfs.random, pathSize)))
}

type randomDirectorySource struct {
	spec   *RandomDirectorySpec
	random *rand.Rand
	next   string
	run    int64
}

func (rds *randomDirectorySource) nextPath() string {
	if rds.next == "" {
		min, max := rds.spec.Depth.MinSize, rds.spec.Depth.MaxSize
		depth := int(min)
		if max > min {
			depth += rds.random.Intn(int(max - min))
		}
		for i := 0; i < depth; i++ {
			rds.next = path.Join(rds.next, string(randutil.Bytes(rds.random, pathSize)))
		}
	}
	dir := rds.next
	rds.run++
	if rds.run == rds.spec.Run {
		rds.next = ""
		rds.run = 0
	}
	return dir

}

func Files(fileSource FileSource, count int) ([]*RandomFile, error) {
	var files []*RandomFile
	for i := 0; i < count; i++ {
		file, err := fileSource.Next()
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		files = append(files, file)
	}
	return files, nil
}
