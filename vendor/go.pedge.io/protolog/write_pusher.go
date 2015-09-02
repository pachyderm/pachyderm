package protolog

import "sync"

var (
	newlineBytes = []byte{'\n'}
)

type writePusher struct {
	writeFlusher WriteFlusher
	marshaller   Marshaller
	encoder      Encoder
	lock         *sync.Mutex
}

func newWritePusher(writeFlusher WriteFlusher, options WritePusherOptions) *writePusher {
	writePusher := &writePusher{
		writeFlusher,
		options.Marshaller,
		options.Encoder,
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
	if w.encoder != nil {
		data, err = w.encoder.Encode(data)
		if err != nil {
			return err
		}
	}
	w.lock.Lock()
	defer w.lock.Unlock()
	if _, err := w.writeFlusher.Write(data); err != nil {
		return err
	}
	if w.encoder == nil {
		_, err := w.writeFlusher.Write(newlineBytes)
		return err
	}
	return nil
}
