package tabwriter

import (
	"bytes"
	"io"

	"github.com/juju/ansiterm"
)

const (
	// termHeight is the default height of a terminal.
	termHeight = 50
)

// Writer is like tabwriter.Writer in the stdlibexcept that it's suitable for
// large numbers of items because it periodically flushes its contents and
// reprints a header when it does.
type Writer struct {
	w      *ansiterm.TabWriter
	lines  int
	header []byte
}

// NewWriter returns a new Writer, it will flush when
// it gets termHeight many lines, including the header line.
// The header line will be reprinted termHeight many lines have been written.
// NewStreamingWriter will panic if it's given a header that doesn't end in \n.
func NewWriter(w io.Writer, header string) *Writer {
	if header[len(header)-1] != '\n' {
		panic("header must end in a new line")
	}
	tabwriter := ansiterm.NewTabWriter(w, 0, 1, 1, ' ', 0)
	tabwriter.Write([]byte(header))
	return &Writer{
		w:      tabwriter,
		lines:  1, // 1 because we just printed the header
		header: []byte(header),
	}
}

// Write writes a line to the tabwriter.
func (w *Writer) Write(buf []byte) (int, error) {
	if w.lines >= termHeight {
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
func (w *Writer) Flush() error {
	w.lines = 0
	return w.w.Flush()
}
