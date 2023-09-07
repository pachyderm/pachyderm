package starlark

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.starlark.net/starlark"
)

func TestLoad(t *testing.T) {
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
			origModules := Modules
			t.Cleanup(func() { Modules = origModules })
			var testOK bool
			Modules["test"] = starlark.StringDict{
				"test_ok": starlark.NewBuiltin("test_ok", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
					t.Log("test.test_ok called")
					testOK = true
					return starlark.None, nil
				}),
			}

			ctx := pctx.TestContext(t)
			_, err := RunProgram(ctx, "testdata/"+test.file, Options{})
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
