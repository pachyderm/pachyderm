// package sdata deals with structured data.
package sdata

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// ErrTupleFields is returned by writers to indicate that
// a tuple is not the right shape to be written to them.
type ErrTupleFields struct {
	Writer TupleWriter
	Tuple  Tuple
	Fields []string
}

func (e ErrTupleFields) Error() string {
	return fmt.Sprintf("tuple has invalid fields for this writer (%T). expected: (%v), only have %d", e.Writer, e.Fields, (e.Tuple))
}

// TupleWriter is the type of Writers for structured data.
type TupleWriter interface {
	WriteTuple(row Tuple) error
	Flush() error
}

type Tuple = []interface{}

// CSVWriter writes Tuples in CSV format.
type CSVWriter struct {
	cw      *csv.Writer
	headers []string

	headersWritten bool
	record         []string
}

// NewCSVWriter returns a CSVWriter writing to w.
// If len(headers) == 0, then no headers will be written
func NewCSVWriter(w io.Writer, headers []string) *CSVWriter {
	return &CSVWriter{cw: csv.NewWriter(w)}
}

func (m *CSVWriter) WriteTuple(row Tuple) error {
	if len(m.record) != len(row) {
		m.record = make([]string, len(row))
	}
	record := m.record
	if len(m.headers) > 0 && !m.headersWritten {
		if err := m.cw.Write(m.headers); err != nil {
			return err
		}
		m.headersWritten = true
	}
	for i := range row {
		var err error
		record[i], err = m.format(row[i])
		if err != nil {
			return err
		}
	}
	return m.cw.Write(record)
}

func (m *CSVWriter) format(x interface{}) (string, error) {
	var y string
	switch x := x.(type) {
	case *int16:
		y = strconv.FormatInt(int64(*x), 10)
	case *int32:
		y = strconv.FormatInt(int64(*x), 10)
	case *int64:
		y = strconv.FormatInt(*x, 10)
	case *uint64:
		y = strconv.FormatUint(*x, 10)
	case *string:
		y = *x
	case *float64:
		y = strconv.FormatFloat(*x, 'f', -1, 64)
	case *sql.RawBytes:
		// TODO: what to do here? might not be printable.
		// Maybe have a list of base64 encoded columns.
		y = string(*x)
	case *sql.NullInt64:
		if x.Valid {
			y = strconv.FormatInt(x.Int64, 10)
		} else {
			y = "null"
		}
	default:
		return "", errors.Errorf("unrecognized value (%v: %T)", x, x)
	}
	return y, nil
}

func (m *CSVWriter) Flush() error {
	m.cw.Flush()
	return m.cw.Error()
}

// JSONWriter writes tuples as newline separated json objects.
type JSONWriter struct {
	bufw   *bufio.Writer
	enc    *json.Encoder
	fields []string
	record map[string]interface{}
}

func NewJSONWriter(w io.Writer, fieldNames []string) *JSONWriter {
	bufw := bufio.NewWriter(w)
	enc := json.NewEncoder(bufw)
	return &JSONWriter{
		bufw:   bufw,
		enc:    enc,
		fields: fieldNames,
		record: make(map[string]interface{}, len(fieldNames)),
	}
}

func (m *JSONWriter) WriteTuple(row Tuple) error {
	if len(row) != len(m.fields) {
		return ErrTupleFields{Writer: m, Fields: m.fields, Tuple: row}
	}
	record := m.record
	for i := range row {
		var y interface{}
		switch x := row[i].(type) {
		case *sql.NullInt64:
			if x.Valid {
				y = x.Int64
			} else {
				y = nil
			}
		default:
			y = row[i]
		}
		record[m.fields[i]] = y
	}
	return m.enc.Encode(record)
}

func (m *JSONWriter) Flush() error {
	return m.bufw.Flush()
}
