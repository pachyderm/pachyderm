package sql

import (
	"bufio"
	"fmt"
	"io"
	"strings"
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
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading pgdump row: %v", err)
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
	if err == io.EOF && len(r.Footer) == 0 {
		return nil, fmt.Errorf("invalid pgdump - missing footer")
	}
	return row, err
}

func (r *PGDumpReader) readHeader() error {
	done := false
	for !done {
		b, err := r.rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("invalid header - missing row inserts")
			}
			return err
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
			return err
		}
	}
	return nil
}
