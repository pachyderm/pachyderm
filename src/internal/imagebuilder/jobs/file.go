package jobs

import (
	"fmt"
)

// File represents something on disk (including directories).
type File struct {
	Name   string
	Path   string
	Digest Digest
}

func (f *File) String() string {
	if f == nil {
		return "<nil file>"
	}
	return fmt.Sprintf("file:%v@%v", f.Path, f.Digest.String())
}
