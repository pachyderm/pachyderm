//go:build k8s

package cmds_test

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestCreatePipeline_noDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	cluster, _ := minikubetestenv.AcquireCluster(t)
	ctx := pctx.TestContext(t)
	p, err := testutil.NewPachctl(ctx, cluster, fmt.Sprintf("%s/test-pach-config-%s.json", t.TempDir(), t.Name()))
	require.NoError(t, err, "must create Pachyderm client")
	output, err := p.RunCommand(ctx, "pachctl create repo input")
	require.NoError(t, err, "must create input repo: %v", output)
	output, err = p.RunCommand(ctx, `pachctl create pipeline <<EOF
{
	"pipeline": {
		"project": {
			"name": "default"
		},
		"name": "pipeline"
	},
	"transform": {
		"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
	},
	"input": {
		"pfs": {
			"project": "default",
			"repo": "input",
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
EOF`)
	require.NoError(t, err, "must create pipeline: %v", output)
	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.resource_requests.disk`)
	require.NoError(t, err, "must extract disk request: %v", output)
	require.Equal(t, "188Mi", output, "disk request must be overridden")
	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.resource_requests.memory`)
	require.NoError(t, err, "must extract memory request: %v", output)
	require.Equal(t, "64M", output, "memory request must be defaulted")
}

func TestCreatePipeline_defaults(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	cluster, _ := minikubetestenv.AcquireCluster(t)
	ctx := pctx.TestContext(t)

	p, err := testutil.NewPachctl(ctx, cluster, fmt.Sprintf("%s/test-pach-config-%s.json", t.TempDir(), t.Name()))
	require.NoError(t, err, "must create Pachyderm client")

	output, err := p.RunCommand(ctx, `pachctl create defaults --cluster <<EOF
{
	"create_pipeline_request": {
		"datumTries": 17,
		"autoscaling": true
	}
}
EOF`)
	require.NoError(t, err, "must create cluster defaults: %v", output)

	output, err = p.RunCommand(ctx, `pachctl create repo input`)
	require.NoError(t, err, "must create input repo: %v", output)

	output, err = p.RunCommand(ctx, `pachctl create pipeline <<EOF
{
	"pipeline": {
		"name": "pipeline"
	},
	"transform": {
		"cmd": ["cp", "r", "/pfs/in", "/pfs/out"]
	},
	"input": {
		"pfs": {
			"project": "default",
			"repo": "input",
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
EOF`)
	require.NoError(t, err, "must create pipeline: %v", output)

	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.resource_requests.disk`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "187Mi", output, "must override disk request")

	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.resource_requests.memory`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "null", output, "must have empty memory request")

	// the raw format currently marshals to snake_case
	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.datum_tries`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "17", output, "must default datum tries")

	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .details.autoscaling`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "null", output, "must override autoscaling")

	// the specs marshal to camelCase
	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .effective_spec_json | jq -j .datumTries`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "17", output, "must default datum tries")

	output, err = p.RunCommand(ctx, `pachctl inspect pipeline pipeline --raw | jq -j .effective_spec_json | jq -j .autoscaling`)
	require.NoError(t, err, "must inspect pipeline: %v", output)
	require.Equal(t, "false", output, "must override autoscaling")
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .details.resource_requests.cpu | match null
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 3
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline").Run())
}

func TestCreatePipeline_yaml(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, tu.PachctlBashCmd(t, c, `
		pachctl create defaults --cluster <<EOF
		create_pipeline_request:
		  datumTries: 17
		  autoscaling: true
		EOF
		pachctl create repo {{.RepoName}}
		pachctl create pipeline <<EOF
		pipeline:
		  project:
		    name: "{{.ProjectName | js}}"
		  name: "{{.PipelineName | js}}"
		transform:
		  cmd:
		  - cp
		  - r
		  - /pfs/in
		  - /pfs/out
		input:
		  pfs:
		    project: default
		    repo: "{{.RepoName | js}}"
		    glob: "/*"
		    name: in
		resource_requests:
		  cpu: null
		  disk: "187Mi"
		autoscaling: false
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .datumTries | match 3
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline").Run())
}

func TestCreatePipeline_leading_zero(t *testing.T) {
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
				"cpu": 0,
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
		pachctl inspect pipeline {{.PipelineName}} --raw | jq -r .effective_spec_json | jq -r .resourceRequests.cpu | match 0
		`,
		"ProjectName", pfs.DefaultProjectName,
		"RepoName", "input",
		"PipelineName", "pipeline",
	).Run())
}
