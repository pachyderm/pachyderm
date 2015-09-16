package protolog

import (
	"bytes"
	"sync"
)

var (
	newlineBytes = []byte{'\n'}
)

type writePusher struct {
	writeFlusher WriteFlusher
	marshaller   Marshaller
	newline      bool
	lock         *sync.Mutex
}

func newWritePusher(writeFlusher WriteFlusher, options WritePusherOptions) *writePusher {
	writePusher := &writePusher{
		writeFlusher,
		options.Marshaller,
		options.Newline,
		&sync.Mutex{},
	}
	if writePusher.marshaller == nil {
		writePusher.marshaller = defaultMarshaller
	}
	return writePusher
}

func (w *writePusher) Flush() error {
	return w.writeFlusher.Flush()
}

func (w *writePusher) Push(entry *Entry) error {
	data, err := w.marshaller.Marshal(entry)
	if err != nil {
		return err
	}
	if w.newline {
		buffer := bytes.NewBuffer(data)
		_, _ = buffer.Write(newlineBytes)
		data = buffer.Bytes()
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, err := w.writeFlusher.Write(data); err != nil {
		return err
	}
	return nil
}
