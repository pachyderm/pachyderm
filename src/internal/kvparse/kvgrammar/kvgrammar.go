package kvgrammar

import (
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

var Lexer = lexer.Rules{
	"Root": {
		{
			Name:    "Whitespace",
			Pattern: `\s+`,
		},
		{
			Name:    "DoubleQuotedString",
			Pattern: `"`,
			Action:  lexer.Push("DoubleQuotedString"),
		},
		{
			Name:    "SingleQuotedString",
			Pattern: `'`,
			Action:  lexer.Push("SingleQuotedString"),
		},
		{
			Name:    "JSON",
			Pattern: `{`,
			Action:  lexer.Push("JSON"),
		},
		{
			Name:    "Terminator",
			Pattern: `;`,
		},
		{
			Name:    "Equals",
			Pattern: `=`,
		},
		{
			Name:    "String",
			Pattern: `[^;="'\s]+`,
		},
	},
	"DoubleQuotedString": {
		{
			Name:    "Escaped",
			Pattern: `\\(?:["\\abfnrtv]|x[0-9a-fA-F]{2}|u[0-9a-fA-F]{4})`, // go allows \U<hex digit>{8}, but we don't.
		},
		{
			Name:    "DoubleQuotedStringEnd",
			Pattern: `"`,
			Action:  lexer.Pop(),
		},
		{
			Name:    "DoubleQuotedChars",
			Pattern: `[^"\\]+`,
		},
	},
	"SingleQuotedString": {
		{
			Name:    "SingleQuotedStringEnd",
			Pattern: `'`,
			Action:  lexer.Pop(),
		},
		{
			Name:    "SingleQuotedChars",
			Pattern: `[^']+`,
		},
	},
	"JSON": {
		{
			Name:    "JSONEnd",
			Pattern: `}`,
			Action:  lexer.Pop(),
		},
		{
			Name:    "Whitespace",
			Pattern: `\s+`,
		},
		{
			Name:    "DoubleQuotedString",
			Pattern: `"`,
			Action:  lexer.Push("DoubleQuotedString"),
		},
		{
			Name:    "Colon",
			Pattern: `:`,
		},
		{
			Name:    "Comma",
			Pattern: `,`,
		},
	},
}

type DoubleQuotedPart struct {
	Escaped *string `parser:"(@Escaped"`
	String  *string `parser:"|@DoubleQuotedChars)"`
}

type DoubleQuoted struct {
	Start string             `parser:"@DoubleQuotedString"`
	Parts []DoubleQuotedPart `parser:"@@+"`
	End   string             `parser:"@DoubleQuotedStringEnd"`
}

func (s *DoubleQuoted) String() string {
	var buf strings.Builder
	for _, part := range s.Parts {
		if s := part.String; s != nil {
			buf.WriteString(*s)
		}
		if e := part.Escaped; e != nil {
			val, _, _, err := strconv.UnquoteChar(*e, '"')
			if err != nil {
				buf.WriteString(*e) // example: \ud800; this isn't a character, but is "grammatically" correct
				continue
			}
			buf.WriteRune(val)
		}
	}
	return buf.String()
}

type SingleQuoted struct {
	Start  string `parser:"@SingleQuotedString"`
	String string `parser:"@SingleQuotedChars"`
	End    string `parser:"@SingleQuotedStringEnd"`
}

type Str struct {
	Literal      string        `parser:"(@String"`
	DoubleQuoted *DoubleQuoted `parser:"| @@"`
	SingleQuoted *SingleQuoted `parser:"| @@)"`
}

func (s Str) String() string {
	if s.Literal != "" {
		return s.Literal
	}
	if s.DoubleQuoted != nil {
		return s.DoubleQuoted.String()
	}
	if s.SingleQuoted != nil {
		return s.SingleQuoted.String
	}
	return ""
}

type KV struct {
	Key        Str    `parser:"@@"`
	Equals     string `parser:"@Equals"`
	Value      Str    `parser:"@@?"` // The value can be the empty string.
	Terminator string `parser:"(@Terminator|@EOF)"`
}

type JSONPair struct {
	Key   DoubleQuoted `parser:"@@"`
	Colon string       `parser:"@Colon"`
	Value DoubleQuoted `parser:"@@"`
}

type TerminatedJSONPair struct {
	Key   DoubleQuoted `parser:"@@"`
	Colon string       `parser:"@Colon"`
	Value DoubleQuoted `parser:"@@"`
	Comma string       `parser:"@Comma?"`
}

type JSON struct {
	Start string               `parser:"@JSON"`
	Pairs []TerminatedJSONPair `parser:"@@*"`
	End   string               `parser:"@JSONEnd"`
}

type KVs struct {
	JSON  JSON `parser:"(@@"`
	Pairs []KV `parser:"|@@+)"`
}
