// yaml_encoder.go is the reverse of decoder.go. It applies the same steps in
// the opposite order:
//  1. Serialize structs to JSON using encoding/json or, in the case of
//     protobufs, using Google's 'protojson' library
//  2. Deserialize that text into a generic interface{}
//  3. Serialize that interface using gopkg.in/yaml.v3
//
// This allows for correct handling of timestamps and such. It also allows data
// that has already been deserialized into a generic interface{} to be written
// out. Finally, it allows for transformations to be applied to the interface{}
// in a structured way (e.g. if the struct being serialized in step 1 contains
// embedded json).
package serde

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"gopkg.in/yaml.v3"
)

// YAMLEncoder is an implementation of serde.Encoder that operates on YAML data
type YAMLEncoder struct {
	e *yaml.Encoder

	// OrigName sets whether this YAMLEncoder uses the original (.proto) name of
	// fields when marshalling to protos
	origName bool
}

// EncodeYAML is a convenience function that encodes yaml data using a
// YAMLEncoder, but can be called inline
func EncodeYAML(v interface{}, options ...EncoderOption) ([]byte, error) {
	var buf bytes.Buffer
	e := NewYAMLEncoder(&buf, options...)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewYAMLEncoder returns a new YAMLEncoder that writes to 'w'
func NewYAMLEncoder(w io.Writer, options ...EncoderOption) *YAMLEncoder {
	e := &YAMLEncoder{e: yaml.NewEncoder(w)}
	for _, o := range options {
		o(e)
	}
	return e
}

// Encode implements the corresponding method of serde.Encoder
func (e *YAMLEncoder) Encode(v interface{}) error {
	return e.EncodeTransform(v, nil)
}

// EncodeTransform implements the corresponding method of serde.Encoder
func (e *YAMLEncoder) EncodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	var buf bytes.Buffer
	j := json.NewEncoder(&buf)
	if err := j.Encode(v); err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing output")
	}

	return e.jsonToYAMLTransform(buf.Bytes(), f)
}

// EncodeProto implements the corresponding method of serde.Encoder
func (e *YAMLEncoder) EncodeProto(v proto.Message) error {
	return e.EncodeProtoTransform(v, nil)
}

// EncodeProtoTransform implements the corresponding method of serde.Encoder
func (e *YAMLEncoder) EncodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	m := protojson.MarshalOptions{
		AllowPartial:  true,
		UseProtoNames: e.origName,
	}
	result, err := m.Marshal(v)
	if err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing output")
	}

	return e.jsonToYAMLTransform(result, f)
}

func (e *YAMLEncoder) jsonToYAMLTransform(intermediateJSON []byte,
	f func(map[string]interface{}) error) error {
	// Unmarshal from JSON to intermediate map ('holder')
	holder := map[string]interface{}{}
	if err := json.Unmarshal(intermediateJSON, &holder); err != nil {
		return errors.Wrapf(err, "deserialization error while canonicalizing output")
	}

	// transform 'holder' (e.g. de-stringifying TFJob)
	if f != nil {
		if err := f(holder); err != nil {
			return err
		}
	}

	// Encode 'holder' to YAML
	if err := e.e.Encode(holder); err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing yaml")
	}
	return nil
}
