// yaml_decoder.go implements the structfmt.Decoder interfaces for the yaml text
// format

package serde

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"gopkg.in/pachyderm/yaml.v3"
)

// DecodeYAML is a convenience function that decodes yaml data using a
// YAMLDecoder, but can be called inline
func DecodeYAML(yamlData []byte, v interface{}) error {
	d := NewYAMLDecoder(bytes.NewReader(yamlData))
	return d.Decode(v)
}

// YAMLDecoder is an implementation of serde.Decoder that operates on yaml data
type YAMLDecoder struct {
	d *yaml.Decoder
}

// NewYAMLDecoder returns a new YAMLDecoder that reads from 'r'
func NewYAMLDecoder(r io.Reader) *YAMLDecoder {
	return &YAMLDecoder{d: yaml.NewDecoder(r)}
}

// Decode implements the corresponding method of serde.Decoder
func (d *YAMLDecoder) Decode(v interface{}) error {
	return d.DecodeTransform(v, nil)
}

// DecodeTransform implements the corresponding method of serde.Decoder
func (d *YAMLDecoder) DecodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	jsonData, err := d.yamlToJSONTransform(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	if err := json.Unmarshal(jsonData, v); err != nil {
		return errors.Wrapf(err, "parse error while canonicalizing yaml")
	}
	return nil
}

// DecodeProto implements the corresponding method of serde.Decoder
func (d *YAMLDecoder) DecodeProto(v proto.Message) error {
	return d.DecodeProtoTransform(v, nil)
}

// DecodeProtoTransform implements the corresponding method of serde.Decoder
func (d *YAMLDecoder) DecodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	jsonData, err := d.yamlToJSONTransform(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	if err := jsonpb.UnmarshalNext(decoder, v); err != nil {
		return errors.Wrapf(err, "error canonicalizing yaml while parsing to proto")
	}
	return nil
}

func (d *YAMLDecoder) yamlToJSONTransform(f func(map[string]interface{}) error) ([]byte, error) {
	// deserialize yaml into 'holder'
	holder := map[string]interface{}{}
	if err := d.d.Decode(&holder); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, errors.Wrapf(err, "could not parse yaml")
	}

	// transform 'holder' (e.g. stringifying TFJob)
	if f != nil {
		if err := f(holder); err != nil {
			return nil, err
		}
	}

	// serialize 'holder' to json
	jsonData, err := json.Marshal(holder)
	if err != nil {
		return nil, errors.Wrapf(err, "serialization error while canonicalizing yaml")
	}
	return jsonData, nil
}
