package starlark

import (
	"fmt"
	"reflect"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
)

type Reflect struct {
	Any any
}

var _ starlark.Value = (*Reflect)(nil)
var _ starlark.Mapping = (*Reflect)(nil)
var _ starlark.HasAttrs = (*Reflect)(nil)

func (r *Reflect) String() string {
	if r.Any == nil {
		return "<nil reflect>"
	}
	return fmt.Sprint(r.Any)
}
func (r *Reflect) Type() string {
	return reflect.ValueOf(r.Any).Type().String()
}
func (r *Reflect) Freeze() {}
func (r *Reflect) Hash() (uint32, error) {
	if r.Any == nil {
		return 0, errors.New("nil object in Reflect{}")
	}
	return uint32(xxh3.HashString(r.String()) >> 32), nil
}
func (r *Reflect) Truth() starlark.Bool { return r.Any == nil }
func (r *Reflect) Get(v starlark.Value) (starlark.Value, bool, error) {
	name, ok := starlark.AsString(v)
	if !ok {
		return nil, false, errors.Errorf("%v cannot be converted to string for key lookup", v)
	}
	v, err := r.Attr(name)
	if err != nil {
		return nil, false, err
	}
	return v, true, nil

}
func (r *Reflect) Attr(name string) (starlark.Value, error) {
	if r.Any == nil {
		return nil, errors.New("nil object in Reflect{}")
	}
	rv := reflect.ValueOf(r.Any)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		fv := rv.FieldByName(name)
		if fv.Kind() == reflect.Invalid {
			return nil, errors.Errorf("no field %q", name)
		}
		return Value(fv.Interface()), nil
	}
	return nil, errors.Errorf("cannot index into type %v (%v)", rv.Type(), rv.Kind())
}
func (r *Reflect) AttrNames() []string {
	if r.Any == nil {
		return nil
	}
	rv := reflect.ValueOf(r.Any)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		var result []string
		n := rv.NumField()
		for i := 0; i < n; i++ {
			result = append(result, rv.Type().Field(i).Name)
		}
		return result
	}
	return nil
}

// Value makes any Go value available to Starlark.
func Value(in any) starlark.Value {
	if x, ok := in.(starlark.Value); ok {
		return x
	}
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Slice:
		n := v.Len()
		var anys []any
		for i := 0; i < n; i++ {
			anys = append(anys, v.Index(i).Interface())
		}
		return ReflectList(anys)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return starlark.MakeUint64(v.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return starlark.MakeInt64(v.Int())
	case reflect.String:
		return starlark.String(v.String())
	}
	return &Reflect{Any: in}
}

// ReflectList returns a starlark.List from any Go slice.
func ReflectList[T any](xs []T) *starlark.List {
	var values []starlark.Value
	for _, x := range xs {
		values = append(values, Value(x))
	}
	return starlark.NewList(values)
}
