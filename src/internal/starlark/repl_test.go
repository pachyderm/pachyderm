package starlark

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

func TestCompletion(t *testing.T) {
	testData := []struct {
		line string
		want []string
		skip bool
	}{
		{
			line: "Fal",
			want: []string{"se"},
		},
		{
			line: "som",
			want: []string{"e_dict", "efoo", "e123"},
		},
		{
			line: "somef",
			want: []string{"oo"},
		},
		{
			line: "\tsomef",
			want: []string{"oo"},
		},
		{
			line: "    somef",
			want: []string{"oo"},
		},
		{
			line: "    some f",
		},
		{
			line: ".",
		},
		{
			line: ".some",
		},
		{
			line: "some_dict.",
			want: []string{"clear", "get", "items", "keys", "pop", "popitem", "setdefault", "update", "values"},
		},
		{
			line: "some_dict.cl",
			want: []string{"ear"},
		},
		{
			line: "some_dict.some_dict.",
		},
		{
			line: "some_dict.Fal",
		},
		{
			line: "exciting_function(Fal",
			want: []string{"se"},
		},
		{
			line: "exciting_function( Fal",
			want: []string{"se"},
		},
		{
			line: "exciting_function(some_dict.cl",
			want: []string{"ear"},
		},
		{
			line: "exciting_function(arg=Fal",
			want: []string{"se"},
		},
		{
			line: "exciting_function(arg=some_dict.cl",
			want: []string{"ear"},
		},
		{
			line: "[None, True, Fal",
			want: []string{"se"},
		},
		{
			line: "nested[0].cl",
			want: []string{"ear"},
			// To complete .clear here, we'd need to eval "nested[0]".  It would be a
			// nice feature, though.
			skip: true,
		},
	}

	for _, test := range testData {
		t.Run(test.line, func(t *testing.T) {
			if test.skip {
				t.Skip("skipped")
			}
			ctx := pctx.TestContext(t)
			var th *starlark.Thread
			var g starlark.StringDict
			if _, err := Run(ctx, "testdata/completion.star", Options{}, func(fileOpts *syntax.FileOptions, thread *starlark.Thread, path string, module string, globals starlark.StringDict) (starlark.StringDict, error) {
				th = thread
				g = globals
				r, err := starlark.ExecFileOptions(fileOpts, thread, module, nil, globals)
				for k, v := range r {
					g[k] = v
				}
				return r, errors.Wrap(err, "ExecFileOptions")
			}); err != nil {
				t.Fatalf("compile example file: %v", err)
			}
			c := &completer{t: th, globals: g}
			rs := []rune(test.line)
			result, _ := c.Do(rs, len(rs))
			var got []string
			for _, x := range result {
				got = append(got, string(x))
			}
			sort := func(x, y string) bool {
				return x < y
			}
			if diff := cmp.Diff(test.want, got, cmpopts.EquateEmpty(), cmpopts.SortSlices(sort)); diff != "" {
				t.Errorf("complete(%v) (-want +got):\n%s", test.line, diff)
			}
		})
	}
}
