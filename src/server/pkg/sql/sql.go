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

// ReadRows parses the pgdump file and populates the header and the footer
// It returns EOF when done, and at that time both the Header and Footer will
// be populated. Both header and footer are required. If either are missing, an
// error is returned
func (r *PGDumpReader) ReadRows(count int64) (rowsDump []byte, rowsRead int64, err error) {
	if len(r.Header) == 0 {
		err = r.readHeader()
		if err != nil {
			return nil, 0, err
		}
	}

	endLine := "\\.\n" // Trailing '\.' denotes the end of the row inserts
	for rowsRead = 0; rowsRead < count; rowsRead++ {
		row, _err := r.rd.ReadBytes('\n')
		err = _err
		if string(row) == endLine {
			r.Footer = append(r.Footer, row...)
			err = r.readFooter()
			break
		}
		if err == io.EOF && len(r.Footer) == 0 {
			return nil, 0, fmt.Errorf("invalid pgdump - missing footer")
		}
		rowsDump = append(rowsDump, row...)
	}
	return rowsDump, rowsRead, err
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
	for true {
		b, err := r.rd.ReadBytes('\n')
		r.Footer = append(r.Footer, b...)
		if err != nil {
			return err
		}
	}
	return nil
}
