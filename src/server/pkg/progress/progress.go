package progress

import (
	"os"

	"github.com/cheggaaa/pb"
)

func NewProxyFile(bar *pb.ProgressBar, file *os.File) *File {
	bar.Set(pb.Bytes, true)
	return &File{
		File: file,
		bar:  bar,
	}
}

// File is a wrapper around a file which updates a progress bar as it's read.
type File struct {
	*os.File
	bar *pb.ProgressBar
}

// Read reads bytes from wrapped file and adds amount of bytes to progress bar
func (f *File) Read(p []byte) (int, error) {
	n, err := f.File.Read(p)
	if err == nil {
		f.bar.Add(n)
	}
	return n, err
}

// Seek seeks the wrapped file and updates the progress bar.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	offset, err := f.File.Seek(offset, whence)
	if err == nil {
		f.bar.SetCurrent(offset)
	}
	return offset, err
}
