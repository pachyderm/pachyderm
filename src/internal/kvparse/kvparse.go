// Package kvparse implements a language for representing key/value pairs.
package kvparse

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/kvparse/kvgrammar"
)

// KV is a parsed key/value pair.
type KV struct {
	Key   string
	Value string
}

var lex = lexer.MustStateful(kvgrammar.Lexer)

func ParseOne(in string) (KV, error) {
	parser, err := participle.Build[kvgrammar.KV](participle.Lexer(lex), participle.Elide("Whitespace"))
	if err != nil {
		return KV{}, errors.Wrap(err, "build kv parser")
	}
	kv, err := parser.ParseString("", in)
	if err != nil {
		return KV{}, errors.Wrap(err, "parse kv")
	}
	return KV{Key: kv.Key.String(), Value: kv.Value.String()}, nil
}

func ParseMany(in string) ([]KV, error) {
	if strings.TrimSpace(in) == "" {
		return nil, nil
	}
	parser, err := participle.Build[kvgrammar.KVs](participle.Lexer(lex), participle.Elide("Whitespace"))
	if err != nil {
		return nil, errors.Wrap(err, "build kv parser")
	}
	kvs, err := parser.ParseString("", in)
	if err != nil {
		return nil, errors.Wrap(err, "parse kv")
	}
	var result []KV
	for _, kv := range kvs.Pairs {
		result = append(result, KV{
			Key:   kv.Key.String(),
			Value: kv.Value.String(),
		})
	}
	for _, kv := range kvs.JSON.Pairs {
		result = append(result, KV{
			Key:   kv.Key.String(),
			Value: kv.Value.String(),
		})
	}
	return result, nil
}
