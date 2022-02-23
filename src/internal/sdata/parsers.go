package sdata

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"io"
	"reflect"
	"strconv"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

type TupleReader interface {
	Next(Tuple) error
}

type jsonParser struct {
	dec      *json.Decoder
	colNames []string

	m map[string]interface{}
}

func NewJSONParser(r io.Reader, colNames []string) TupleReader {
	return &jsonParser{
		dec:      json.NewDecoder(r),
		colNames: colNames,
	}
}

func (p *jsonParser) Next(row Tuple) error {
	if len(row) != len(p.colNames) {
		return ErrTupleFields{Fields: p.colNames, Tuple: row}
	}
	m := p.getMap()
	if err := p.dec.Decode(m); err != nil {
		return err
	}
	for i := range row {
		colName := p.colNames[i]
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

func NewCSVParser(r io.Reader, colNames []string) TupleReader {
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
		if err := convertString(row[i], rec[i]); err != nil {
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
	// TODO: check it's a pointer type
	switch d := dest.(type) {
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
	default:
		return errors.Errorf("unrecognized type %T", dest)
	}
	return nil
}
