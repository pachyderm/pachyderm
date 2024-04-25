package kvparse

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmputil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestParseOne(t *testing.T) {
	testData := []struct {
		input   string
		want    KV
		wantErr string
	}{
		{
			input:   "",
			wantErr: `/unexpected token "<EOF>"/`,
		},
		{
			input: "key=value",
			want:  KV{Key: "key", Value: "value"},
		},
		{
			input: `"key with space"="value with space"`,
			want:  KV{Key: "key with space", Value: "value with space"},
		},
		{
			input: `'key with "double quotes"'="value with 'single quotes'"`,
			want:  KV{Key: `key with "double quotes"`, Value: "value with 'single quotes'"},
		},
		{
			input: `";"=';'`,
			want:  KV{Key: `;`, Value: ";"},
		},
		{
			input: `"="="="`,
			want:  KV{Key: "=", Value: "="},
		},
		{
			input: `"=\""="="`,
			want:  KV{Key: `="`, Value: `=`},
		},
		{
			input:   `{this is=not json}`,
			wantErr: `/invalid input text/`,
			// We could use a different lexer for single a KV pair to allow this.
		},
		{
			input: `💯=100`,
			want:  KV{Key: "💯", Value: "100"},
		},
		{
			input: `'💯'=100`,
			want:  KV{Key: "💯", Value: "100"},
		},
		{
			input: `"💯"=100`,
			want:  KV{Key: "💯", Value: "100"},
		},
		{
			input: `♁=🌎`,
			want:  KV{Key: "♁", Value: "🌎"},
		},
		{
			input:   `=`,
			wantErr: `/unexpected token "="/`,
		},
		{
			input:   `=;`,
			wantErr: `/unexpected token "="/`,
		},
		{
			input: `k=`,
			want:  KV{Key: "k"},
		},
		{
			input: `k=;`,
			want:  KV{Key: "k"},
		},
		{
			input: `"\u2641"="\x0a"`,
			want:  KV{Key: "♁", Value: "\n"},
		},
	}

	for _, test := range testData {
		t.Run(test.input, func(t *testing.T) {
			got, err := ParseOne(test.input)
			if test.wantErr != "" {
				msg := "<nil>"
				if err != nil {
					msg = err.Error()
				}
				require.NoDiff(t, test.wantErr, msg, []cmp.Option{cmputil.RegexpStrings()}, "error should match")
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			require.NoDiff(t, test.want, got, nil, "parser output should match")
		})
	}
}

func TestParseMany(t *testing.T) {
	testData := []struct {
		name    string
		input   string
		want    []KV
		wantErr string
	}{
		{
			name: "empty",
			want: nil,
		},
		{
			name:  "empty json",
			input: `{ }`,
			want:  nil,
		},
		{
			name:  "single",
			input: "key=value",
			want: []KV{
				{Key: "key", Value: "value"},
			},
		},
		{
			name:  "single terminated",
			input: "key=value;  ",
			want: []KV{
				{Key: "key", Value: "value"},
			},
		},
		{
			name:    "not json",
			input:   `{"this is not:"="json"}`,
			wantErr: `/invalid input/`,
		},
		{
			name:  "simple json",
			input: `{"key":"value"}`,
			want: []KV{
				{Key: "key", Value: "value"},
			},
		},
		{
			name:  "simple json with spaces",
			input: `{	"key"  :	"value" }`,
			want: []KV{
				{Key: "key", Value: "value"},
			},
		},
		{
			name:  "simple json with trailing comma",
			input: `{"key":"value",}`,
			want: []KV{
				{Key: "key", Value: "value"},
			},
		},
		{
			name:  "complex json",
			input: `{"key":"value","key2":"value2",";":"=","'":"\"",":":"}"}`,
			want: []KV{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "value2"},
				{Key: ";", Value: "="},
				{Key: "'", Value: `"`},
				{Key: ":", Value: `}`},
			},
		},
		{
			name: "json with trailing comma and spaces",
			input: `{ "key":   "value"  ,   "key2"	:"value2"
				,";":"=",
				"'":"\"",":": "}",
			}`,
			want: []KV{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "value2"},
				{Key: ";", Value: "="},
				{Key: "'", Value: `"`},
				{Key: ":", Value: `}`},
			},
		},
		{
			name: "doc example",
			input: `key=value;key2=value2;"key with space"="value with space";
'key with "double quotes"'="value with 'single quotes'";
'";"'="';'";"="="="`,
			want: []KV{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "value2"},
				{Key: "key with space", Value: "value with space"},
				{Key: `key with "double quotes"`, Value: "value with 'single quotes'"},
				{Key: `";"`, Value: "';'"},
				{Key: "=", Value: "="},
			},
		},
		{
			name:    "0",
			input:   `0`,
			wantErr: `/must match at least once/`,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseMany(test.input)
			if test.wantErr != "" {
				msg := "<nil>"
				if err != nil {
					msg = err.Error()
				}
				require.NoDiff(t, test.wantErr, msg, []cmp.Option{cmputil.RegexpStrings()}, "error should match")
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			require.NoDiff(t, test.want, got, nil, "parser output should match")
		})
	}
}

func FuzzParseOne(f *testing.F) {
	f.Add("key=value")
	f.Add("key=value;")
	f.Add(`"key"="value"`)
	f.Add(`'key'='value'`)
	f.Fuzz(func(t *testing.T, in string) {
		ParseOne(in) //nolint:errcheck
	})
}

func FuzzParseMany(f *testing.F) {
	f.Add("key=value")
	f.Add("key=value;key2=value2")
	f.Add(`"key"="value"`)
	f.Add(`'key'='value'`)
	f.Add(`{"key":"value"}`)
	f.Add(`{"key":"value","key2": "value2"}`)
	f.Fuzz(func(t *testing.T, in string) {
		ParseMany(in) //nolint:errcheck
	})
}
