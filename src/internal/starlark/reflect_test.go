package starlark_test

import (
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	ourstar "github.com/pachyderm/pachyderm/v2/src/internal/starlark"
	"github.com/pachyderm/pachyderm/v2/src/internal/starlark/startest"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type SimpleStruct struct {
	Field string
	Me    *SimpleStruct
}

func (s *SimpleStruct) String() string {
	return "simple struct" // test fmt.Stringer passthru
}

func TestValue(t *testing.T) {
	simple := SimpleStruct{
		Field: "simple",
	}
	simple.Me = &simple

	predefined := starlark.StringDict{
		"wrappednil": ourstar.Value(&ourstar.Reflect{}),
		"struct":     ourstar.Value(simple),
		"struct_ptr": ourstar.Value(&simple),
		"string":     ourstar.Value("string"),
		"bool":       ourstar.Value(bool(true)),
		"byte":       ourstar.Value([]byte("bytes")),
		"int":        ourstar.Value(42),
		"float":      ourstar.Value(123.45),
		"stringlist": ourstar.Value([]starlark.Value{starlark.String("string"), starlark.String("string")}),
		"map":        ourstar.Value(map[string]int{"one": 1, "two": 2}),
		"array":      ourstar.Value([10]int{5: 5}),
		"int8":       ourstar.Value(int8(-128)),
		"uint8":      ourstar.Value(uint8(255)),
	}
	startest.RunTest(t, "tests/reflect_test.star", ourstar.Options{Predefined: predefined})
}

func TestRoundTrip(t *testing.T) {
	want := map[string]any{
		"nil":        nil,
		"struct":     SimpleStruct{Field: "simple"},
		"struct_ptr": &SimpleStruct{Field: "pointer"},
		"string":     "string",
		"byte":       []byte("bytes"),
		"int64":      int64(math.MinInt64),
		"unt64":      uint64(math.MaxInt64 + 1),
		"float":      float64(123.45),
		"list":       []any{"string", "string"},
		"complex":    complex(2, 1),
		"dict":       map[string]any{"one": int64(1), "two": int64(2)},
	}
	th := new(starlark.Thread)
	result, err := starlark.EvalOptions(&syntax.FileOptions{}, th, "<test>", "lambda x: x", starlark.StringDict{})
	if err != nil {
		t.Fatal(err)
	}
	f := result.(*starlark.Function)
	got := map[string]any{}
	for k, v := range want {
		out, err := starlark.Call(th, f, starlark.Tuple([]starlark.Value{ourstar.Value(v)}), nil)
		if err != nil {
			t.Fatalf("Call(%v): %v", k, err)
		}
		got[k] = ourstar.FromStarlark(out)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("round trip (-want +got):\n%s", diff)
	}
}
