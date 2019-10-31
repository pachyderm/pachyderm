package serde

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

type JSONEncoder struct {
	w io.Writer
	e *json.Encoder

	// origName sets whether this YAMLEncoder uses the original (.proto) name of
	// fields when marshalling to protos
	origName bool

	// indentSpaces determines the number of spaces used at each layer of
	// indentation when encoding data
	indentSpaces int
}

func EncodeJSON(v interface{}, options ...EncoderOption) ([]byte, error) {
	var buf bytes.Buffer
	e := NewJSONEncoder(&buf, options...)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func NewJSONEncoder(w io.Writer, options ...EncoderOption) *JSONEncoder {
	e := &JSONEncoder{w: w, e: json.NewEncoder(w)}
	for _, o := range options {
		o(e)
	}
	return e
}

func (e *JSONEncoder) Encode(v interface{}) error {
	// shortcut encoding
	return e.e.Encode(v)
}

func (e *JSONEncoder) EncodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	var buf bytes.Buffer
	j := json.NewEncoder(&buf)
	if err := j.Encode(v); err != nil {
		return fmt.Errorf("serialization error while canonicalizing output: %v", err)
	}

	return e.encodeTransform(buf.Bytes(), f)
}

func (e *JSONEncoder) EncodeProto(v proto.Message) error {
	// shortcut encoding
	m := jsonpb.Marshaler{
		OrigName: e.origName,
		Indent:   strings.Repeat(" ", e.indentSpaces),
	}
	return m.Marshal(e.w, v)
}

func (e *JSONEncoder) EncodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	var buf bytes.Buffer
	m := jsonpb.Marshaler{
		OrigName: e.origName,
	} // output text is produced by 'e.e', not 'm', so no need to set 'Indent'
	if err := m.Marshal(&buf, v); err != nil {
		return fmt.Errorf("serialization error while canonicalizing output: %v", err)
	}

	return e.encodeTransform(buf.Bytes(), f)
}

func (e *JSONEncoder) encodeTransform(intermediateJSON []byte,
	f func(map[string]interface{}) error) error {
	// Unmarshal from JSON to intermediate map ('holder')
	holder := map[string]interface{}{}
	if err := json.Unmarshal(intermediateJSON, &holder); err != nil {
		return fmt.Errorf("deserialization error while canonicalizing output: %v", err)
	}

	// transform 'holder' (e.g. de-stringifying TFJob)
	if f != nil {
		if err := f(holder); err != nil {
			return err
		}
	}

	// Encode 'holder' to back to JSON to be re-encoded into a struct or protobuf
	if err := e.e.Encode(holder); err != nil {
		return fmt.Errorf("serialization error while canonicalizing json: %v", err)
	}
	return nil
}
