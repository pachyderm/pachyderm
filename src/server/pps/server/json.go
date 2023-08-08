package server

import (
	"encoding/json"
	"strconv"
	"sync"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type canonicalizer func(value any) (any, error)

var canonicalizerMap sync.Map

func identityCanonicalizer(value any) (any, error) {
	return value, nil
}

func makeEnumCanonicalizer(d protoreflect.EnumDescriptor) canonicalizer {
	if c, ok := canonicalizerMap.Load(d.FullName()); ok {
		return c.(canonicalizer)
	}
	var (
		values    = d.Values()
		numValues = values.Len()
		numberMap = make(map[json.Number]string, numValues)
	)
	for i := 0; i < numValues; i++ {
		value := values.Get(i)
		numberMap[json.Number(strconv.Itoa(int(value.Number())))] = string(value.Name())
	}
	var c canonicalizer = func(value any) (any, error) {
		switch value := value.(type) {
		case json.Number:
			if n, ok := numberMap[value]; ok {
				return n, nil
			}
			return value, nil
		case string:
			return value, nil
		default:
			return nil, errors.Errorf("expected number or string; got %T", value)
		}
	}
	canonicalizerMap.Store(d.FullName(), c)
	return c
}

func makeMessageCanonicalizer(d protoreflect.MessageDescriptor) (canonicalizer, error) {
	if c, ok := canonicalizerMap.Load(d.FullName()); ok {
		return c.(canonicalizer), nil
	}
	var (
		fields              = d.Fields()
		fieldsLen           = fields.Len()
		fieldMap            = make(map[string]protoreflect.FieldDescriptor, 2*fieldsLen)
		fieldCanonicalizers = make(map[string]canonicalizer, 2*fieldsLen)
	)

	for i := 0; i < fieldsLen; i++ {
		var (
			field    = fields.Get(i)
			name     = string(field.Name())
			jsonName = field.JSONName()
		)
		fieldMap[name] = field
		fieldMap[jsonName] = field
		switch field.Kind() {
		case protoreflect.BoolKind,
			protoreflect.Int32Kind,
			protoreflect.Sint32Kind,
			protoreflect.Uint32Kind,
			protoreflect.Int64Kind,
			protoreflect.Sint64Kind,
			protoreflect.Uint64Kind,
			protoreflect.Sfixed32Kind,
			protoreflect.Fixed32Kind,
			protoreflect.FloatKind,
			protoreflect.Sfixed64Kind,
			protoreflect.Fixed64Kind,
			protoreflect.DoubleKind,
			protoreflect.StringKind,
			protoreflect.BytesKind:
			fieldCanonicalizers[name] = identityCanonicalizer
			fieldCanonicalizers[jsonName] = identityCanonicalizer
		case protoreflect.EnumKind:
			// FIXME: canonicalise to strings
			c := makeEnumCanonicalizer(field.Enum())
			fieldCanonicalizers[name] = c
			fieldCanonicalizers[jsonName] = c
		case protoreflect.MessageKind:
			// have to defer creation of the canonicalizer until runtime due to recursive messages
			c := func(d protoreflect.MessageDescriptor) canonicalizer {
				return func(value any) (any, error) {
					c, err := makeMessageCanonicalizer(d)
					if err != nil {
						return nil, errors.Wrapf(err, "could not make nested canonicalizer for %s", d.FullName())
					}
					return c(value)
				}
			}(field.Message())
			fieldCanonicalizers[name] = c
			fieldCanonicalizers[jsonName] = c
		default:
			return nil, errors.Errorf("don’t know how to canonicalize %s", field.Kind())
		}
	}
	var c canonicalizer = func(value any) (any, error) {
		valueMap, ok := value.(map[string]any)
		if !ok {
			return nil, errors.Errorf("expected map[string]any for %s; got %T", d.FullName(), value)
		}
		var newValue = make(map[string]any)
		for k, v := range valueMap {
			f, ok := fieldMap[k]
			if !ok {
				return nil, errors.Errorf("unexpected field %q in %s", k, d.FullName())
			}
			// FIXME: this can be checked once at make-time
			switch {
			case f.IsList():
				vv, ok := v.([]any)
				if !ok {
					return nil, errors.Errorf("expected []any; got %T in %s", v, f.FullName())
				}
				var l = make([]any, len(vv))
				for i, o := range vv {
					var err error
					if l[i], err = fieldCanonicalizers[k](o); err != nil {
						return nil, errors.Wrapf(err, "couldn’t canonicalize %s", f.FullName())
					}
				}
				newValue[f.JSONName()] = l
			case f.IsMap():
				var (
					m   = make(map[string]any)
					err error
				)
				vm, ok := v.(map[string]any)
				if !ok {
					return nil, errors.Errorf("expected map[string]any; got %T in %s", v, f.FullName())
				}
				for kk, vv := range vm {
					var o any
					if o, err = fieldCanonicalizers[k](map[string]any{"key": kk, "value": vv}); err != nil {
						return nil, errors.Wrapf(err, "couldn’t canonicalize %s[%s] (%v)", f.FullName(), kk, vv)
					}
					oo, ok := o.(map[string]any)
					if !ok {
						return nil, errors.Errorf("map canonicalizer: expected map[string]any; got %T", o)
					}
					m[kk] = oo["value"]
				}
				newValue[f.JSONName()] = m
			default:
				var err error
				newValue[f.JSONName()], err = fieldCanonicalizers[k](v)
				if err != nil {
					return nil, errors.Wrapf(err, "could not canonicalize field %s (%v)", f.FullName(), v)
				}
			}
		}
		return newValue, nil
	}
	canonicalizerMap.Store(d.FullName(), c)
	return c, nil
}
