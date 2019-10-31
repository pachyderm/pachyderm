// yaml_decoder.go implements the structfmt.{Encoder,Decoder} interfaces for the
// yaml text format

package serde

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"gopkg.in/pachyderm/yaml.v3"
)

func DecodeYAML(yamlData []byte, v interface{}) error {
	d := NewYAMLDecoder(bytes.NewReader(yamlData))
	return d.Decode(v)
}

type YAMLDecoder struct {
	d *yaml.Decoder
}

func NewYAMLDecoder(r io.Reader) *YAMLDecoder {
	return &YAMLDecoder{d: yaml.NewDecoder(r)}
}

func (d *YAMLDecoder) Decode(v interface{}) error {
	return d.DecodeTransform(v, nil)
}

func (d *YAMLDecoder) DecodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	jsonData, err := d.yamlToJSONTransform(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	if err := json.Unmarshal(jsonData, v); err != nil {
		return fmt.Errorf("parse error while canonicalizing yaml: %v", err)
	}
	return nil
}

func (d *YAMLDecoder) DecodeProto(v proto.Message) error {
	return d.DecodeProtoTransform(v, nil)
}

func (d *YAMLDecoder) DecodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	jsonData, err := d.yamlToJSONTransform(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	if err := jsonpb.UnmarshalNext(decoder, v); err != nil {
		return fmt.Errorf("error canonicalizing yaml while parsing to proto: %v", err)
	}
	return nil
}

func (d *YAMLDecoder) yamlToJSONTransform(f func(map[string]interface{}) error) ([]byte, error) {
	// deserialize yaml into 'holder'
	holder := map[string]interface{}{}
	if err := d.d.Decode(&holder); err != nil {
		return nil, fmt.Errorf("could not parse yaml: %v", err)
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
		return nil, fmt.Errorf("serialization error while canonicalizing yaml: %v", err)
	}
	return jsonData, nil
}
