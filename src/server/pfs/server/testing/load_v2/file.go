package load

import (
	"math/rand"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

type FilesSpec struct {
	Count         int             `yaml:"count,omitempty"`
	FuzzFileSpecs []*FuzzFileSpec `yaml:"fuzzFile,omitempty"`
}

func Files(spec *FilesSpec) ([]tarutil.File, error) {
	var files []tarutil.File
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

func File(spec *FileSpec) (tarutil.File, error) {
	return RandomFile(spec.RandomFileSpec)
}

type RandomFileSpec struct {
	FuzzSizeSpecs []*FuzzSizeSpec `yaml:"fuzzSize,omitempty"`
}

func RandomFile(spec *RandomFileSpec) (tarutil.File, error) {
	name := uuid.NewWithoutDashes()
	sizeSpec := FuzzSize(spec.FuzzSizeSpecs)
	min, max := sizeSpec.Min, sizeSpec.Max
	size := min
	if max > min {
		size += rand.Intn(max - min)
	}
	return tarutil.NewMemFile("/"+name, chunk.RandSeq(size)), nil
}

type SizeSpec struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}
