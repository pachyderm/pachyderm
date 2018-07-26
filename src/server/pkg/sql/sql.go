package sql

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type pgDumpReader struct {
	Header []byte
	Footer []byte
	rd     *bufio.Reader
}

func NewPGDumpReader(r *bufio.Reader) *pgDumpReader {
	return &pgDumpReader{
		//		Header: make([]byte, 0),
		//	Footer: make([]byte, 0),
		rd: r,
	}
}

// ReadRows parses the pgdump file and populates the header and the footer
// It returns EOF when done, and at that time both the Header and Footer will
// be populated. Both header and footer are required. If either are missing, an
// error is returned
func (r *pgDumpReader) ReadRows(count int64) (rowsDump []byte, rowsRead int64, err error) {
	endLine := "\\.\n" // Trailing '\.' denotes the end of the row inserts
	if len(r.Header) == 0 {
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
			r.Header = append(r.Header, b...)
		}
		fmt.Printf("read off the header (%v) w err(%v)\n", string(r.Header), err)
	}

	//	rowsDump = append(rowsDump, r.Header...)

	for rowsRead = 0; rowsRead < count; rowsRead++ {
		row, _err := r.rd.ReadBytes('\n')
		err = _err
		if string(row) == endLine {
			fmt.Printf("this row (%v) is an endline\n", string(row))
			r.Footer = append(r.Footer, row...)
			err = r.readFooter() // We will return any rows we did read + the error
			fmt.Printf("read off the footer (%v) w err(%v)\n", string(r.Footer), err)
			if count == 1 {
				// In this case, when we see and endline, we don't want to return any content
				return nil, 0, io.EOF
			}
			break
		}
		rowsDump = append(rowsDump, row...)
	}
	//	rowsDump = append(rowsDump, []byte(endLine)...)
	return rowsDump, rowsRead, err
}

func (r *pgDumpReader) readFooter() error {
	for true {
		b, err := r.rd.ReadBytes('\n')
		r.Footer = append(r.Footer, b...)
		if err != nil {
			return err
		}
	}
	return nil
}
