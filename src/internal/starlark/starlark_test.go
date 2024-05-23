package starlark

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/starlark/starcmp"
	"go.starlark.net/starlark"
)

func TestRunProgram(t *testing.T) {
	testData := []struct {
		file    string
		wantErr bool
	}{
		{file: "good.star"},
		{file: "nested/good.star"},
		{file: "cycle.star", wantErr: true},
	}

	for _, test := range testData {
		t.Run(test.file, func(t *testing.T) {
			var testOK bool
			opts := Options{
				Modules: map[string]starlark.StringDict{
					"test": {
						"test_ok": starlark.NewBuiltin("test_ok", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
							t.Log("test.test_ok called")
							testOK = true
							return starlark.None, nil
						}),
					},
				},
			}
			ctx := pctx.TestContext(t)
			_, err := RunProgram(ctx, "testdata/"+test.file, opts)
			if err != nil {
				t.Log(err)
				if test.wantErr {
					return
				}
				t.Fatal(err)
			}
			if test.wantErr {
				t.Error("want error, but got success")
			}
			if !testOK {
				t.Errorf("test never called test_ok()")
			}
		})
	}
}

func TestRunScript(t *testing.T) {
	ctx := pctx.TestContext(t)
	list := ReflectList([]int{42: 1234})
	got, err := RunScript(ctx, "test", "y = x[42] + 1", Options{Predefined: starlark.StringDict{"x": list}})
	if err != nil {
		t.Fatalf("RunScript: %v", err)
	}
	want := starlark.StringDict{"y": starlark.MakeInt(1235)}
	if diff := cmp.Diff(want, got, starcmp.Compare); diff != "" {
		t.Errorf("RunScript (-want +got):\n%s", diff)
	}
}
