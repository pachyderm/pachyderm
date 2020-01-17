// json_decoder.go implements the structfmt.Decoder interfaces for the json text
// format. This library is nearly unnecessary, except that pachyderm/yaml cannot
// parse JSON structures with a leading tab. This minor difference affects
// pps/cmds/cmds_test.go, so until the parser is fixed or the tests changed, we
// need to use this to parse pipeline specs, for backwards compatibility

package serde

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// DecodeJSON is a convenience function that decodes json data using a
// JSONDecoder, but can be called inline
func DecodeJSON(jsonData []byte, v interface{}) error {
	return json.Unmarshal(jsonData, v)
}

// JSONDecoder is an implementation of serde.Decoder that operates on JSON data.
type JSONDecoder struct {
	d *json.Decoder
}

// NewJSONDecoder returns a new JSONDecoder that reads from 'r'
func NewJSONDecoder(r io.Reader) *JSONDecoder {
	return &JSONDecoder{d: json.NewDecoder(r)}
}

// Decode implements the corresponding method of serde.Decoder
func (d *JSONDecoder) Decode(v interface{}) error {
	// fast path
	return d.d.Decode(v)
}

// DecodeTransform implements the corresponding method of serde.Decoder
func (d *JSONDecoder) DecodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	jsonBytes, err := d.transformDecode(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	if err := json.Unmarshal(jsonBytes, v); err != nil {
		return fmt.Errorf("parse error while canonicalizing json: %v", err)
	}
	return nil
}

// DecodeProto implements the corresponding method of serde.Decoder
func (d *JSONDecoder) DecodeProto(v proto.Message) error {
	// fast path
	m := &jsonpb.Unmarshaler{}
	return m.UnmarshalNext(d.d, v)
}

// DecodeProtoTransform implements the corresponding method of serde.Decoder
func (d *JSONDecoder) DecodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	jsonBytes, err := d.transformDecode(f)
	if err != nil {
		return err
	}

	// parse transformed JSON into 'v'
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	if err := jsonpb.UnmarshalNext(decoder, v); err != nil {
		return fmt.Errorf("error canonicalizing json while parsing to proto: %v", err)
	}
	return nil
}

func (d *JSONDecoder) transformDecode(f func(map[string]interface{}) error) ([]byte, error) {
	// deserialize json into 'holder'
	holder := map[string]interface{}{}
	if err := d.d.Decode(&holder); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("could not parse json: %v", err)
	}

	// transform 'holder' (e.g. stringifying TFJob)
	if f != nil {
		if err := f(holder); err != nil {
			return nil, err
		}
	}

	// serialize 'holder' to json
	jsonBytes, err := json.Marshal(holder)
	if err != nil {
		return nil, fmt.Errorf("serialization error while canonicalizing json: %v", err)
	}
	return jsonBytes, nil
}
