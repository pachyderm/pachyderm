/*
Package env handles environment variables in a structured manner.

See the README at https://github.com/peter-edge/go-env/blob/master/README.md for more details.
*/
package env

import (
	"io"
	"reflect"
)

// Decoder decodes an env file.
type Decoder interface {
	Decode() (map[string]string, error)
}

// NewEnvFileDecoder returns a new Decoder that decodes an env file
// of the form key=value.
func NewEnvFileDecoder(reader io.Reader) Decoder {
	return newEnvFileDecoder(reader)
}

// NewJSONDecoder returns a new Decoder that decodes a JSON object.
func NewJSONDecoder(reader io.Reader) Decoder {
	return newJSONDecoder(reader)
}

// PopulateOptions are the options to pass to Populate.
type PopulateOptions struct {
	// RestrictTo is the set of env keys that are allowed
	// to be read by structs. Use at the application level to make sure
	// no env keys are being read that are not made aware of
	RestrictTo []string
	// Decoders is a list of Decoders to use, in order,
	// to read additional env variables.
	Decoders []Decoder
	// Defaults sets default values if no value is present in the environment.
	Defaults map[string]string
}

// Populate populates an object with environment variables.
// See the test for an example.
func Populate(object interface{}, populateOptions PopulateOptions) error {
	return populate(reflect.ValueOf(object), populateOptions, false)
}
