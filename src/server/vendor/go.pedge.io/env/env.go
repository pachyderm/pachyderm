/*
Package env handles environment variables in a structured manner.

See the README at https://github.com/peter-edge/env-go/blob/master/README.md for more details.
*/
package env // import "go.pedge.io/env"

import (
	"fmt"
	"io"
	"os"
	"strings"
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

// Populate populates an object with environment variables.
//
// The environment has precedence over the decoders, earlier
// decoders have precedence over later decoders.
func Populate(object interface{}, decoders ...Decoder) error {
	return populate(object, decoders)
}

// Main runs the common functionality needed in a go main function.
// appEnv will be populated and passed to do, defaultEnv can be nil
// if there is an error, os.Exit(1) will be called.
func Main(do func(interface{}) error, appEnv interface{}, decoders ...Decoder) {
	if err := Populate(appEnv, decoders...); err != nil {
		mainError(err)
	}
	if err := do(appEnv); err != nil {
		mainError(err)
	}
	os.Exit(0)
}

func mainError(err error) {
	if errString := strings.TrimSpace(err.Error()); errString != "" {
		fmt.Fprintf(os.Stderr, "%s\n", errString)
	}
	os.Exit(1)
}
