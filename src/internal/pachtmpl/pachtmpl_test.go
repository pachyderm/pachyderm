package pachtmpl

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestEval(t *testing.T) {
	fsCtx := map[string][]byte{
		"entry.jsonnet": []byte(`
						local myFunction = import "myFunction.jsonnet";
						local mattsInternetFunction = import "https://raw.githubusercontent.com/pachyderm/jsonnet-pipelines-test/main/edges.jsonnet" ;
						{
							a: myFunction(1),
							b: mattsInternetFunction("suffix", "src"),
						}
				`),
		"myFunction.jsonnet": []byte(`function (arg1) std.toString(arg1)`),
	}
	output, err := Eval(fsCtx, "entry.jsonnet")
	require.NoError(t, err)
	t.Log(string(output))
}
