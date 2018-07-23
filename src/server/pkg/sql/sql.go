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

func (r *pgDumpReader) ReadRows(count uint64) (rowsDump []byte, rowsRead uint64, err error) {
	endLine := "\\." // Trailing '\.' denotes the end of the row inserts
	if len(r.schemaHeader) == 0 {
		done := false
		for !done {
			b, err := r.rd.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					return nil, 0, fmt.Errorf("file does not contain row inserts")
				}
				return nil, 0, err
			}
			if strings.HasPrefix(string(b), "COPY") {
				done = true
			}
			r.schemaHeader = append(r.schemaHeader, b...)
		}
	}

	rowsDump = append(rowsDump, r.schemaHeader...)

	for rowsRead = 0; rowsRead < count; rowsRead++ {
		row, _err := r.rd.ReadBytes('\n')
		err = _err
		if string(row) == endLine {
			if count == 1 {
				// In this case, when we see and endline, we don't want to return any content
				return nil, 0, io.EOF
			}
			err = io.EOF
			break
		}
		rowsDump = append(rowsDump, row...)
	}
	rowsDump = append(rowsDump, []byte(endLine)...)
	return rowsDump, rowsRead, err
}
