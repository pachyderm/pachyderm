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

func TestCreatePipeline_defaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create defaults --cluster <<EOF
		{
			"create_pipeline_request": {
				"datumTries": 17,
				"autoscaling": true
			}
		}
		EOF
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.memory | match null
		# the raw format currently marshals to snake case
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.datum_tries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.autoscaling | match null
		# the specs marshal to camelCase
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .autoscaling | match false
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
	).Run())
}

func TestCreatePipeline_regenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create defaults --cluster <<EOF
		{
			"create_pipeline_request": {
				"datumTries": 17,
				"autoscaling": true
			}
		}
		EOF
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.memory | match null
		# the raw format currently marshals to snake case
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.datum_tries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.autoscaling | match null
		# the specs marshal to camelCase
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .autoscaling | match false
		# updating defaults without --regenerate does not change the pipeline
		pachctl create defaults --cluster <<EOF
		{
			"create_pipeline_request": {
				"datumTries": 18,
				"autoscaling": true
			}
		}
		EOF
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 17
		# updating defaults with --regenerate does change the pipeline
		pachctl create defaults --cluster --regenerate <<EOF
		{
			"create_pipeline_request": {
				"datumTries": 8,
				"autoscaling": true
			}
		}
		EOF
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 8
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline").Run())
}

func TestCreatePipeline_delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create defaults --cluster <<EOF
		{
			"create_pipeline_request": {
				"datumTries": 17,
				"autoscaling": true
			}
		}
		EOF
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.memory | match null
		# the raw format currently marshals to snake case
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.datum_tries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.autoscaling | match null
		# the specs marshal to camelCase
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 17
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .autoscaling | match false
		# deleting defaults without --regenerate does not change the pipeline
		pachctl delete defaults --cluster
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 17
		# delete defaults with --regenerate does change the pipeline
		pachctl delete defaults --cluster --regenerate
		EOF
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match null
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline").Run())
}
