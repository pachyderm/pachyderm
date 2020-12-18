package testutil

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

// TestPipelineBuildLifecycle is a parameterizable test for build pipelines,
// used by both pps cli and auth
func TestPipelineBuildLifecycle(t *testing.T, lang, dir string, directoryDepth int) string {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	prefix := strings.Repeat("../", directoryDepth)

	// reset and create some test input
	require.NoError(t, BashCmd(`
		yes | pachctl delete all
		pachctl create repo in
		pachctl put file -r in@master:/ -f {{.prefix}}etc/testing/pipeline-build/input
	`, "prefix", prefix).Run())

	// give pipeline a unique name to work around
	// github.com/kubernetes/kubernetes/issues/82130
	pipeline := UniqueString("test-pipeline-build")
	spec := fmt.Sprintf(`
		{
		  "pipeline": {
		    "name": %q
		  },
		  "transform": {
		    "build": {
		      "language": %q
		    }
		  },
		  "input": {
		    "pfs": {
		      "repo": "in",
		      "glob": "/*"
		    }
		  }
		}
	`, pipeline, lang)

	// test a barebones pipeline with a build spec and verify results
	require.NoError(t, BashCmd(`
		cd {{.prefix}}/etc/testing/pipeline-build/{{.dir}}
		pachctl create pipeline <<EOF
		{{.spec}}
		EOF
		pachctl flush commit in@master
		`,
		"dir", dir,
		"spec", spec,
		"prefix", prefix,
	).Run())
	require.YesError(t, BashCmd(fmt.Sprintf(`
		pachctl list pipeline --state failure | match %s
	`, pipeline)).Run())
	verifyPipelineBuildOutput(t, pipeline, "0")

	// update the barebones pipeline and verify results
	require.NoError(t, BashCmd(`
		cd {{.prefix}}etc/testing/pipeline-build/{{.dir}}
		pachctl update pipeline <<EOF
		{{.spec}}
		EOF
		pachctl flush commit in@master
		`,
		"dir", dir,
		"spec", spec,
		"prefix", prefix,
	).Run())
	verifyPipelineBuildOutput(t, pipeline, "0")

	// update the pipeline with a custom cmd and verify results
	spec = fmt.Sprintf(`
		{
		  "pipeline": {
		    "name": %q
		  },
		  "transform": {
		    "cmd": [
		      "sh",
		      "/pfs/build/run.sh",
		      "_"
		    ],
		    "build": {
		      "language": %q,
		      "path": "."
		    }
		  },
		  "input": {
		    "pfs": {
		      "repo": "in",
		      "glob": "/*"
		    }
		  }
		}
	`, pipeline, lang)

	require.NoError(t, BashCmd(`
		cd {{.prefix}}etc/testing/pipeline-build/{{.dir}}
		pachctl update pipeline --reprocess <<EOF
		{{.spec}}
		EOF
		pachctl flush commit in@master
		`,
		"dir", dir,
		"spec", spec,
		"prefix", prefix,
	).Run())
	verifyPipelineBuildOutput(t, pipeline, "_")
	return pipeline
}

func verifyPipelineBuildOutput(t *testing.T, pipeline, prefix string) {
	t.Helper()

	require.NoError(t, BashCmd(`
		pachctl flush commit in@master
		pachctl get file {{.pipeline}}@master:/1.txt | match {{.prefix}}{{.prefix}}{{.prefix}}1
		pachctl get file {{.pipeline}}@master:/11.txt | match {{.prefix}}{{.prefix}}11
		pachctl get file {{.pipeline}}@master:/111.txt | match {{.prefix}}111
		`,
		"pipeline", pipeline,
		"prefix", prefix,
	).Run())
}
