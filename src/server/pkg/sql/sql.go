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
	fmt.Printf("reading %v rows\n", count)
	// Trailing '\.' denotes the end of the row inserts
	endLine := "\\."
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

	// when count > 1 ... and I see the endline ... I want to return the
	// cumulative stuff I've read ... but I need to append the final line
	// when count == 1 ... I want to return immediately ... because it means the
	// row I just read is fluff ... and I don't want to return a header ... or
	// anything

	if count == 1 {
		row, err := r.rd.ReadBytes('\n')
		if string(row) == endLine {
			fmt.Printf("read endline %v\n", string(row))
			fmt.Printf("rowdump %v\n", string(rowsDump))
			return nil, 0, io.EOF
		}
		rowsDump = append(rowsDump, row...)
		rowsDump = append(rowsDump, []byte(endLine)...)
		return rowsDump, rowsRead, err
	}

	for rowsRead = 0; rowsRead < count; rowsRead++ {
		fmt.Printf("reading row %v of %v\n", rowsRead, count)
		row, _err := r.rd.ReadBytes('\n')
		err = _err
		if string(row) == endLine {
			fmt.Printf("read endline %v\n", string(row))
			fmt.Printf("rowdump %v\n", string(rowsDump))
			err = io.EOF
			break
		}
		rowsDump = append(rowsDump, row...)
	}
	rowsDump = append(rowsDump, []byte(endLine)...)
	return rowsDump, rowsRead, err
}
