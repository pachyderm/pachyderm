// json_encoder.go is similar to yaml_encoder.go; see its comment on general
// approach. Because the final output is json, the steps this uses are a little
// more redundant, but still the reverse of the steps in decoder.go (allowing
// for future-proof reversability):
//  1. Serialize structs to JSON using encoding/json or, in the case of protobufs, using Google's
//     'protojson' library.
//  2. Deserialize that text into a generic interface{}.
//  3. Serialize that interface using encoding/json.
//
// Because the data is serialized to JSON twice, this approach seems more
// redundant in the context of json_encoder.go, but it's symmetrical with the
// encoding process.
// TODO(msteffen) src/server/identity/cmds/cmds.go uses this to serialize
// Pachyderm identity configs. That code could be written to modify the holder
// (map[string]interface{}) instead of using connectorConfig as an intermediate
// struct, which must be kept in sync.
package serde

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// JSONEncoder is an implementation of serde.Encoder that operates on JSON data
type JSONEncoder struct {
	w io.Writer
	e *json.Encoder

	// origName sets whether this JSONEncoder uses the original (.proto) name of
	// fields when marshalling to protos
	origName bool

	// indentSpaces determines the number of spaces used at each layer of
	// indentation when encoding data
	indentSpaces int
}

// EncodeJSON is a convenience function that encodes json data using a
// JSONEncoder, but can be called inline
func EncodeJSON(v interface{}, options ...EncoderOption) ([]byte, error) {
	var buf bytes.Buffer
	e := NewJSONEncoder(&buf, options...)
	if err := e.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NewJSONEncoder returns a new JSONEncoder that writes to 'w'
func NewJSONEncoder(w io.Writer, options ...EncoderOption) *JSONEncoder {
	e := &JSONEncoder{w: w, e: json.NewEncoder(w)}
	for _, o := range options {
		o(e)
	}
	return e
}

// Encode implements the corresponding method of serde.Encoder
func (e *JSONEncoder) Encode(v interface{}) error {
	// shortcut encoding
	return errors.EnsureStack(e.e.Encode(v))
}

// EncodeTransform implements the corresponding method of serde.Encoder
func (e *JSONEncoder) EncodeTransform(v interface{}, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	var buf bytes.Buffer
	j := json.NewEncoder(&buf)
	if err := j.Encode(v); err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing output")
	}

	return e.encodeTransform(buf.Bytes(), f)
}

// EncodeProto implements the corresponding method of serde.Encoder
func (e *JSONEncoder) EncodeProto(v proto.Message) error {
	// shortcut encoding
	m := protojson.MarshalOptions{
		AllowPartial:  true,
		UseProtoNames: e.origName,
		Indent:        strings.Repeat(" ", e.indentSpaces),
	}
	result, err := m.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshal json")
	}
	if _, err := e.w.Write(result); err != nil {
		return errors.Wrap(err, "write")
	}
	return nil
}

// EncodeProtoTransform implements the corresponding method of serde.Encoder
func (e *JSONEncoder) EncodeProtoTransform(v proto.Message, f func(map[string]interface{}) error) error {
	// Encode to JSON first
	m := protojson.MarshalOptions{AllowPartial: true, UseProtoNames: e.origName}

	// output text is produced by 'e.e', not 'm', so no need to set 'Indent'
	result, err := m.Marshal(v)
	if err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing output")
	}
	return e.encodeTransform(result, f)
}

func (e *JSONEncoder) encodeTransform(intermediateJSON []byte,
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

	// Encode 'holder' to back to JSON to be re-encoded into a struct or protobuf
	if err := e.e.Encode(holder); err != nil {
		return errors.Wrapf(err, "serialization error while canonicalizing json")
	}
	return nil
}
