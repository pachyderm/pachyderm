// structfmt is a super-package that contains Pachyderm-specific libraries for
// marshalling and unmarshalling Go structs and maps to structured text formats
// (currently just json and yaml).
//
// Similar to https://github.com/ghodss/yaml, all implementations of the Format
// interface marshal and unmarshal data using the following pipeline:
//
//    Go struct/map (fully structured)
//      <-> JSON document
//      <-> map[string]interface{}
//      <-> target format document
//
// Despite the redundant round of marshalling and unmarshalling, there are two
// main advantages to this approach:
//   - YAML (and any future storage formats) can re-use existing `json:` struct
//     tags
//   - The intermediate map[string]interface{} can be manipulated, making it
//     possible to have flexible converstions between structs and documents. For
//     examples, PPS pipelines may include a full TFJob spec, which is converted
//     to a string and stored in the 'TFJob' field of Pachyderm's
//     CreatePipelineRequest struct.
package serde

import (
	"fmt"
	"io"
	"strings"

	"github.com/gogo/protobuf/proto"
)

// EncoderOptions are functions that modify the behavior of an encoder and can
// be passed to GetEncoder.
//
// Currently all EncoderOptions are encoder-agnostic (i.e. the option works on
// both the JSON encoder and the YAML encoder). Future EncoderOptions that are
// only compatible with a single encoder type, however, should silently ignore
// any passed Encoders of other types. For example, if a user calls GetEncoder
// requesting a JSON encoder, but also passes an option that only applies to
// YAML encoders, then the option should exit silently. This makes it possible
// to pass all options to GetEncoder, but defer the decision about what kind of
// encoding to use until runtime, like so:
//
// enc, _ := GetEncoder(outputFlag, os.Stdout,
//     ...options to use if json...,
//     ...options to use if yaml...,
// )
// enc.Encode(obj)
type EncoderOption func(Encoder)

// WithOrigName is an EncoderOption that, if set, encodes proto messages using
// the name set in the original .proto message definition, rather than the
// munged version of the generated struct's field name
func WithOrigName(origName bool) func(d Encoder) {
	return func(d Encoder) {
		switch e := d.(type) {
		case *YAMLEncoder:
			e.origName = origName
		case *JSONEncoder:
			e.origName = origName
		}
	}
}

func WithIndent(numSpaces int) func(d Encoder) {
	return func(d Encoder) {
		switch e := d.(type) {
		case *YAMLEncoder:
			e.e.SetIndent(numSpaces)
		case *JSONEncoder:
			e.e.SetIndent("", strings.Repeat(" ", numSpaces))
			e.indentSpaces = numSpaces // used by EncodeProto shortcut
		}
	}
}

// GetEncoder dynamically creates and returns an Encoder for the text format
// 'encoding' (currently, 'encoding' must be "yaml" or "json").
// 'defaultEncoding' specifies the text format that should be used if 'encoding'
// is "". 'opts' are the list of options that should be applied to any result,
// if any are applicable.  Typically EncoderOptions are encoder-specific (e.g.
// setting json indentation). If an option is passed to GetEncoder for e.g. a
// json encoder but a yaml encoder is requested, then the option will be
// ignored. This makes it possible to pass all options to GetEncoder, but defer
// the decision about what kind of encoding to use until runtime, like so:
//
// enc, _ := GetEncoder(outputFlag, os.Stdout,
//     ...options to use if json...,
//     ...options to use if yaml...,
// )
// enc.Encode(obj)
//
// Note: There is currently no corresponding GetDecoder, because the only
// implementation of the Decoder interface is YAMLDecoder
func GetEncoder(encoding string, w io.Writer, opts ...EncoderOption) (Encoder, error) {
	switch encoding {
	case "yaml":
		return NewYAMLEncoder(w, opts...), nil
	case "json":
		return NewJSONEncoder(w, opts...), nil
	default:
		return nil, fmt.Errorf("unrecognized encoding: %q (must be \"yaml\" or \"json\")", encoding)
	}
}

// Encoder is an interface for encoding data to an output stream (every
// implementation should provide the ability to construct an Encoder tied to an
// output stream, to which encoded text should be written)
type Encoder interface {
	// Marshall converts 'v' (a struct or Go map) to a structured-text document
	// and writes it to this Encoder's output stream
	Encode(v interface{}) error

	// EncodeProto is similar to Encode, but instead of converting between the
	// canonicalized JSON and 'v' using 'encoding/json', it does so using
	// 'gogo/protobuf/jsonpb'.  This allows callers to take advantage of more
	// sophisticated timestamp parsing and such in the 'jsonpb' library.
	//
	// TODO(msteffen): We can *almost* avoid the Encode/EncodeProto split by
	// checking if 'v' implements 'proto.Message', except for one case: the
	// kubernetes client library includes structs that are pseudo-protobufs.
	// Structs in the kubernetes go client implement the 'proto.Message()'
	// interface but are hand-generated and contain embedded structs, which
	// 'jsonpb' can't handle when parsing. If jsonpb is ever extended to be able
	// to parse JSON into embedded structs (even though the protobuf compiler
	// itself never generates such structs) then we could merge this into
	// Encode() and rely on:
	//
	//   if msg, ok := v.(proto.Message); ok {
	//     ... use jsonpb ...
	//   } else {
	//     ... use encoding/json ...
	//   }
	EncodeProto(proto.Message) error

	// EncodeTransform is similar to Encode, but users can manipulate the
	// intermediate map[string]interface generated by Format implementations by
	// passing a function. Note that 'Encode(v)' is equivalent to
	// 'EncodeTransform(v, nil)'
	EncodeTransform(interface{}, func(map[string]interface{}) error) error

	// EncodeProtoTransform is similar to EncodeTransform(), but instead of
	// converting between the canonicalized JSON and 'v' using 'encoding/json', it
	// does so using 'gogo/protobuf/jsonpb'.  This allows callers to take
	// advantage of more sophisticated timestamp parsing and such in the 'jsonpb'
	// library.
	//
	// TODO(msteffen) same comment re: proto.Message as for EncodeProto()
	EncodeProtoTransform(proto.Message, func(map[string]interface{}) error) error
}

// Decoder is the interface for decoding a particular serialization format.
// Currently, the only implementation of Decoder is in serde/yaml/decoder.go
// (the point of having a common interface is to make it possible to switch
// between decoding different serialization formats at runtime, but we have no
// need to switch between a YAML decoder and a JSON decoder, as our YAML decoder
// can decode JSON).
type Decoder interface {
	// Decode parses an underlying stream of text in the given format into 'v'
	// (a struct or Go map)
	Decode(interface{}) error

	// DecodeProto is similar to Decode, but instead of converting between
	// the canonicalized JSON and 'v' using 'encoding/json', it does so using
	// 'gogo/protobuf/jsonpb'.  This allows callers to take advantage of more
	// sophisticated timestamp parsing and such in the 'jsonpb' library.
	//
	// TODO(msteffen) same comment re: proto.Message as for
	// Encoder.EncodeProto()
	DecodeProto(proto.Message) error

	// DecodeTransform is similar to Decode, but users can manipulate the
	// intermediate map[string]interface generated by Format implementations by
	// passing a function. Note that 'Encode(v)' is equivalent to
	// 'EncodeTransform(v, nil)'
	DecodeTransform(interface{}, func(map[string]interface{}) error) error

	// DecodeProtoTransform is similar to DecodeTransform, but instead of
	// converting between the canonicalized JSON and 'v' using 'encoding/json', it
	// does so using 'gogo/protobuf/jsonpb'.  This allows callers to take
	// advantage of more sophisticated timestamp parsing and such in the 'jsonpb'
	// library.
	//
	// TODO(msteffen) same comment re: proto.Message as for
	// Encoder.EncodeProto()
	DecodeProtoTransform(proto.Message, func(map[string]interface{}) error) error
}
