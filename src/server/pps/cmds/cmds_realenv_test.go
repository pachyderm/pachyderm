//go:build unit_test

package cmds

import (
	"context"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
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
			require.NoError(t, tu.BashCmd(synonymCommand).Run())
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
		require.NoError(t, tu.BashCmd(synonymCommand).Run())
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})

	require.NoError(t, tu.PachctlBashCmd(t, env.PachClient, `
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, tu.PachctlBashCmd(t, env.PachClient, `
		pachctl inspect defaults --cluster | match '{.*}'
	`,
	).Run())
}

func TestCreateClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, tu.PachctlBashCmd(t, env.PachClient, `
		pachctl inspect defaults --cluster | match '{.*}'
		echo '{"create_pipeline_request": {"autoscaling": false}}' | pachctl create defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"autoscaling":false}}'
	`,
	).Run())
}

func TestDeleteClusterDefaults(t *testing.T) {
	ctx := pctx.TestContext(t)
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, tu.PachctlBashCmd(t, env.PachClient, `
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
	env := realenv.NewRealEnv(ctx, t, dockertestenv.NewTestDBConfig(t))
	env.MockPachd.Admin.InspectCluster.Use(func(context.Context, *admin.InspectClusterRequest) (*admin.ClusterInfo, error) {
		return &admin.ClusterInfo{
			Id:           "dev",
			DeploymentId: "dev",
			WarningsOk:   true,
		}, nil
	})
	require.NoError(t, tu.PachctlBashCmd(t, env.PachClient, `
		echo '{"create_pipeline_request": {"autoscaling": false}}' | pachctl create defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"autoscaling":false}}'
		echo '{"create_pipeline_request": {"datum_tries": "4"}}' | pachctl update defaults --cluster -f - || exit 1
		pachctl inspect defaults --cluster | match '{"create_pipeline_request":{"datum_tries": "4"}}'
		pachctl delete defaults --cluster
		pachctl inspect defaults --cluster | match '{}'
	`,
	).Run())
}
