package progress

import (
	"os"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

var (
	// Template is used when you know the total size of the operation (i.e.
	// when you're reading from a file)
	Template pb.ProgressBarTemplate = `{{string . "prefix"}}: {{counters . }} {{bar . "[" "=" ">" " " "]"}} {{percent . }} {{speed . }} {{rtime . "ETA %s"}}{{string . "suffix"}}`
	// PipeTemplate is used when you don't know the total size of the operation
	// (i.e. when you're reading from stdin.
	PipeTemplate pb.ProgressBarTemplate = `{{string . "prefix"}}: {{counters . }} {{cycle . "[    ]" "[>   ]" "[=>  ]" "[==> ]" "[ ==>]" "[  ==]" "[   =]" "[    ]" "[   <]" "[  <=]" "[ <==]" "[<===]" "[=== ]" "[==  ]" "[=   ]"}} {{speed . }} {{string . "suffix"}}`

	// mu makes sure that only one progress bar is running at a time this is
	// necessary because multiple bars running at the same time leads to weird
	// terminal output.
	mu sync.Mutex
)

func start(prefix string, bar *pb.ProgressBar) {
	bar.Set("prefix", prefix)
	bar.Set(pb.Bytes, true)
	bar.Start()
}

// Open is identical to os.Open except that file is wrapped in a progress bar
// that updates as you read from it .
func Open(path string) (*File, error) {
	mu.Lock()
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}
	bar := Template.New(int(fi.Size()))
	start(path, bar)
	return &File{
		File: file,
		bar:  bar,
	}, nil
}

// Stdin returns os.Stdin except that it's wrapped in a progress bar that
// updates as you read from it.
func Stdin() *File {
	mu.Lock()
	bar := PipeTemplate.New(0)
	start("stdin", bar)
	return &File{
		File: os.Stdin,
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

// Close closes the wrapped file and finishes the progress bar.
func (f *File) Close() error {
	f.Finish()
	return f.File.Close()
}

// Finish finishes the progress bar without closing the wrapped file, this
// should be used if the wrapped file is something you don't want to close (for
// example stdin), but you don't want future reads to be printed as progress.
func (f *File) Finish() {
	f.bar.Finish()
	mu.Unlock()
}
