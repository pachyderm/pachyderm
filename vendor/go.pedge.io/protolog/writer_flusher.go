package protolog

import "io"

type writerFlusher struct {
	io.Writer
}

func newWriterFlusher(writer io.Writer) *writerFlusher {
	return &writerFlusher{writer}
}

func (w *writerFlusher) Flush() error {
	return nil
}
