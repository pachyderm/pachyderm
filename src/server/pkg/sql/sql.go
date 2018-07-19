package sql

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type pgDumpReader struct {
	schemaHeader []byte
	rd           *bufio.Reader
}

func NewPGDumpReader(r *bufio.Reader) *pgDumpReader {
	return &pgDumpReader{
		schemaHeader: make([]byte, 0),
		rd:           r,
	}
}

func (r *pgDumpReader) ReadRows(count uint64) (rowsDump []byte, err error) {
	endLine := "\\."
	// use r.ReadBytes(delim) to read up until insert
	// store the header, then read each row until there are no more rows left
	// The first read needs to populate the header
	if len(r.schemaHeader) == 0 {
		done := false
		for !done {
			b, err := r.rd.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("file does not contain row inserts")
				}
				return nil, err
			}
			if strings.HasPrefix(string(b), "COPY") {
				done = true
			}
			r.schemaHeader = append(r.schemaHeader, b...)
		}
	}

	rowsDump = append(rowsDump, r.schemaHeader...)

	var i uint64
	for i = 0; i < count; i++ {
		// From here on .. we want to read each line ... and end when we see the trailing '\.'
		row, err := r.rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		// Trailing '\.' denotes the end of the row inserts
		if string(row) == endLine {
			return nil, io.EOF
		}
		rowsDump = append(rowsDump, row...)
	}
	rowsDump = append(rowsDump, []byte(endLine)...)
	return rowsDump, nil
}
