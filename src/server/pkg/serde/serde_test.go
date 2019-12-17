package serde

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

func s(text []byte) string {
	return string(bytes.TrimSpace(text))
}

func TestJSONBasic(t *testing.T) {
	type foo struct {
		A, B string
	}
	encoded := []byte(`{"A":"first","B":"second"}`)

	var f foo
	require.NoError(t, DecodeYAML(encoded, &f))
	require.Equal(t, foo{"first", "second"}, f)
	out, err := EncodeJSON(f)
	require.NoError(t, err)
	require.Equal(t, s(encoded), s(out))
}

func TestTags(t *testing.T) {
	type foo struct {
		A string `json:"alpha"`
		B string `json:"beta"`
	}
	encoded := []byte(`{"alpha":"first","beta":"second"}`)

	var f foo
	require.NoError(t, DecodeYAML(encoded, &f))
	require.Equal(t, foo{"first", "second"}, f)
	out, err := EncodeJSON(f)
	require.NoError(t, err)
	require.Equal(t, s(encoded), s(out))
}

func TestYAMLBasic(t *testing.T) {
	type foo struct {
		A, B string
	}
	encoded := []byte("A: first\nB: second\n")

	var f foo
	require.NoError(t, DecodeYAML(encoded, &f))
	require.Equal(t, foo{"first", "second"}, f)
	out, err := EncodeYAML(f)
	require.NoError(t, err)
	require.Equal(t, s(encoded), s(out))
}

func TestJSONTagsWhenDecodingYAML(t *testing.T) {
	type foo struct {
		A string `json:"alpha"`
		B string `json:"beta"`
	}
	encoded := []byte("alpha: first\nbeta: second\n")

	var f foo
	require.NoError(t, DecodeYAML(encoded, &f))
	require.Equal(t, foo{"first", "second"}, f)
	out, err := EncodeYAML(f)
	require.NoError(t, err)
	require.Equal(t, s(encoded), s(out))
}

func TestDecodeYAMLTransform(t *testing.T) {
	type foo struct {
		A string
		B string
	}
	encoded := []byte("A: first\nC: third\n")

	var f foo
	d := NewYAMLDecoder(bytes.NewReader(encoded))
	require.NoError(t, d.DecodeTransform(&f, func(m map[string]interface{}) error {
		m["B"] = "second"
		delete(m, "C")
		return nil
	}))
	require.Equal(t, foo{"first", "second"}, f)

	var buf bytes.Buffer
	e := NewYAMLEncoder(&buf)
	require.NoError(t, e.EncodeTransform(f, func(m map[string]interface{}) error {
		m["C"] = "third"
		delete(m, "B")
		return nil
	}))
	require.Equal(t, s(encoded), strings.TrimSpace(buf.String()))
}

func TestEncodeOptions(t *testing.T) {
	type foo struct {
		A string
		B string
	}
	data := struct {
		F foo
	}{
		F: foo{"first", "second"},
	}

	var buf bytes.Buffer
	e := NewYAMLEncoder(&buf, WithIndent(3)) // unusual indent, to demo option
	require.NoError(t, e.Encode(data))
	expected := []byte("F:\n   A: first\n   B: second\n")
	// don't trim space to ensure leading space is the same
	require.Equal(t, string(expected), buf.String())
}

// TODO(msteffen) add proto tests

// TODO(msteffen) add proto tests
