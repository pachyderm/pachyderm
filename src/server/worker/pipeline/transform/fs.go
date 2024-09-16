package transform

import (
	"fmt"
	"io/fs"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/pjs"
)

var (
	workerFS   fstest.MapFS
	workerHash []byte
)

func init() {
	workerFS = fstest.MapFS{
		"name": &fstest.MapFile{Data: []byte("preprocessing")},
	}
	var err error
	if workerHash, err = pjs.HashFS(workerFS); err != nil {
		panic(fmt.Sprintf("could not hash workerFS: %v", err))
	}
}

// ProgramFS returns a filesystem representing the PPS worker program.
func ProgramFS() fs.FS {
	var w = make(fstest.MapFS, len(workerFS))
	for k, v := range workerFS {
		var vv = new(fstest.MapFile)
		*vv = *v
		w[k] = vv
	}
	return w
}

// ProgramHash returns a hash representing the PPS worker program.
func ProgramHash() []byte {
	return workerHash[:]
}
