package cmds

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"

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
	case "go":
		// A go value, like []byte("\x01\x02").  These come from fuzz tests corpora.
		fs := token.NewFileSet()
		expr, err := parser.ParseExprFrom(fs, "(arg)", input, 0)
		if err != nil {
			return nil, errors.Wrap(err, "parse expr")
		}
		switch x := expr.(type) {
		case *ast.CallExpr:
			if len(x.Args) != 1 {
				return nil, errors.New("expected exactly 1 argument to call")
			}
			arg := x.Args[0]

			arrayType, ok := x.Fun.(*ast.ArrayType)
			if !ok {
				return nil, errors.New("expected array constructor")
			}
			if arrayType.Len != nil {
				return nil, errors.New("expected []byte or primitive type")
			}
			elt, ok := arrayType.Elt.(*ast.Ident)
			if !ok || elt.Name != "byte" {
				return nil, errors.New("expected []byte")
			}
			lit, ok := arg.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return nil, errors.New("string literal required for type []byte")
			}
			s, err := strconv.Unquote(lit.Value)
			if err != nil {
				return nil, errors.Wrap(err, "unquote")
			}
			return []byte(s), nil
		case *ast.BasicLit:
			if x.Kind != token.STRING {
				return nil, errors.New("string literal required")
			}
			s, err := strconv.Unquote(x.Value)
			if err != nil {
				return nil, errors.Wrap(err, "unquote")
			}
			return []byte(s), nil
		default:
			return nil, errors.Errorf("unknown go expression %v", expr)
		}
	case "raw":
		// Passthrough.
		raw = input
	default:
		return nil, errors.Errorf("no known input format %q", format)
	}
	return raw, nil
}
