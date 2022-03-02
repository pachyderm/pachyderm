package sdata

import (
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type TupleReader interface {
	// Next attempts to read one Tuple into x.
	// If the next data is the wrong shape for x then an error is returned.
	Next(x Tuple) error
}

type jsonParser struct {
	dec        *json.Decoder
	fieldNames []string

	m map[string]interface{}
}

func NewJSONParser(r io.Reader, fieldNames []string) TupleReader {
	return &jsonParser{
		dec:        json.NewDecoder(r),
		fieldNames: fieldNames,
	}
}

func (p *jsonParser) Next(row Tuple) error {
	if len(row) != len(p.fieldNames) {
		return ErrTupleFields{Fields: p.fieldNames, Tuple: row}
	}
	m := p.getMap()
	if err := p.dec.Decode(m); err != nil {
		return err
	}
	for i := range row {
		colName := p.fieldNames[i]
		v, exists := m[colName]
		if !exists {
			row[i] = nil
			continue
		}
		ty1 := reflect.TypeOf(v)
		ty2 := reflect.TypeOf(row[i])
		if ty2.AssignableTo(ty1) {
			row[i] = v
		} else if !ty1.ConvertibleTo(ty2) {
			return errors.Errorf("cannot convert %v to %v", ty1, ty2)
		} else {
			row[i] = reflect.ValueOf(v).Interface()
		}
	}
	return nil
}

func (p *jsonParser) getMap() map[string]interface{} {
	if p.m == nil {
		p.m = make(map[string]interface{})
	}
	return p.m
}

type csvParser struct {
	dec *csv.Reader
}

func NewCSVParser(r io.Reader) TupleReader {
	return &csvParser{
		dec: csv.NewReader(r),
	}
}

func (p *csvParser) Next(row Tuple) error {
	rec, err := p.dec.Read()
	if err != nil {
		return err
	}
	if len(rec) != len(row) {
		return errors.Errorf("csv parsing: wrong number of fields HAVE: %d WANT: %d ", len(rec), len(row))
	}
	for i := range row {
		// reflect.ValueOf(slice[i]) will always be a pointer to that value in the slice
		// it will have Kind() == reflect.Ptr
		v := reflect.ValueOf(row[i])
		if err := convertString(v.Interface(), rec[i]); err != nil {
			return err
		}
	}
	return nil
}

// convertString
func convertString(dest interface{}, x string) error {
	dty := reflect.TypeOf(dest)
	if dty.Kind() != reflect.Ptr {
		panic("dest must be pointer")
	}
	isNull := x == "null" || x == "nil" || len(x) == 0
	switch d := dest.(type) {
	case *string:
		*d = x
	case *int:
		n, err := strconv.Atoi(x)
		if err != nil {
			return err
		}
		*d = n
	case *float64:
		fl, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return err
		}
		*d = fl
	case *[]byte:
		codec := base64.StdEncoding
		data, err := codec.DecodeString(x)
		if err != nil {
			return err
		}
		*d = append((*d)[:0], data...)
	case *time.Time:
		t, err := parseTime(x)
		if err != nil {
			return err
		}
		*d = t
	case *sql.NullBool:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseBool(x)
			if err != nil {
				return err
			}
			d.Bool = x2
			d.Valid = true
		}
	case *sql.NullByte:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseUint(x, 10, 8)
			if err != nil {
				return err
			}
			d.Byte = byte(x2)
			d.Valid = true
		}
	case *sql.NullInt16:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseInt(x, 10, 16)
			if err != nil {
				return err
			}
			d.Int16 = int16(x2)
			d.Valid = true
		}
	case *sql.NullInt32:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseInt(x, 10, 32)
			if err != nil {
				return err
			}
			d.Int32 = int32(x2)
			d.Valid = true
		}
	case *sql.NullInt64:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseInt(x, 10, 32)
			if err != nil {
				return err
			}
			d.Int64 = x2
			d.Valid = true
		}
	case *sql.NullFloat64:
		if isNull {
			d.Valid = false
		} else {
			x2, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return err
			}
			d.Float64 = x2
			d.Valid = true
		}
	case *sql.NullTime:
		if isNull {
			d.Valid = false
		} else {
			t, err := parseTime(x)
			if err != nil {
				return err
			}
			d.Time = t
			d.Valid = true
		}
	case *sql.NullString:
		if isNull {
			d.Valid = false
		} else {
			d.String = x
			d.Valid = true
		}
	default:
		return errors.Errorf("cannot convert string to %T", dest)
	}
	return nil
}

// parseTime attempts to parse the time using every format and returns the first one.
func parseTime(x string) (t time.Time, err error) {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC1123Z,
		time.RFC822Z,
		time.Kitchen,
		time.ANSIC,
	} {
		t, err := time.Parse(layout, x)
		if err == nil {
			return t, err
		}
	}
	return t, err
}
