package cmds

import (
	"bytes"
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutilpachctl"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	datum    = "datum"
	job      = "job"
	pipeline = "pipeline"
	secret   = "secret"

	create  = "create"
	del     = "delete"
	edit    = "edit"
	inspect = "inspect"
	list    = "list"
	restart = "restart"
	start   = "start"
	stop    = "stop"
	update  = "update"
	wait    = "wait"
)

// TestSynonyms walks through the command tree for each resource and verb combination defined in PPS.
// A template is filled in that calls the help flag and the output is compared. It seems like 'match'
// is unable to compare the outputs correctly, but we can use diff here which returns an exit code of 0
// if there is no difference.
func TestSynonyms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	synonymCheckTemplate := `
		pachctl {{VERB}} {{RESOURCE_SYNONYM}} -h > synonym.txt
		pachctl {{VERB}} {{RESOURCE}} -h > singular.txt
		diff synonym.txt singular.txt
		rm synonym.txt singular.txt
	`

	resources := resourcesMap()
	synonyms := synonymsMap()

	for resource, verbs := range resources {
		withResource := strings.ReplaceAll(synonymCheckTemplate, "{{RESOURCE}}", resource)
		withResources := strings.ReplaceAll(withResource, "{{RESOURCE_SYNONYM}}", synonyms[resource])

		for _, verb := range verbs {
			synonymCommand := strings.ReplaceAll(withResources, "{{VERB}}", verb)
			t.Logf("Testing %s %s -h\n", verb, resource)
			require.NoError(t, testutilpachctl.BashCmd(synonymCommand).Run())
		}
	}
}

// TestSynonymsDocs is like TestSynonyms except it only tests commands registered by CreateDocsAliases.
func TestSynonymsDocs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	synonymCheckTemplate := `
		pachctl {{RESOURCE_SYNONYM}} -h > synonym.txt
		pachctl {{RESOURCE}} -h > singular.txt
		diff synonym.txt singular.txt
		rm synonym.txt singular.txt
	`

	synonyms := synonymsMap()

	for resource := range synonyms {
		if resource == "secret" {
			// no help doc defined for secret yet.
			continue
		}

		withResource := strings.ReplaceAll(synonymCheckTemplate, "{{RESOURCE}}", resource)
		synonymCommand := strings.ReplaceAll(withResource, "{{RESOURCE_SYNONYM}}", synonyms[resource])

		t.Logf("Testing %s -h\n", resource)
		require.NoError(t, testutilpachctl.BashCmd(synonymCommand).Run())
	}
}

func resourcesMap() map[string][]string {
	return map[string][]string{
		datum:    {inspect, list, restart},
		job:      {del, inspect, list, stop, wait},
		pipeline: {create, del, edit, inspect, list, start, stop, update},
		secret:   {create, del, inspect, list},
	}
}

func synonymsMap() map[string]string {
	return map[string]string{
		datum:    datums,
		job:      jobs,
		pipeline: pipelines,
		secret:   secrets,
	}
}

// TestListDatumFromFile sets up input repos in a non-default project
// and tests the ability to list datums from a pps-spec file without creating a pipeline.
func TestListDatumFromFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})

	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl create project {{.project}}
		pachctl config update context --project {{.project}}
		pachctl create repo {{.repo1}}
		pachctl create repo {{.repo2}}
		echo "foo" | pachctl put file {{.repo1}}@master:/foo
		echo "foo" | pachctl put file {{.repo2}}@master:/foo

		cat <<EOF | pachctl list datums -f -
		{
			"pipeline": {
				"name": "does-not-matter"
			},
			"input": {
				"cross":[
					{
						"pfs": {
							"repo": "{{.repo1}}",
							"glob": "/*",
						}
					},
					{
						"pfs": {
							"repo": "{{.repo2}}",
							"glob": "/*",
						}
					},
				],
			},
			"transform": {
				"cmd": ["does", "not", "matter"]
			}
		}
		EOF
	`,
		"project", tu.UniqueString("project-"),
		"repo1", tu.UniqueString("repo1-"),
		"repo2", tu.UniqueString("repo2-"),
	).Run())

}

func TestInspectClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl inspect defaults --cluster | match '{.*}'
	`,
	).Run())
}

func TestCreateClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl inspect defaults --cluster | match '{.*}'
		echo '{"create_pipeline_request": {"autoscaling": false}}' | pachctl create defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"autoscaling":false}}'
	`,
	).Run())
}

func TestDeleteClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		pachctl inspect defaults --cluster | match '{.*}'
		echo '{"create_pipeline_request": {"autoscaling": false}}' | pachctl create defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"autoscaling":false}}'
		pachctl delete defaults --cluster || exit 1
		pachctl inspect defaults --cluster | match '{}'
	`,
	).Run())
}

func TestUpdateClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t).PachConfigOption)
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, env.PachClient, `
		echo '{"create_pipeline_request": {"autoscaling": false}}' | pachctl create defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"autoscaling":false}}'
		echo '{"create_pipeline_request": {"datum_tries": "4"}}' | pachctl update defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"datum_tries":"4"}}'
		pachctl delete defaults --cluster
		pachctl inspect defaults --cluster | match '{}'
	`,
	).Run())
}

const badJSON1 = `
{
"356weryt

}
`

const badJSON2 = `{
{
    "a": 1,
    "b": [23,4,4,64,56,36,7456,7],
    "c": {"e,f,g,h,j,j},
    "d": 3452.36456,
}
`

func TestSyntaxErrorsReportedCreatePipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		echo -n '{{.badJSON1}}' \
		  | ( pachctl create pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"

		echo -n '{{.badJSON2}}' \
		  | ( pachctl create pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"
		`,
		"badJSON1", badJSON1,
		"badJSON2", badJSON2,
	).Run())
}

func TestSyntaxErrorsReportedUpdatePipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		echo -n '{{.badJSON1}}' \
		  | ( pachctl update pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"

		echo -n '{{.badJSON2}}' \
		  | ( pachctl update pipeline -f - 2>&1 || true ) \
		  | match "malformed pipeline spec"
		`,
		"badJSON1", badJSON1,
		"badJSON2", badJSON2,
	).Run())
}

func TestDeletePipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	pipelineName := tu.UniqueString("p-")
	projectName := tu.UniqueString("proj")
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
		      "stdin": ["cp /pfs/data/file /pfs/out"]
		    }
		  }
		EOF
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}2"},
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
		pachctl create project {{.project}}
		pachctl create pipeline --project {{.project}}<<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
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
		pachctl create pipeline --project {{.project}} <<EOF
		  {
		    "pipeline": {"name": "{{.pipeline}}2"},
		    "input": {
		      "pfs": {
		        "glob": "/*",
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
	// Deleting the first default project pipeline should not error, leaving
	// three remaining.
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl delete pipeline {{.pipeline}}`, "pipeline", pipelineName).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		if [ $(pachctl list pipeline --raw --all-projects | grep '"transform"' | wc -l) -ne 3 ]
		then
			exit 1
		fi
	`).Run())
	// Deleting all in the default project should delete one pipeline, leaving two.
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl delete pipeline --all`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		if [ $(pachctl list pipeline --raw --all-projects | grep '"transform"' | wc -l) -ne 2 ]
		then
			exit 1
		fi
	`).Run())
	// Deleting the first non-default project pipeline should not error, leaving
	// one remaining.
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl delete pipeline {{.pipeline}} --project {{.project}}`,
		"pipeline", pipelineName,
		"project", projectName,
	).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		if [ $(pachctl list pipeline --raw --all-projects | grep '"transform"' | wc -l) -ne 1 ]
		then
			exit 1
		fi
	`).Run())
	// Deleting all in all projects should delete one pipeline, leaving two.
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `pachctl delete pipeline --all --all-projects`).Run())
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		if [ $(pachctl list pipeline --raw --all-projects | grep '"transform"' | wc -l) -ne 0 ]
		then
			exit 1
		fi
	`).Run())
}

// TestYAMLError tests that when creating pipelines using a YAML spec with an
// error, you get an error indicating the problem in the YAML, rather than an
// error complaining about multiple documents.
//
// Note that because we parse, serialize, and re-parse pipeline specs to work
// around protobuf-generated struct tags and such, the error will refer to
// "json" (the format used for the canonicalized pipeline), but the issue
// referenced by the error (use of a string instead of an array for 'cmd') is
// the main problem below
func TestYAMLError(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	// "cmd" should be a list, instead of a string
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo input
		( pachctl create pipeline -f - 2>&1 <<EOF || true
		pipeline:
		  name: first
		input:
		  pfs:
		    glob: /*
		    repo: input
		transform:
		  cmd: /bin/bash # should be list, instead of string
		  stdin:
		    - "cp /pfs/input/* /pfs/out"
		EOF
		) | match "syntax error"
		`,
	).Run())
}

// TestYAMLTimestamp tests creating a YAML pipeline with a timestamp (i.e. the
// fix for https://github.com/pachyderm/pachyderm/v2/issues/4209)
func TestYAMLTimestamp(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	// Note that BashCmd dedents all lines below including the YAML (which
	// wouldn't parse otherwise)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		# If the pipeline comes up without error, then the YAML parsed
		pachctl create pipeline -f - <<EOF
		  pipeline:
		    name: pipeline
		  input:
		    cron:
		      name: in
		      start: "2019-10-10T22:30:05Z"
		      spec: "@yearly"
		  transform:
		    cmd: [ /bin/bash ]
		    stdin:
		      - "cp /pfs/in/* /pfs/out"
		EOF
		pachctl list pipeline | match 'pipeline'
		`,
	).Run())
}

func TestEditPipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
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
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		EDITOR="cat -u" pachctl edit pipeline my-pipeline -o yaml \
		| match 'name: my-pipeline' \
		| match 'repo: data' \
		| match 'cmd:' \
		| match 'cp /pfs/data/\* /pfs/out'
		`).Run())
	// changing the pipeline name should be an error
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `EDITOR="sed -i -e s/my-pipeline/my-pipeline2/" pachctl edit pipeline my-pipeline -o yaml`).Run())
	// changing the project name should be an error
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `EDITOR="sed -i -e s/default/default2/" pachctl edit pipeline my-pipeline -o yaml`).Run())

}

func TestMissingPipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	// should fail because there's no pipeline object in the spec
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create pipeline <<EOF
		  {
		    "transform": {
			  "image": "ubuntu:20.04"
		    },
		    "input": {
		      "pfs": {
		        "repo": "in",
		        "glob": "/*"
		      }
		    }
		  }
		EOF
	`).Run())
}

func TestUnnamedPipeline(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	// should fail because there's no pipeline name
	require.YesError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {},
		    "transform": {
			  "image": "ubuntu:20.04"
		    },
		    "input": {
		      "pfs": {
		        "repo": "in",
		        "glob": "/*"
		      }
		    }
		  }
		EOF
	`).Run())
}

func runPipelineWithImageGetStderr(t *testing.T, c *client.APIClient, image string) (string, error) {
	ctx := pctx.TestContext(t)
	// reset and create some test input
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo in
		echo "foo" | pachctl put file in@master:/file1
		echo "bar" | pachctl put file in@master:/file2
	`).Run())

	cmd := testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create pipeline <<EOF
		  {
		    "pipeline": {
		      "name": "first"
		    },
		    "transform": {
		      "image": "{{.image}}"
		    },
		    "input": {
		      "pfs": {
		        "repo": "in",
		        "glob": "/*"
		      }
		    }
		  }
		EOF
	`, "image", image)
	buf := &bytes.Buffer{}
	cmd.Stderr = buf
	// cmd.Stdout = os.Stdout // uncomment for debugging
	cmd.Env = os.Environ()
	err := cmd.Run()
	stderr := buf.String()
	return stderr, errors.EnsureStack(err)
}

func TestWarningLatestTag(t *testing.T) {
	c := pachd.NewTestPachd(t)
	// should emit a warning because user specified latest tag on docker image
	stderr, err := runPipelineWithImageGetStderr(t, c, "ubuntu:latest")
	require.NoError(t, err, "%v", err)
	require.Matches(t, "WARNING", stderr)
}

func TestWarningEmptyTag(t *testing.T) {
	c := pachd.NewTestPachd(t)
	// should emit a warning because user specified empty tag, equivalent to
	// :latest
	stderr, err := runPipelineWithImageGetStderr(t, c, "ubuntu")
	require.NoError(t, err)
	require.Matches(t, "WARNING", stderr)
}

func TestNoWarningTagSpecified(t *testing.T) {
	c := pachd.NewTestPachd(t)
	// should not emit a warning (stderr should be empty) because user
	// specified non-empty, non-latest tag
	stderr, err := runPipelineWithImageGetStderr(t, c, "ubuntu:20.04")
	require.NoError(t, err)
	require.Equal(t, "", stderr)
}

func TestJsonnetPipelineTemplateError(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create repo data
		pachctl put file data@master:/foo <<<"foo-data"
		# expect 'create pipeline' to fail, but still run 'match'
		set +e +o pipefail
		# create a pipeline, and confirm that error message appears on stderr
		# Note: discard stdout, and send stderr on stdout instead
		createpipeline_stderr="$(
		  pachctl create pipeline \
		    --jsonnet - --arg parallelism=NaN 2>&1 >/dev/null <<EOF
		function(parallelism) {
		  pipeline: { name: "pipeline" },
		  input: {
		    pfs: {
		      name: "input",
		      glob: "/*",
		      repo: "data"
		    }
		  },
		  parallelism_spec: {
		    constant: std.parseInt(parallelism)
		  },
		  transform: {
		    cmd: [ "/bin/bash" ],
		    stdin: [ "cp /pfs/input/* /pfs/out/" ]
		  }
		}
		EOF
		  )"
		createpipeline_status="$?"
		# reset error options before testing output
		set -e -o pipefail
		echo "${createpipeline_status}" | match -v "0"
		echo "${createpipeline_stderr}" | match 'NaN is not a base 10 integer'
		pachctl list pipeline | match -v foo-pipeline
		`).Run())
}

func TestPipelineWithoutSecrets(t *testing.T) {
	ctx := pctx.TestContext(t)
	c := pachd.NewTestPachd(t)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		(
		set +e +o pipefail
		pachctl create repo data
		pachctl create pipeline 2>&1 <<EOF
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
		      "stdin": ["cp -rp /pfs/data/ /pfs/out"],
		      "secrets": [{"name": "{{.secret}}", "env_var": "PASSWORD", "key": "PASSWORD"}]
		    }
		}
		EOF
		true) | match "missing Kubernetes secret"
	`, "pipeline", tu.UniqueString("p-"), "secret", tu.UniqueString("supersecret")).Run())
}

func TestPipelineWithMissingSecret(t *testing.T) {
	ctx := pctx.TestContext(t)
	var k8sClient kubernetes.Interface
	c := pachd.NewTestPachd(t, pachd.GetK8sClient(&k8sClient))
	secretName := "supersecret"
	_, err := k8sClient.CoreV1().Secrets("default").Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"NOTPASSWORD": []byte("correcthorsebatterystaple"),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		(
		set +e +o pipefail
		pachctl create repo data
		pachctl create pipeline 2>&1 <<EOF
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
		      "stdin": ["cp -rp /pfs/data/ /pfs/out"],
		      "secrets": [{"name": "{{.secret}}", "env_var": "PASSWORD", "key": "PASSWORD"}]
		    }
		}
		EOF
		true) | match "missing key"
	`, "pipeline", tu.UniqueString("p-"), "secret", secretName).Run())
}

func TestPipelineWithSecret(t *testing.T) {
	ctx := pctx.TestContext(t)
	var k8sClient kubernetes.Interface
	c := pachd.NewTestPachd(t, pachd.GetK8sClient(&k8sClient))
	secretName := "supersecret"
	_, err := k8sClient.CoreV1().Secrets("default").Create(context.Background(), &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		Data: map[string][]byte{
			"PASSWORD": []byte("correcthorsebatterystaple"),
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, c, `
		pachctl create project myProject
		pachctl create repo data --project myProject
		pachctl create pipeline --project myProject 2>&1 <<EOF
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
		      "stdin": ["cp -rp /pfs/data/ /pfs/out"],
		      "secrets": [{"name": "{{.secret}}", "env_var": "PASSWORD", "key": "PASSWORD"}]
		    }
		}
		EOF
	`, "pipeline", tu.UniqueString("p-"), "secret", secretName).Run())
}

func TestProjectDefaultsMetadata(t *testing.T) {
	ctx := pctx.TestContext(t)
	rootClient := pachd.NewTestPachd(t, pachd.ActivateAuthOption(""))
	projectName := tu.UniqueString("proj")
	require.NoError(t, testutilpachctl.PachctlBashCmdCtx(ctx, t, rootClient, `
		pachctl create project {{.projectName}}
		echo '{"createPipelineRequest": {"autoscaling": true}}' | pachctl create defaults --project {{.projectName}}
		pachctl inspect defaults --project {{.projectName}} --raw | match '"createdBy":"pach:root"'
	`,
		"projectName", projectName,
	).Run())
}
