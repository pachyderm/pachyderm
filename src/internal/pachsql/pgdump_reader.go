package pachsql

import (
	"bufio"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// PGDumpReader parses a pgdump file into a header, rows, and a footer
type PGDumpReader struct {
	Header []byte
	Footer []byte
	rd     *bufio.Reader
}

// NewPGDumpReader creates a new PGDumpReader
func NewPGDumpReader(r *bufio.Reader) *PGDumpReader {
	return &PGDumpReader{
		rd: r,
	}
}

// ReadRow parses the pgdump file and populates the header and the footer
// It returns EOF when done, and at that time both the Header and Footer will
// be populated. Both header and footer are required. If either are missing, an
// error is returned
func (r *PGDumpReader) ReadRow() ([]byte, error) {
	if len(r.Header) == 0 {
		err := r.readHeader()
		if err != nil {
			return nil, err
		}
	}
	endLine := "\\.\n" // Trailing '\.' denotes the end of the row inserts
	row, err := r.rd.ReadBytes('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, errors.Wrapf(err, "error reading pgdump row")
	}
	// corner case: some pgdump files separate lines with \r\n (even on linux),
	// so clean this case up so all handling below is unified
	if len(row) >= 2 && row[len(row)-2] == '\r' {
		row[len(row)-2] = '\n'
		row = row[:len(row)-1]
	}
	if string(row) == endLine {
		r.Footer = append(r.Footer, row...)
		err = r.readFooter()
		row = nil // The endline is part of the footer
	}
	if errors.Is(err, io.EOF) && len(r.Footer) == 0 {
		return nil, errors.Errorf("invalid pgdump - missing footer")
	}
	return row, err
}

func (r *PGDumpReader) readHeader() error {
	done := false
	for !done {
		b, err := r.rd.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Errorf("invalid header - missing row inserts")
			}
			return errors.EnsureStack(err)
		}
		if strings.HasPrefix(string(b), "COPY") {
			done = true
		}
		r.Header = append(r.Header, b...)
	}
	return nil
}

func (r *PGDumpReader) readFooter() error {
	for {
		b, err := r.rd.ReadBytes('\n')
		r.Footer = append(r.Footer, b...)
		if err != nil {
			return errors.EnsureStack(err)
		}
	}
}
