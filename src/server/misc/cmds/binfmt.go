package cmds

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var asciiSpace = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1}

func removeSpaces(s []byte) int {
	var j int
	for _, b := range s {
		if asciiSpace[b] == 0 {
			s[j] = b
			j++
		}
	}
	return j
}

func decodeBinfmt(input []byte, format string) ([]byte, error) {
	// Decode the binary encoding.
	var raw []byte
	switch format {
	case "hex":
		// Hex format, like "369cf1215181b1e".

		// Remove spaces, because they can't mean anything.
		n := removeSpaces(input)
		input = input[:n]

		// Remove 0x, which postgres prefixes hex binary data with.
		input = bytes.TrimPrefix(input, []byte("0x"))

		// Decode.
		raw = make([]byte, hex.DecodedLen(len(input)))
		n, err := hex.Decode(raw, input)
		if err != nil {
			return nil, errors.Wrap(err, "decode hex")
		}
		raw = raw[:n]
	case "base64":
		// Normal base64.

		// Remove spaces, because they can't mean anything.
		n := removeSpaces(input)
		input = input[:n]

		// Decode.
		raw = make([]byte, base64.StdEncoding.DecodedLen(len(input)))
		n, err := base64.StdEncoding.Decode(raw, input)
		if err != nil {
			return nil, errors.Wrap(err, "decode base64")
		}
		raw = raw[:n]
	case "raw":
		// Passthrough.
		raw = input
	default:
		return nil, errors.Errorf("no known input format %q", format)
	}
	return raw, nil
}
