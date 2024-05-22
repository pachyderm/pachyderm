// decoder.go implements general-purpose decoding for hand-written Pachyderm
// data (e.g. pipeline specs, identity/auth configs, pachctl configs, etc). We
// use YAML as our markup language for all such data; yaml is a superset of
// json, so this allows users to write these configs using, effectively, either
// of two common markup languages.
//
// In general, the way Pachyderm decodes a struct from YAML is to:
// 1) Parse YAML to a generic interface{}
// 2) Serialize that interface{} to JSON
// 3) Parse the JSON using encoding/json or, in the case of protobufs, using
//    Google's 'protojson' library.
//
// This approach works around a lot of Go's inherent deserialization quirks
// (e.g. generated protobuf structs including 'json' struct tags but not 'yaml'
// struct tags; see https://web.archive.org/web/20190722213934/http://ghodss.com/2014/the-right-way-to-handle-yaml-in-golang/).
// Using jsonpb for step 3 also allows this library to correctly handle complex
// protobuf corner cases, such as parsing timestamps.

package serde

import (
	"bytes"
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/protoutil"
)

// Decode is a convenience function that decodes the serialized YAML in
// 'yamlData' into the non-proto object 'v'
func Decode(yamlData []byte, v interface{}) error {
	var holder interface{}
	err := yaml.Unmarshal(yamlData, &holder)
	if err != nil {
		return errors.EnsureStack(err)
	}
	return RoundTrip(holder, v)
}

// RoundTrip is a helper function used by Decode(), as well as
// PipelineManifestReader in src/internal/ppsutil/decoder.go. It factors steps 2
// and 3 (described up top) out of Decode(), as step 1 may be done in different
// ways. In particular, PipelineManifestReader may call RoundTrip repeatedly, if
// multiple PipelineSpecs are present in the same YAML doc.
func RoundTrip(holder interface{}, dest interface{}) error {
	requestJSON, err := json.Marshal(holder)
	if err != nil {
		return errors.Wrapf(err,
			"could not serialize to intermediate JSON for final parse: %v",
			holder)
	}
	if pb, ok := dest.(proto.Message); ok {
		r := bytes.NewReader(requestJSON)
		dec := protoutil.NewProtoJSONDecoder(r, protojson.UnmarshalOptions{
			DiscardUnknown: true,
			AllowPartial:   true,
		})
		// NOTE(jonathan): This code used jsonpb.UnmarshalNext before the gogo removal
		// refactoring.  This has been preserved, but I think it's a mistake.
		if err := dec.UnmarshalNext(pb); err != nil {
			return errors.Wrapf(err, "could not parse intermediate JSON to protobuf")
		}
		return nil
	}
	if err := json.Unmarshal(requestJSON, dest); err != nil {
		return errors.Wrapf(err, "could not parse intermediate JSON")
	}
	return nil
}
