package starcmp

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.starlark.net/starlark"
)

func TestCompare(t *testing.T) {
	testData := []struct {
		name      string
		a, b      starlark.Value
		wantEqual bool
	}{
		{
			name:      "not equal strings",
			a:         starlark.String("a"),
			b:         starlark.String("b"),
			wantEqual: false,
		},
		{
			name:      "equal strings",
			a:         starlark.String("a"),
			b:         starlark.String("a"),
			wantEqual: true,
		},
		{
			name:      "not equal numbers",
			a:         starlark.MakeInt(42),
			b:         starlark.MakeInt(43),
			wantEqual: false,
		},
		{
			name:      "equal numbers",
			a:         starlark.MakeInt(42),
			b:         starlark.MakeInt(42),
			wantEqual: true,
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			diff := cmp.Diff(test.a, test.b, Compare)
			if got, want := (diff == ""), test.wantEqual; got != want {
				if diff != "" {
					t.Errorf("values differ, but want same:\n%s", diff)
				} else {
					t.Error("values unexpectedly equal")
				}
			}
		})
	}
}
