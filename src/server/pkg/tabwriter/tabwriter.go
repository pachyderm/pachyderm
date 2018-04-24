package tabwriter

import (
	"bytes"
	"text/tabwriter"
)

// StreamingTabWriter is like tabwriter except that it's suitable for large
// numbers of items because it periodically flushes its contents.
type StreamingTabWriter struct {
	w      *tabwriter.Writer
	lines  int
	height int
	header []byte
}

// NewStreamingWriter returns a new streaming tabwriter, it will flush when
// it gets height many lines, including the header line.
// The header line will be reprinted everytime height many lines have been written.
// NewStreamingTabWriter will panice if it's given a height < 2
func NewStreamingWriter(w *tabwriter.Writer, height int, header string) *StreamingTabWriter {
	if height < 2 {
		panic("cannot create a StreamingTabWriter with height less than 2")
	}
	if header[len(header)-1] != '\n' {
		panic("header must end in a new line")
	}
	w.Write([]byte(header))
	return &StreamingTabWriter{
		w:      w,
		lines:  1, // 1 because we just printed the header
		height: height,
		header: []byte(header),
	}
}

// Write writes a line to the tabwriter.
func (w *StreamingTabWriter) Write(buf []byte) (int, error) {
	if w.lines >= w.height {
		if err := w.Flush(); err != nil {
			return 0, err
		}
		if _, err := w.w.Write(w.header); err != nil {
			return 0, err
		}
		w.lines++
	}
	w.lines += bytes.Count(buf, []byte{'\n'})
	return w.w.Write(buf)
}

// Flush flushes the underlying tab writer.
func (w *StreamingTabWriter) Flush() error {
	w.lines = 0
	return w.w.Flush()
}
