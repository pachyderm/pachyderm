package lion

import (
	"io"
	"os"
	"sync"

	"github.com/mattn/go-isatty"
)

var (
	newlineBytes = []byte{'\n'}
)

type syncer interface {
	Sync() error
}

type writePusher struct {
	writer     io.Writer
	marshaller Marshaller
	lock       *sync.Mutex
}

func newWritePusher(writer io.Writer, marshaller Marshaller) *writePusher {
	writePusher := &writePusher{
		writer,
		marshaller,
		&sync.Mutex{},
	}
	if file, ok := writer.(*os.File); ok {
		if textMarshaller, ok := writePusher.marshaller.(TextMarshaller); ok {
			if isatty.IsTerminal(file.Fd()) {
				writePusher.marshaller = textMarshaller.WithColors()
			}
		}
	}
	return writePusher
}

func (w *writePusher) Flush() error {
	if syncer, ok := w.writer.(syncer); ok {
		return syncer.Sync()
	} else if flusher, ok := w.writer.(Flusher); ok {
		return flusher.Flush()
	}
	return nil
}

func (w *writePusher) Push(entry *Entry) error {
	data, err := w.marshaller.Marshal(entry)
	if err != nil {
		return err
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, err := w.writer.Write(data); err != nil {
		return err
	}
	return nil
}
