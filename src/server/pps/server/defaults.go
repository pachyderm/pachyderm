package server

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	v1 "k8s.io/api/core/v1"

	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
		case nil:
			return nil, nil
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
	switch d.FullName() {
	case "google.protobuf.Timestamp",
		"google.protobuf.Duration",
		"google.protobuf.Any",
		"google.protobuf.Struct",
		"google.protobuf.ListValue",
		"google.protobuf.Value",
		"google.protobuf.FieldMask",
		"google.protobuf.Empty",
		"google.protobuf.Int32Value",
		"google.protobuf.Int64Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.FloatValue",
		"google.protobuf.DoubleValue",
		"google.protobuf.StringValue",
		"google.protobuf.BytesValue",
		"google.protobuf.NullValue":
		return identityCanonicalizer, nil
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
					if value == nil {
						return nil, nil
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
			// TODO(CORE-1897): This can be checked once at
			// make-time.  Requires a little bit of thought to do so
			// without instantiating the field canonicalizer at
			// make-time.
			switch {
			case f.IsList():
				vv, ok := v.([]any)
				if !ok {
					return nil, errors.Errorf("expected []any; got %T in %s", v, f.FullName())
				}
				var l = make([]any, len(vv))
				for i, o := range vv {
					var err error
					if o == nil {
						l[i] = nil
						continue
					}
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
					if vv == nil {
						m[kk] = nil
						continue
					}
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
				if v == nil {
					newValue[f.JSONName()] = nil
					continue
				}
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

// jsonMergePatch merges a JSON patch in string form with a JSON target, also in
// string form.
func jsonMergePatch(target, patch string, f canonicalizer) (string, error) {
	var targetObject, patchObject any
	// The default json decoder will decode numbers to floats, which can
	// lose precision; by explicitly creating a decoder and using
	// json.Number we avoid that.
	if err := unmarshalJSON(target, &targetObject); err != nil {
		return "", errors.Wrap(err, "could not unmarshal target JSON")
	}
	if err := unmarshalJSON(patch, &patchObject); err != nil {
		return "", errors.Wrap(err, "could not unmarshal patch JSON")
	}
	if f != nil {
		var err error
		if targetObject, err = f(targetObject); err != nil {
			return "", errors.Wrap(err, "could not canonicalize target object")
		}
		if patchObject, err = f(patchObject); err != nil {
			return "", errors.Wrap(err, "could not canonicalize target object")
		}
	}
	result, err := json.Marshal(mergePatch(targetObject, patchObject))
	if err != nil {
		return "", errors.Wrap(err, "could not marshal merge patch result")
	}
	return string(result), nil
}

// mergePatch implements the RFC 7396 algorithm.  To quote the RFC “If the patch
// is anything other than an object, the result will always be to replace the
// entire target with the entire patch.  Also, it is not possible to patch part
// of a target that is not an object, such as to replace just some of the values
// in an array.”  If the patch _is_ an object, then non-null values replace
// target values, and null values delete target values.
func mergePatch(target, patch any) any {
	switch patch := patch.(type) {
	case map[string]any:
		var targetMap map[string]any
		switch t := target.(type) {
		case map[string]any:
			targetMap = t
		default:
			targetMap = make(map[string]any)
		}
		for name, value := range patch {
			if value == nil {
				delete(targetMap, name)
			} else {
				targetMap[name] = mergePatch(targetMap[name], value)
			}
		}
		return targetMap
	default:
		return patch
	}
}

var clusterDefaultsCanonicalizer canonicalizer

func init() {
	var err error
	clusterDefaultsCanonicalizer, err = makeMessageCanonicalizer((&pps.ClusterDefaults{}).ProtoReflect().Descriptor())
	if err != nil {
		panic(fmt.Sprintf("could not make ClusterDefaults canonicalizer: %v", err))
	}
}

// makeEffectiveSpec creates an effective spec from the cluster defaults (a
// JSON-encoded ClusterDefaults) and user spec (a JSON-encoded
// CreatePipelineRequest) by merging the user spec into the cluster defaults.
// It returns the effective spec as both JSON and a CreatePipelineRequest.
func makeEffectiveSpec(clusterDefaultsJSON, userSpecJSON string) (string, *pps.CreatePipelineRequest, error) {
	type wrapper struct {
		CreatePipelineRequest json.RawMessage `json:"createPipelineRequest"`
	}
	userWrapper, err := json.Marshal(wrapper{CreatePipelineRequest: []byte(userSpecJSON)})
	if err != nil {
		return "", nil, errors.Wrapf(err, "could not marshal user spec %s", userSpecJSON)
	}
	wrappedSpecJSON, err := jsonMergePatch(clusterDefaultsJSON, string(userWrapper), clusterDefaultsCanonicalizer)
	if err != nil {
		return "", nil, errors.Wrapf(err, "could not merge user wrapper %s into cluster defaults %s", string(userWrapper), clusterDefaultsJSON)
	}
	d := json.NewDecoder(strings.NewReader(wrappedSpecJSON))
	var w map[string]any
	if err := d.Decode(&w); err != nil {
		return "", nil, errors.Wrapf(err, "could not unmarshal wrapped spec %s", wrappedSpecJSON)
	}
	createPipelineRequest, ok := w["createPipelineRequest"]
	if !ok {
		return "", nil, errors.Wrapf(err, "missing createPipelineRequest in wrapped spec %s", wrappedSpecJSON)
	}
	effectiveSpecJSON, err := json.Marshal(createPipelineRequest)
	if err != nil {
		return "", nil, errors.Wrapf(err, "could not marshal effective spec %v", createPipelineRequest)
	}

	var effectiveWrapper pps.ClusterDefaults
	if err := protojson.Unmarshal([]byte(wrappedSpecJSON), &effectiveWrapper); err != nil {
		return "", nil, errors.Wrapf(err, "could not unmarshal effective spec %s", wrappedSpecJSON)
	}
	if err := validateSpec(effectiveWrapper.CreatePipelineRequest); err != nil {
		return "", nil, errors.Wrapf(err, "invalid effective spec %s", string(effectiveSpecJSON))
	}
	return string(effectiveSpecJSON), effectiveWrapper.CreatePipelineRequest, nil
}

func unmarshalJSON(s string, v any) error {
	d := json.NewDecoder(strings.NewReader(s))
	d.UseNumber()
	return errors.Wrapf(d.Decode(v), "could not unmarshal %q as JSON", s)
}

func validateSpec(req *pps.CreatePipelineRequest) error {
	// TODO(msteffen) eventually TFJob and Transform will be alternatives, but
	// currently TFJob isn't supported
	if req.TfJob != nil {
		return errors.New("embedding TFJobs in pipelines is not supported yet")
	}
	if !(req.ReprocessSpec == "" || req.ReprocessSpec == client.ReprocessSpecUntilSuccess && req.ReprocessSpec != client.ReprocessSpecEveryJob) {
		return errors.Errorf("invalid pipeline spec: ReprocessSpec must be one of %q or %q", client.ReprocessSpecUntilSuccess, client.ReprocessSpecEveryJob)
	}
	var tolErrs error
	for i, t := range req.GetTolerations() {
		if _, err := transformToleration(t); err != nil {
			errors.JoinInto(&tolErrs, errors.Errorf("toleration %d/%d: %v", i+1, len(req.GetTolerations()), err))
		}
	}
	if tolErrs != nil {
		return tolErrs
	}

	if req.PodSpec != "" && !json.Valid([]byte(req.PodSpec)) {
		return errors.Errorf("malformed PodSpec")
	}
	if req.PodPatch != "" && !json.Valid([]byte(req.PodPatch)) {
		return errors.Errorf("malformed PodPatch")
	}
	if req.Service != nil {
		validServiceTypes := map[v1.ServiceType]bool{
			v1.ServiceTypeClusterIP:    true,
			v1.ServiceTypeLoadBalancer: true,
			v1.ServiceTypeNodePort:     true,
		}

		if !validServiceTypes[v1.ServiceType(req.Service.Type)] {
			return errors.Errorf("the following service type %s is not allowed", req.Service.Type)
		}
	}
	return nil
}
