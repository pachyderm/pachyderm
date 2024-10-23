//go:build k8s

// TODO(msteffen) Add tests for:
//
// - restart datum
// - stop job
// - delete job
//
// - inspect job
// - list job
//
// - create pipeline
// - create pipeline --push-images (re-enable existing test)
// - update pipeline
// - delete pipeline
//
// - inspect pipeline
// - list pipeline
//
// - start pipeline
// - stop pipeline
//
// - list datum
// - inspect datum
// - logs

package cmds

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestRawFullPipelineInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create project my_project
		pachctl create repo data --project my_project
		pachctl put file data@master:/file <<<"This is a test" --project my_project
		pachctl create pipeline --project my_project <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
		        "repo": "data"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["cp /pfs/data/file /pfs/out"]
		    }
		  }
		EOF
		`,
		"pipeline", tu.UniqueString("p-")).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl wait commit data@master --project my_project

		# make sure the results have the full pipeline info, including version
		pachctl list job --raw --project my_project \
			| match "pipeline_version"
		`).Run())
}

func TestUnrunnableJobInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	ctx := pctx.TestContext(t)
	require.NoErrorWithinTRetry(t, 2*time.Minute, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		yes | pachctl delete all
	`).Run)
	pipeline1 := tu.UniqueString("p-")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo data
		pachctl put file data@master:/file <<<"This is a test"
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
		        "repo": "data"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["exit 1"]
		    }
		  }
		EOF
		`,
		"pipeline", pipeline1).Run())
	pipeline2 := tu.UniqueString("p-")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
		        "repo": "{{.inputPipeline}}"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["exit 0"]
		    }
		  }
		EOF
		`,
		"pipeline", pipeline2, "inputPipeline", pipeline1).Run())
	require.NoErrorWithinTRetryConstant(t, 60*time.Second, func() error {
		return errors.EnsureStack(testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl wait commit data@master

		# make sure that there is a not-run job
		pachctl list job --raw \
			| match "JOB_UNRUNNABLE"
		# make sure the results have the full pipeline info, including version
		pachctl list pipeline \
			| match "unrunnable"
		`, "pipeline", pipeline2).Run())
	}, 1*time.Second)
}

// TestJSONMultiplePipelines tests that pipeline specs with multiple pipelines
// are accepted, as long as they're separated by a YAML document separator
func TestJSONMultiplePipelines(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
		pachctl create repo input
		pachctl create pipeline -f - <<EOF
		{
		  "pipeline": {
		    "name": "first"
		  },
		  "input": {
		    "pfs": {
		      "glob": "/*",
		      "repo": "input"
		    }
		  },
		  "transform": {
		    "cmd": [ "/bin/bash" ],
		    "stdin": [
		      "cp /pfs/input/* /pfs/out"
		    ]
		  }
		}
		---
		{
		  "pipeline": {
		    "name": "second"
		  },
		  "input": {
		    "pfs": {
		      "glob": "/*",
		      "repo": "first"
		    }
		  },
		  "transform": {
		    "cmd": [ "/bin/bash" ],
		    "stdin": [
		      "cp /pfs/first/* /pfs/out"
		    ]
		  }
		}
		EOF

		pachctl start commit input@master
		echo foo | pachctl put file input@master:/foo
		echo bar | pachctl put file input@master:/bar
		echo baz | pachctl put file input@master:/baz
		pachctl finish commit input@master
		pachctl wait commit second@master
		pachctl get file second@master:/foo | match foo
		pachctl get file second@master:/bar | match bar
		pachctl get file second@master:/baz | match baz
		`,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `pachctl list pipeline`).Run())
}

// TestJSONStringifiedNumberstests that JSON pipelines may use strings to
// specify numeric values such as a pipeline's parallelism (a feature of gogo's
// JSON parser).
func TestJSONStringifiedNumbers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
		pachctl create repo input
		pachctl create pipeline -f - <<EOF
		{
		  "pipeline": {
		    "name": "first"
		  },
		  "input": {
		    "pfs": {
		      "glob": "/*",
		      "repo": "input"
		    }
		  },
		  "parallelism_spec": {
		    "constant": "1"
		  },
		  "transform": {
		    "cmd": [ "/bin/bash" ],
		    "stdin": [
		      "cp /pfs/input/* /pfs/out"
		    ]
		  }
		}
		EOF

		pachctl start commit input@master
		echo foo | pachctl put file input@master:/foo
		echo bar | pachctl put file input@master:/bar
		echo baz | pachctl put file input@master:/baz
		pachctl finish commit input@master
		pachctl wait commit first@master
		pachctl get file first@master:/foo | match foo
		pachctl get file first@master:/bar | match bar
		pachctl get file first@master:/baz | match baz
		`,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `pachctl list pipeline`).Run())
}

// TestYAMLPipelineSpec tests creating a pipeline with a YAML pipeline spec
func TestYAMLPipelineSpec(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	// Note that BashCmd dedents all lines below including the YAML (which
	// wouldn't parse otherwise)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
		pachctl create project P
		pachctl create repo input --project P
		pachctl create pipeline -f - <<EOF
		pipeline:
		  name: first
		  project:
		    name: P
		input:
		  pfs:
		    glob: /*
		    repo: input
		    project: P
		transform:
		  cmd: [ /bin/bash ]
		  stdin:
		    - "cp /pfs/input/* /pfs/out"
		---
		pipeline:
		  name: second
		input:
		  pfs:
		    glob: /*
		    repo: first
		    project: P
		transform:
		  cmd: [ /bin/bash ]
		  stdin:
		    - "cp /pfs/first/* /pfs/out"
		EOF
		pachctl start commit input@master --project P
		echo foo | pachctl put file input@master:/foo --project P
		echo bar | pachctl put file input@master:/bar --project P
		echo baz | pachctl put file input@master:/baz --project P
		pachctl finish commit input@master --project P
		pachctl wait commit second@master
		pachctl get file second@master:/foo | match foo
		pachctl get file second@master:/bar | match bar
		pachctl get file second@master:/baz | match baz
		`,
	).Run())
}

func TestListPipelineFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())
	pipeline1, pipeline2 := tu.UniqueString("pipeline1-"), tu.UniqueString("pipeline2-")
	out, err := testutilpachctl.PachctlBashCmdCtx(pctx.TestContext(t), t, c, `
		yes | pachctl delete all
		pachctl create project myProject
		pachctl create repo input
		pachctl create pipeline -f - <<EOF
		{
			"pipeline": {
				"name": "{{.pipeline1}}"
			},
			"input": {
				"pfs": {
					"glob": "/*",
					"repo": "input"
				}
			},
			"parallelism_spec": {
				"constant": "1"
			},
			"transform": {
				"cmd": [
					"/bin/bash"
				],
				"stdin": [
					"cp /pfs/input/* /pfs/out"
				]
			}
		}
		EOF

		pachctl create pipeline -f - <<EOF
		{
			"pipeline": {
				"name": "{{.pipeline2}}",
				"project": {
					"name": "myProject"
				}
			},
			"input": {
				"pfs": {
					"name": "in",
					"glob": "/*",
					"repo": "{{.pipeline1}}",
					"project": "default"
				}
			},
			"parallelism_spec": {
				"constant": "1"
			},
			"transform": {
				"cmd": [
					"/bin/bash"
				],
				"stdin": [
					"cp /pfs/in/* /pfs/out"
				]
			}
		}
		EOF

		echo foo | pachctl put file input@master:/foo
		pachctl wait commit {{.pipeline1}}@master
		# make sure we see the pipeline with the appropriate state filters
		pachctl list pipeline | match {{.pipeline1}}
		pachctl list pipeline | match -v {{.pipeline2}}
		pachctl list pipeline --project myProject | match {{.pipeline2}}
		pachctl list pipeline --all-projects | match {{.pipeline1}}
		pachctl list pipeline --all-projects | match {{.pipeline2}}
		pachctl list pipeline --state crashing --state failure | match -v {{.pipeline1}}
		pachctl list pipeline --state starting --state running --state standby | match {{.pipeline1}}
	`,
		"pipeline1", pipeline1,
		"pipeline2", pipeline2,
	).Output()
	require.NoError(t, err, "failed pachctl command with output %v", string(out))
}

func TestInspectWaitJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())
	pipeline1 := tu.UniqueString("pipeline1-")
	project := tu.UniqueString("myNewProject")

	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create project {{ .project }}
		pachctl create repo input --project {{ .project }}
		pachctl create pipeline -f - <<EOF
		{
			"pipeline": {
				"name": "{{.pipeline1}}",
				"project": {
					"name": "{{ .project }}"
				}
			},
			"input": {
				"pfs": {
					"glob": "/*",
					"project": "{{ .project }}",
					"repo": "input"
				}
			},
			"parallelism_spec": {
				"constant": "1"
			},
			"transform": {
				"cmd": [
					"/bin/bash"
				],
				"stdin": [
					"sleep 15 && cp /pfs/input/* /pfs/out"
				]
			}
		}
		EOF
		echo foo | pachctl put file input@master:/foo --project {{ .project }}
	`,
		"pipeline1", pipeline1,
		"project", project,
	).Run())
	jobs, err := c.ListJob(project, pipeline1, nil, -1, false)
	require.NoError(t, err)
	require.Equal(t, 2, len(jobs))

	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl inspect job {{ .pipeline1 }}@{{ .job }} --project {{ .project }} | match {{ .job }}
		pachctl wait job {{ .pipeline1 }}@{{ .job }} --project {{ .project }}
		pachctl inspect job {{ .pipeline1 }}@{{ .job }} --project {{ .project }} --no-color | match "State: success"
	`,
		"pipeline1", pipeline1,
		"project", project,
		"job", jobs[0].Job.GetId(),
	).Run())
}

// TestYAMLSecret tests creating a YAML pipeline with a secret (i.e. the fix for
// https://github.com/pachyderm/pachyderm/v2/issues/4119)
func TestYAMLSecret(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, ns := minikubetestenv.AcquireCluster(t)
	// Note that BashCmd dedents all lines below including the YAML (which
	// wouldn't parse otherwise)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all

		# kubectl get secrets >&2
		kubectl delete secrets/test-yaml-secret -n {{ .namespace }} || true
		kubectl create secret generic test-yaml-secret --from-literal=my-key=my-value -n {{ .namespace }}

		pachctl create repo input
		pachctl put file input@master:/foo <<<"foo"
		pachctl create pipeline -f - <<EOF
		  pipeline:
		    name: pipeline
		  input:
		    pfs:
		      glob: /*
		      repo: input
		  transform:
		    cmd: [ /bin/bash ]
		    stdin:
		      - "env | grep MY_SECRET >/pfs/out/vars"
		    secrets:
		      - name: test-yaml-secret
		        env_var: MY_SECRET
		        key: my-key
		EOF
		pachctl wait commit pipeline@master
		pachctl get file pipeline@master:/vars | match MY_SECRET=my-value
		`, "namespace", ns,
	).Run())
}

func TestPipelineCrashingRecovers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())

	// prevent new pods from being scheduled
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		kubectl cordon minikube
	`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data
		pachctl create pipeline <<EOF
		  pipeline:
		    name: my-pipeline
		  input:
		    pfs:
		      glob: /*
		      repo: data
		  transform:
		    cmd: [ /bin/bash ]
		    stdin:
		      - "cp /pfs/data/* /pfs/out"
		EOF
		`).Run())

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		return errors.EnsureStack(testutilpachctl.PachctlBashCmd(t, c, `
		pachctl list pipeline \
		| match my-pipeline \
		| match crashing
		`).Run())
	})

	// allow new pods to be scheduled
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		kubectl uncordon minikube
	`).Run())

	require.NoErrorWithinTRetry(t, 30*time.Second, func() error {
		return errors.EnsureStack(testutilpachctl.PachctlBashCmd(t, c, `
		pachctl list pipeline \
		| match my-pipeline \
		| match running
		`).Run())
	})
}

func TestJsonnetPipelineTemplate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data
		pachctl put file data@master:/foo <<<"foo-data"
		pachctl create pipeline --jsonnet - --arg name=foo --arg output=bar <<EOF
		function(name, output) {
		  pipeline: { name: name+"-pipeline" },
		  input: {
		    pfs: {
		      name: "input",
		      glob: "/*",
		      repo: "data"
		    }
		  },
		  transform: {
		    cmd: [ "/bin/bash" ],
		    stdin: [ "cp /pfs/input/* /pfs/out/"+output ]
		  }
		}
		EOF
		pachctl list pipeline | match foo-pipeline
		pachctl wait commit foo-pipeline@master
		pachctl get file foo-pipeline@master:/bar | match foo-data
		`).Run())
}

func TestJsonnetPipelineTemplateMulti(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		yes | pachctl delete all
	`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data
		pachctl put file data@master:/foo <<<"foo-data"
		pachctl create pipeline --jsonnet - \
		  --arg firstname=foo --arg lastname=bar --arg output=baz <<EOF
		function(firstname, lastname, output) [ {
		  pipeline: { name: firstname+"-pipeline" },
		  input: {
		    pfs: {
		      name: "input",
		      glob: "/*",
		      repo: "data"
		    }
		  },
		  transform: {
		    cmd: [ "/bin/bash" ],
		    stdin: [ "cp /pfs/input/* /pfs/out/"+output+"-middle" ]
		  }
		}, {
		  pipeline: { name: lastname+"-pipeline" },
		  input: {
		    pfs: {
		      name: "input",
		      glob: "/*",
		      repo: firstname+"-pipeline"
		    }
		  },
		  transform: {
		    cmd: [ "/bin/bash" ],
		    stdin: [ "cp /pfs/input/"+output+"-middle /pfs/out/"+output+"-final" ]
		  }
		} ]
		EOF
		pachctl list pipeline | match foo-pipeline | match bar-pipeline
		pachctl wait commit foo-pipeline@master
		pachctl get file foo-pipeline@master:/baz-middle | match foo-data
		pachctl wait commit bar-pipeline@master
		pachctl get file bar-pipeline@master:/baz-final | match foo-data
		`).Run())
}

func TestJobDatumCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data
		pachctl put file data@master:/foo <<<"foo-data"
		pachctl create pipeline <<EOF
		{
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
		        "repo": "data"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["cp -rp /pfs/data/ /pfs/out"]
		    }
		  }
		EOF
	`, "pipeline", tu.UniqueString("p-")).Run())
	// need some time for the pod to spin up
	require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
		//nolint:wrapcheck
		return testutilpachctl.PachctlBashCmd(t, c, `
pachctl list job -x | match ' / 1'
`).Run()
	}, "expected to see one datum")
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `pachctl put file data@master:/bar <<<"bar-data"`).Run())
	// with the new datum, should see the pipeline run another job with two datums

	require.NoErrorWithinTRetry(t, 2*time.Minute, func() error {
		//nolint:wrapcheck
		return testutilpachctl.PachctlBashCmd(t, c, `
pachctl list job -x | match ' / 2'
`).Run()
	}, "expected to see two datums")
}

func TestListJobWithProject(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, "yes | pachctl delete all").Run())

	// Two projects, where pipeline1 is in the default project, while pipeline2 is in a different project.
	// pipeline2 takes the output of pipeline1 so that we can generate jobs across both projects.
	// We leverage the job's pipeline name to determine whether the filtering is working.
	projectName := tu.UniqueString("project-")
	pipeline1, pipeline2 := tu.UniqueString("pipeline1-"), tu.UniqueString("pipeline2-")
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data

		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"name": "{{.pipeline1}}",
				"project": {"name": "default"}
			},
			"input": {
				"pfs": {
					"name": "in",
					"glob": "/*",
					"repo": "data",
					"project": "default"
				}
			},
			"transform": {
			"cmd": ["cp", "/pfs/in/file", "/pfs/out"]
			}
		}
		EOF

		pachctl create project {{.project}}
		pachctl create pipeline <<EOF
		{
			"pipeline": {
				"name": "{{.pipeline2}}",
				"project": {"name": "{{.project}}"}
			},
			"input": {
				"pfs": {
					"name": "in",
					"glob": "/*",
					"repo": "{{.pipeline1}}",
					"project": "default"
				}
			},
			"transform": {
			"cmd": ["cp", "/pfs/in/file", "/pfs/out"]
			}
		}
		EOF

		pachctl put file data@master:/file <<<"This is a test"
		`,
		"project", projectName,
		"pipeline1", pipeline1,
		"pipeline2", pipeline2,
	).Run())

	require.NoErrorWithinTRetry(t, time.Minute, func() error {
		return errors.Wrap(testutilpachctl.PachctlBashCmd(t, c, `
			pachctl list job --raw --all-projects | match {{.pipeline1}} | match {{.pipeline2}}
			pachctl list job --raw | match {{.pipeline1}} | match -v {{.pipeline2}}
			pachctl list job --raw --project {{.project}} | match {{.pipeline2}} | match -v {{.pipeline1}}
			pachctl list job --raw --project notmyproject | match -v {{.pipeline1}} | match -v {{.pipeline2}}

			pachctl list job -x --all-projects | match {{.pipeline1}} | match {{.pipeline2}}
			pachctl list job -x | match {{.pipeline1}} | match -v {{.pipeline2}}
			pachctl list job -x --project {{.project}} | match {{.pipeline2}} | match -v {{.pipeline1}}
			pachctl list job -x --project notmyproject | match -v {{.pipeline1}} | match -v {{.pipeline2}}
		`,
			"project", projectName,
			"pipeline1", pipeline1,
			"pipeline2", pipeline2,
		).Run(),
			"failed to filter list jobs based on project")
	}, "expected to see job")
}

func TestRerunPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	c, _ := minikubetestenv.AcquireCluster(t)
	pipelineName := tu.UniqueString("p-")
	projectName := tu.UniqueString("proj")
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl create repo data
		pachctl put file data@master:/file <<<"This is a test"
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/",
		        "repo": "data"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["cp /pfs/data/file /pfs/out"]
		    }
		  }
		EOF
		pachctl create project {{.project}}
		pachctl create pipeline --project {{.project}} <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/",
		        "repo": "data",
				"project": "default"
		      }
		    },
		    "transform": {
		      "cmd": ["bash"],
		      "stdin": ["cp /pfs/data/file /pfs/out"]
		    }
		  }
		EOF
		`,
		"pipeline", pipelineName,
		"project", projectName,
	).Run())

	// Should rerun pipeine
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl rerun pipeline {{.pipeline}}
		pachctl inspect pipeline {{.pipeline}} --raw | grep '"version"' | match '"2"'
	`,
		"pipeline", pipelineName,
	).Run())

	//Wait for initial project job to complete
	jobs, err := c.ListJob(projectName, pipelineName, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl wait job {{.pipeline}}@{{.job}} --project {{.project}}
		pachctl inspect job {{.pipeline}}@{{.job}} --project {{.project}} | match {{.job}}
		pachctl inspect pipeline {{.pipeline}} --raw --project {{.project}} | grep '"version"' | match '"1"'
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | grep '"data_processed"' | match '"1"'
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | match -v "data_skipped"
	`,
		"project", projectName,
		"pipeline", pipelineName,
		"job", jobs[0].Job.GetId(),
	).Run())

	//Should rerun pipeine in the specified project.
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl rerun pipeline {{.pipeline}} --project {{.project}}
	`,
		"project", projectName,
		"pipeline", pipelineName,
	).Run())
	jobs, err = c.ListJob(projectName, pipelineName, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl wait job {{.pipeline}}@{{.job}} --project {{.project}}
		pachctl inspect job {{.pipeline}}@{{.job}} --project {{.project}} | match {{.job}}
		pachctl inspect pipeline {{.pipeline}} --raw --project {{.project}} | grep '"version"' | match '"2"'
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | match -v "data_processed"
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | grep '"data_skipped"' | match '"1"'
	`,
		"project", projectName,
		"pipeline", pipelineName,
		"job", jobs[0].Job.GetId(),
	).Run())

	// Should reprocess all datums if flag is used.
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl rerun pipeline {{.pipeline}} --reprocess --project {{.project}}
	`,
		"project", projectName,
		"pipeline", pipelineName,
	).Run())

	jobs, err = c.ListJob(projectName, pipelineName, nil, 0, false)
	require.NoError(t, err)
	require.Equal(t, 1, len(jobs))
	require.NoError(t, testutilpachctl.PachctlBashCmd(t, c, `
		pachctl wait job {{.pipeline}}@{{.job}} --project {{.project}}
		pachctl inspect job {{.pipeline}}@{{.job}} --project {{.project}} | match {{.job}}
		pachctl inspect pipeline {{.pipeline}} --raw --project {{.project}} | grep '"version"' | match '"3"'
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | grep '"data_processed"' | match '"1"'
		pachctl inspect job {{.pipeline}}@{{.job}} --raw --project {{.project}} | match -v "data_skipped"
	`,
		"project", projectName,
		"pipeline", pipelineName,
		"job", jobs[0].Job.GetId(),
	).Run())
}
