package lexer

import (
	"testing"
)

func TestLexGood(t *testing.T) {
	for id, test := range []struct {
		pattern string
		items   []Token
	}{
		{
			pattern: "",
			items: []Token{
				{EOF, ""},
			},
		},
		{
			pattern: "hello",
			items: []Token{
				{Text, "hello"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{rate,[0-9]]}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "rate"},
				{Separator, ","},
				{RangeOpen, "["},
				{Text, "0-9"},
				{RangeClose, "]"},
				{Text, "]"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "/files/(*)/*.jpg",
			items: []Token{
				{Text, "/files/"},
				{CaptureOpen, "@("},
				{Any, "*"},
				{CaptureClose, ")"},
				{Text, "/"},
				{Any, "*"},
				{Text, ".jpg"},
				{EOF, ""},
			},
		},
		{
			pattern: "/files/@(*)/*.jpg",
			items: []Token{
				{Text, "/files/"},
				{CaptureOpen, "@("},
				{Any, "*"},
				{CaptureClose, ")"},
				{Text, "/"},
				{Any, "*"},
				{Text, ".jpg"},
				{EOF, ""},
			},
		},
		{
			pattern: "/files/*@(*)/*.jpg",
			items: []Token{
				{Text, "/files/"},
				{Any, "*"},
				{CaptureOpen, "@("},
				{Any, "*"},
				{CaptureClose, ")"},
				{Text, "/"},
				{Any, "*"},
				{Text, ".jpg"},
				{EOF, ""},
			},
		},
		{
			pattern: "/files/*(*)/*.jpg",
			items: []Token{
				{Text, "/files/"},
				{CaptureOpen, "*("},
				{Any, "*"},
				{CaptureClose, ")"},
				{Text, "/"},
				{Any, "*"},
				{Text, ".jpg"},
				{EOF, ""},
			},
		},
		{
			pattern: "/files/+*(*)/*.jpg",
			items: []Token{
				{Text, "/files/"},
				{Text, "+"},
				{CaptureOpen, "*("},
				{Any, "*"},
				{CaptureClose, ")"},
				{Text, "/"},
				{Any, "*"},
				{Text, ".jpg"},
				{EOF, ""},
			},
		},
		{
			pattern: "@(b+(c)d|e*(f)g?|?(h)i@(j|k))",
			items: []Token{
				{CaptureOpen, "@("},
				{Text, "b"},
				{CaptureOpen, "+("},
				{Text, "c"},
				{CaptureClose, ")"},
				{Text, "d"},
				{Separator, "|"},
				{Text, "e"},
				{CaptureOpen, "*("},
				{Text, "f"},
				{CaptureClose, ")"},
				{Text, "g"},
				{Single, "?"},
				{Separator, "|"},
				{CaptureOpen, "?("},
				{Text, "h"},
				{CaptureClose, ")"},
				{Text, "i"},
				{CaptureOpen, "@("},
				{Text, "j"},
				{Separator, "|"},
				{Text, "k"},
				{CaptureClose, ")"},
				{CaptureClose, ")"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\,world",
			items: []Token{
				{Text, "hello,world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello\\{world",
			items: []Token{
				{Text, "hello{world"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello?",
			items: []Token{
				{Text, "hello"},
				{Single, "?"},
				{EOF, ""},
			},
		},
		{
			pattern: "hellof*",
			items: []Token{
				{Text, "hellof"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "hello**",
			items: []Token{
				{Text, "hello"},
				{Super, "**"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日-語]",
			items: []Token{
				{RangeOpen, "["},
				{Text, "日-語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[[:alpha:]]",
			items: []Token{
				{RangeOpen, "["},
				{Text, "[:alpha:]"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[![:alpha:]]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "[:alpha:]"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日-語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "日-語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "[!日本語]",
			items: []Token{
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "日本語"},
				{RangeClose, "]"},
				{EOF, ""},
			},
		},
		{
			pattern: "{a,b}",
			items: []Token{
				{TermsOpen, "{"},
				{Text, "a"},
				{Separator, ","},
				{Text, "b"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{z,ab}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "z"},
				{Separator, ","},
				{Text, "ab"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "+(a|b\\[)*",
			items: []Token{
				{CaptureOpen, "+("},
				{Text, "a"},
				{Separator, "|"},
				{Text, "b["},
				{CaptureClose, ")"},
				{Any, "*"},
				{EOF, ""},
			},
		}, {
			pattern: "ab**(e|f)",
			items: []Token{
				{Text, "ab"},
				{Any, "*"},
				{CaptureOpen, "*("},
				{Text, "e"},
				{Separator, "|"},
				{Text, "f"},
				{CaptureClose, ")"},
				{EOF, ""},
			},
		},
		{
			pattern: "/{z,ab,!([a-z]*|?)}*",
			items: []Token{
				{Text, "/"},
				{TermsOpen, "{"},
				{Text, "z"},
				{Separator, ","},
				{Text, "ab"},
				{Separator, ","},
				{CaptureOpen, "!("},
				{RangeOpen, "["},
				{Text, "a-z"},
				{RangeClose, "]"},
				{Any, "*"},
				{Separator, "|"},
				{Single, "?"},
				{CaptureClose, ")"},
				{TermsClose, "}"},
				{Any, "*"},
				{EOF, ""},
			},
		},
		{
			pattern: "{[!日-語],*,?,{a,b,\\c}}",
			items: []Token{
				{TermsOpen, "{"},
				{RangeOpen, "["},
				{Not, "!"},
				{Text, "日-語"},
				{RangeClose, "]"},
				{Separator, ","},
				{Any, "*"},
				{Separator, ","},
				{Single, "?"},
				{Separator, ","},
				{TermsOpen, "{"},
				{Text, "a"},
				{Separator, ","},
				{Text, "b"},
				{Separator, ","},
				{Text, "c"},
				{TermsClose, "}"},
				{TermsClose, "}"},
				{EOF, ""},
			},
		},
	} {
		lexer := NewLexer(test.pattern)
		for i, exp := range test.items {
			act := lexer.Next()
			if act.Type != exp.Type {
				t.Errorf("#%d %q: wrong %d-th item type: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Type, act.Type, exp, act)
			}
			if act.Raw != exp.Raw {
				t.Errorf("#%d %q: wrong %d-th item contents: exp: %q; act: %q\n\t(%s vs %s)", id, test.pattern, i, exp.Raw, act.Raw, exp, act)
			}
		}
	}
}
