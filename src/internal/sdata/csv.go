package sdata

import (
	"database/sql"
	"io"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/sdata/csv"
)

const NULL = ""

// CSVWriter writes Tuples in CSV format.
type CSVWriter struct {
	cw      *csv.Writer
	headers []string

	headersWritten bool
	record         []*string
}

func stringsToPointers(strings []string) (ptrs []*string) {
	for i := range strings {
		ptrs = append(ptrs, &strings[i])
	}
	return
}

// NewCSVWriter returns a CSVWriter writing to w.
// If len(headers) == 0, then no headers will be written
func NewCSVWriter(w io.Writer, headers []string) *CSVWriter {
	return &CSVWriter{
		cw:      csv.NewWriter(w),
		headers: headers,
	}
}

func (m *CSVWriter) WriteTuple(row Tuple) error {
	if m.record == nil {
		m.record = make([]*string, len(row))
	}
	if len(m.record) != len(row) {
		return ErrTupleFields{
			Writer: m,
			Tuple:  row,
		}
	}
	record := m.record
	if len(m.headers) > 0 && !m.headersWritten {
		if err := m.cw.Write(stringsToPointers(m.headers)); err != nil {
			return errors.EnsureStack(err)
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
	return errors.EnsureStack(m.cw.Write(record))
}

func (m *CSVWriter) format(x interface{}) (*string, error) {
	var y string
	switch x := x.(type) {
	case *bool:
		y = strconv.FormatBool(*x)
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
	case *float32:
		y = strconv.FormatFloat(float64(*x), 'f', -1, 32)
	case *float64:
		y = strconv.FormatFloat(*x, 'f', -1, 64)
	case *sql.RawBytes:
		// TODO: what to do here? might not be printable.
		// Maybe have a list of base64 encoded columns.
		y = string(*x)
	case *time.Time:
		y = formatTimestampNTZ(x.Format(time.RFC3339Nano))
	case *sql.NullBool:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatBool(x.Bool)
	case *sql.NullByte:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatUint(uint64(x.Byte), 10)
	case *sql.NullInt16:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatInt(int64(x.Int16), 10)
	case *sql.NullInt32:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatInt(int64(x.Int32), 10)
	case *sql.NullInt64:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatInt(x.Int64, 10)
	case *sql.NullFloat64:
		if !x.Valid {
			return nil, nil
		}
		y = strconv.FormatFloat(x.Float64, 'f', -1, 64)
	case *sql.NullString:
		if !x.Valid {
			return nil, nil
		}
		y = x.String
	case *sql.NullTime:
		if !x.Valid {
			return nil, nil
		}
		y = formatTimestampNTZ(x.Time.Format(time.RFC3339Nano))
	default:
		return nil, errors.Errorf("unrecognized value (%v: %T)", x, x)
	}
	return &y, nil
}

func (m *CSVWriter) Flush() error {
	m.cw.Flush()
	return errors.EnsureStack(m.cw.Error())
}

type CSVParser struct {
	dec *csv.Reader
}

func NewCSVParser(r io.Reader) TupleReader {
	return &CSVParser{
		dec: csv.NewReader(r),
	}
}

func (p *CSVParser) Next(row Tuple) error {
	rec, err := p.dec.Read()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if len(rec) != len(row) {
		return errors.Errorf("csv parsing: wrong number of fields HAVE: %d WANT: %d ", len(rec), len(row))
	}
	for i := range row {
		// row[i] will be a pointer to something. See Tuple comments
		if err := convert(row[i], rec[i]); err != nil {
			return err
		}
	}
	return nil
}
