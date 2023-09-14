package starlark

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/zeebo/xxh3"
	"go.starlark.net/starlark"
)

// Reflect is a starlark.Value that exposes the fields of struct types.  For example, `package foo;
// struct Test { Foo string }` would be of type "foo.Test" and have one attribute, "Foo", in
// starlark, used like value.Foo.
type Reflect struct {
	Any any
}

var _ starlark.Value = (*Reflect)(nil)
var _ starlark.Mapping = (*Reflect)(nil)
var _ starlark.HasAttrs = (*Reflect)(nil)
var _ starlark.TotallyOrdered = (*Reflect)(nil)

func (r *Reflect) String() string {
	if r.Any == nil {
		return "<nil reflect>"
	}
	if s, ok := r.Any.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%#v", r.Any)
}
func (r *Reflect) Type() string {
	return reflect.TypeOf(r.Any).String()
}
func (r *Reflect) Freeze() {}
func (r *Reflect) Hash() (uint32, error) {
	if r.Any == nil {
		return 0, errors.New("nil object in Reflect{}")
	}
	return uint32(xxh3.HashString(r.String())), nil
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
func (r *Reflect) Cmp(y starlark.Value, _ int) (int, error) {
	sr := FromStarlark(r)
	sy := FromStarlark(y)
	if reflect.DeepEqual(sr, sy) {
		return 0, nil
	}
	return 0, errors.Errorf("cannot compare a %T (%v) to a %T (%v)", r, r, sy, sy)
}

// Value makes any Go value available to Starlark.
func Value(in any) starlark.Value {
	switch x := in.(type) {
	case starlark.Value:
		return x
	case []starlark.Value:
		return ReflectList(x)
	case []byte:
		return starlark.Bytes(string(x))
	case string:
		return starlark.String(x)
	case bool:
		return starlark.Bool(x)
	case int:
		return starlark.MakeInt(x)
	case int8:
		return starlark.MakeInt64(int64(x))
	case int16:
		return starlark.MakeInt64(int64(x))
	case int32:
		return starlark.MakeInt64(int64(x))
	case int64:
		return starlark.MakeInt64(x)
	case uint:
		return starlark.MakeUint(x)
	case uint8:
		return starlark.MakeUint64(uint64(x))
	case uint16:
		return starlark.MakeUint64(uint64(x))
	case uint32:
		return starlark.MakeUint64(uint64(x))
	case uint64:
		return starlark.MakeUint64(x)
	case *big.Int:
		return starlark.MakeBigInt(x)
	case float32:
		return starlark.Float(float64(x))
	case float64:
		return starlark.Float(x)
	}
	return value(reflect.ValueOf(in))
}

func value(v reflect.Value) starlark.Value {
	if v.Kind() == reflect.Invalid {
		return starlark.None
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		n := v.Len()
		var anys []any
		for i := 0; i < n; i++ {
			anys = append(anys, v.Index(i).Interface())
		}
		return ReflectList(anys)
	case reflect.Map:
		d := starlark.NewDict(v.Len())
		iter := v.MapRange()
		for iter.Next() {
			k := Value(iter.Key().Interface())
			v := Value(iter.Value().Interface())
			d.SetKey(k, v)
		}
		return d
	}
	return &Reflect{Any: v.Interface()}
}

// ReflectList returns a starlark.List from any Go slice.
func ReflectList[T any](xs []T) *starlark.List {
	var values []starlark.Value
	for _, x := range xs {
		values = append(values, Value(x))
	}
	return starlark.NewList(values)
}

// FromStarlark returns a Go value for a Starlark value.
func FromStarlark(v starlark.Value) any {
	switch x := v.(type) {
	case *Reflect:
		return x.Any
	case starlark.NoneType:
		return nil
	case starlark.Bool:
		return bool(x)
	case starlark.Bytes:
		return []byte(x)
	case starlark.String:
		return string(x)
	case starlark.Int:
		i := x.BigInt()
		if i.IsInt64() {
			return i.Int64()
		} else if i.IsUint64() {
			return i.Uint64()
		}
		return i
	case starlark.Float:
		return float64(x)
	case starlark.Tuple:
		var result []any
		n := x.Len()
		for i := 0; i < n; i++ {
			v := x.Index(i)
			result = append(result, FromStarlark(v))
		}
		return result
	case *starlark.List:
		n := x.Len()
		var result []any
		for i := 0; i < n; i++ {
			v := x.Index(i)
			result = append(result, FromStarlark(v))
		}
		return result
	case *starlark.Dict:
		result := make(map[string]any)
		for _, item := range x.Items() {
			if len(item) != 2 {
				// Something is weird; bail out.
				return x
			}
			var key string
			switch x := item[0].(type) {
			case starlark.String:
				key = string(x)
			default:
				key = x.String()
			}
			result[key] = FromStarlark(item[1])
		}
		return result
	default:
		return v.String()
	}
}
