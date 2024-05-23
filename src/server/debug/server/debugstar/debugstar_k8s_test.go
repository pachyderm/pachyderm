//go:build k8s

package debugstar

import (
	"testing"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestBuiltinScripts(t *testing.T) {
	var ran bool
	for name, script := range BuiltinScripts {
		t.Run(name, func(t *testing.T) {
			ctx := pctx.TestContext(t)
			env := &Env{FS: testDumpFS(fstest.MapFS{})}
			if err := env.RunStarlark(ctx, script, string(script)); err != nil {
				t.Errorf("run script: %v", err)
			}
			ran = true
		})
	}
	if !ran {
		t.Fatal("expected to run at least 1 built-in script")
	}
}
