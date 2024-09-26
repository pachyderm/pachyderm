package sdata

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// JSONWriter writes tuples as newline separated json objects.
type JSONWriter struct {
	bufw   *bufio.Writer
	enc    *json.Encoder
	fields []string
	record map[string]interface{}
}

// NewJSONWriter returns a new JSONWriter which will write the indicated fields to an io.Writer.
//
// TODO: figure out some way to specify a projection so that we can write nested
// structures.
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

// WriteTuple writes a Tuple to a JSONWriter.
func (m *JSONWriter) WriteTuple(row Tuple) error {
	if len(row) != len(m.fields) {
		return ErrTupleFields{Writer: m, Fields: m.fields, Tuple: row}
	}
	record := m.record
	for i := range row {
		var y interface{}
		switch x := row[i].(type) {
		case *sql.NullBool:
			if x.Valid {
				y = x.Bool
			} else {
				y = nil
			}
		case *sql.NullByte:
			if x.Valid {
				y = x.Byte
			} else {
				y = nil
			}
		case *sql.NullInt16:
			if x.Valid {
				y = x.Int16
			} else {
				y = nil
			}
		case *sql.NullInt32:
			if x.Valid {
				y = x.Int32
			} else {
				y = nil
			}
		case *sql.NullInt64:
			if x.Valid {
				y = x.Int64
			} else {
				y = nil
			}
		case *sql.NullFloat64:
			if x.Valid {
				y = x.Float64
			} else {
				y = nil
			}
		case *sql.NullString:
			if x.Valid {
				y = x.String
			} else {
				y = nil
			}
		case *sql.NullTime:
			if x.Valid {
				y = formatTimestampNTZ(x.Time.Format(time.RFC3339Nano))
			} else {
				y = nil
			}
		case *time.Time:
			y = formatTimestampNTZ(x.Format(time.RFC3339Nano))
		case *sql.RawBytes:
			y = string(*x)
		default:
			y = row[i]
		}
		record[m.fields[i]] = y
	}
	return errors.EnsureStack(m.enc.Encode(record))
}

// Flush flushes the underlying io.Writer.
func (m *JSONWriter) Flush() error {
	return errors.EnsureStack(m.bufw.Flush())
}

// A JSONParser parses JSON.
type JSONParser struct {
	dec        *json.Decoder
	fieldNames []string

	m map[string]interface{}
}

// NewJSONParser returns a TupleReader which will parse tuples with the
// indicated field names from an underlying io.Reader.
func NewJSONParser(r io.Reader, fieldNames []string) TupleReader {
	dec := json.NewDecoder(r)
	// UseNumber() is necessary to correctly parse large int64s, we have to first parse them
	// as json.Numbers and then handle those in convert.  Otherwise they are parsed as float64s
	// and precision is lost for values above ~2^53.
	dec.UseNumber()
	return &JSONParser{
		dec:        dec,
		fieldNames: fieldNames,
	}
}

// Next returns the next tuple from the underlying io.Reader.
func (p *JSONParser) Next(row Tuple) error {
	if len(row) != len(p.fieldNames) {
		return ErrTupleFields{Fields: p.fieldNames, Tuple: row}
	}
	m := p.getMap()
	if err := p.dec.Decode(&m); err != nil {
		return errors.EnsureStack(err)
	}
	for i := range row {
		colName := p.fieldNames[i]
		v, exists := m[colName]
		if !exists {
			row[i] = nil
			continue
		}
		if err := convert(row[i], v); err != nil {
			return err
		}
	}
	return nil
}

func (p *JSONParser) getMap() map[string]interface{} {
	if p.m == nil {
		p.m = make(map[string]interface{})
	}
	for k := range p.m {
		delete(p.m, k)
	}
	return p.m
}
