package progress

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

// XXX DO NOT WRAP ERRORS IN THIS PACKAGE XXX
// This file replicates the go file interface, as such the errors it returns
// need to be the true errors that the os api returns, not wrapped versions of
// them. The standard library does not unwrap errors which means that things
// like io.Copy will start breaking if this returns wrapped versions of io.EOF
// instead of the real io.EOF.

var (
	containerInit sync.Once
	container     *mpb.Progress
)

func initContainer() {
	containerInit.Do(func() {
		container = mpb.New()
	})
}

func addBar(path string, size int64) *mpb.Bar {
	return container.AddBar(size,
		mpb.PrependDecorators(decor.Name(path),
			decor.Name(" "),
			decor.CountersKiloByte(" % .2f / % .2f")),
		mpb.AppendDecorators(decor.EwmaETA(decor.ET_STYLE_GO, 90),
			decor.Name(" "),
			decor.EwmaSpeed(decor.UnitKB, "% .2f", 60)))
}

// Create is identical to os.Create except that file is wrapped in a progress
// bar that updates as you write to it.
func Create(path string, size int64) (*File, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return newFile(file, path, size), nil
}

// Open is identical to os.Open except that file is wrapped in a progress bar
// that updates as you read from it .
func Open(path string) (*File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return newFile(file, path, fi.Size()), nil
}

// Stdin returns os.Stdin except that it's wrapped in a progress bar that
// updates as you read from it.
func Stdin() *File {
	return newFile(os.Stdin, "stdin", 0)
}

// File is a wrapper around a file which updates a progress bar as it's read.
type File struct {
	*os.File
	bar *mpb.Bar
	t   time.Time
}

func newFile(f *os.File, path string, size int64) *File {
	initContainer()
	return &File{
		File: f,
		bar:  addBar(path, size),
		t:    time.Now(),
	}
}

func (f *File) updateT() {
	now := time.Now()
	f.bar.DecoratorEwmaUpdate(now.Sub(f.t))
	f.t = now
}

// Read reads bytes from wrapped file and adds amount of bytes read to the
// progress bar
func (f *File) Read(p []byte) (int, error) {
	n, err := f.File.Read(p)
	if err == nil {
		f.bar.IncrBy(n)
		f.updateT()
	}
	return n, err
}

// Write writes bytes to the wrapped file and adds amount of bytes written to
// the progress bar
func (f *File) Write(p []byte) (int, error) {
	n, err := f.File.Write(p)
	if err == nil {
		f.bar.IncrBy(n)
		f.updateT()
	}
	return n, err
}

// WriteAt writes to the wrapped file at the given offset and adds amount of
// bytes written to the progress bar
func (f *File) WriteAt(b []byte, offset int64) (int, error) {
	n, err := f.File.WriteAt(b, offset)
	if err == nil {
		f.bar.SetCurrent(offset + int64(n))
		f.updateT()
	}
	return n, err
}

// ReadFrom writes the contents of r to f and adds the amount of bytes written
// to the progress bar
func (f *File) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.File.ReadFrom(r)
	if err == nil {
		f.bar.IncrBy(int(n))
		f.updateT()
	}
	return n, err
}

// Seek seeks the wrapped file and updates the progress bar.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	offset, err := f.File.Seek(offset, whence)
	if err == nil {
		f.bar.SetCurrent(offset)
		f.updateT()
	}
	return offset, err
}

// Close closes the wrapped file and finishes the progress bar.
func (f *File) Close() error {
	f.Finish()
	return f.File.Close()
}

// Finish finishes the progress bar without closing the wrapped file, this
// should be used if the wrapped file is something you don't want to close (for
// example stdin), but you don't want future reads to be printed as progress.
func (f *File) Finish() {
	f.bar.SetTotal(0, true)
}
