//go:build k8s

package cmds_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func TestCreatePipeline_noDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create repo {{.RepoName}}
		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"project": {
					"name": "{{.ProjectName | js}}"
				},
				"name": "{{.PipelineName | js}}"
			},
			"transform": {
				"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
			},
			"input": {
				"pfs": {
					"project": "default",
					"repo": "{{.RepoName | js}}",
					"glob": "/*",
					"name": "in"
				}
			},
			"resource_requests": {
				"cpu": null,
				"disk": "187Mi"
			},
			"autoscaling": false
		}
		EOF
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.disk | match 187Mi
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.memory | match 64M
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
	).Run())
}
