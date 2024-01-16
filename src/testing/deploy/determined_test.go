//go:build k8s

package main

import (
	"context"
	"testing"
	"time"

	"github.com/determined-ai/determined/proto/pkg/userv1"
	"github.com/determined-ai/determined/proto/pkg/workspacev1"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	det "github.com/pachyderm/pachyderm/v2/src/internal/determined"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"golang.org/x/exp/maps"
)

func TestDeterminedUserSync(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
		PortOffset: portOffset,
		Determined: true,
	}
	valueOverrides := make(map[string]string)
	maps.Copy(valueOverrides, globalValueOverrides)
	valueOverrides["pachd.replicas"] = "1"
	opts.ValueOverrides = valueOverrides
	minikubetestenv.PutNamespace(t, ns)
	t.Logf("Determined installing in namespace %s", ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	// collect initial determined users list
	detUrl := minikubetestenv.DetNodeportHttpUrl(t, ns)
	ctx := pctx.TestContext(t)
	dc, cf, err := det.NewClient(ctx, detUrl.String(), false)
	require.NoError(t, err)
	defer func() { require.NoError(t, cf()) }()
	token, err := det.MintToken(ctx, dc, "admin", "")
	require.NoError(t, err)
	ctx = det.WithToken(ctx, token)
	previous, err := det.GetUsers(ctx, dc)
	require.NoError(t, err)
	// login to pachyderm with mock user
	mockIDPLogin(t, c)
	// assert that after logging into pachyderm, a determined user is created
	current, err := det.GetUsers(ctx, dc)
	require.NoError(t, err)
	require.Equal(t, len(previous)+1, len(current), "the new pipeline has created an additional service user in Determined")

}

func TestDeterminedInstallAndIntegration(t *testing.T) {
	t.Parallel()
	valueOverrides := make(map[string]string)
	maps.Copy(valueOverrides, globalValueOverrides)
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
		PortOffset: portOffset,
		Determined: true,
	}
	valueOverrides["pachd.replicas"] = "1"
	opts.ValueOverrides = valueOverrides
	t.Logf("Determined installing in namespace %s", ns)
	ctx := pctx.TestContext(t)
	c := minikubetestenv.InstallRelease(t, ctx, ns, k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// Log in and create a non-admin user with the kilgore email from pachyderm.
	// The user should already exist in Determined due to the user synching system.
	detUrl := minikubetestenv.DetNodeportHttpUrl(t, ns)
	dc, cf, err := det.NewClient(ctx, detUrl.String(), false)
	require.NoError(t, err)
	defer func() { require.NoError(t, cf()) }()
	token, err := det.MintToken(ctx, dc, "admin", "")
	require.NoError(t, err)
	ctx = det.WithToken(ctx, token)
	repoName := "images"
	pipelineName := "edges"
	workspaceName := "pach-test-workspace"
	workspace, err := det.CreateWorkspace(ctx, dc, workspaceName)
	require.NoError(t, err)
	previous, err := det.GetUsers(ctx, dc)
	require.NoError(t, err)
	var user *userv1.User
	for _, u := range previous {
		if u.GetUsername() == "kilgore@kilgore.trout" {
			user = u
			break
		}
	}
	require.NotEqual(t, user, nil)
	role, err := det.GetRole(ctx, dc, "Editor")
	require.NoError(t, err)
	require.NoError(t, det.AssignRole(ctx, dc, user.Id, role.RoleId, []*workspacev1.Workspace{workspace}))
	// create repo and pipeline that should make the determined service user
	require.NoError(t, c.CreateRepo(pfs.DefaultProjectName, repoName))
	_, err = c.PpsAPIClient.CreatePipeline(
		c.Ctx(),
		&pps.CreatePipelineRequest{
			Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName),
			Transform: &pps.Transform{
				Image: "pachyderm/opencv:1.0",
				Cmd:   []string{"python3", "/edges.py"},
				Stdin: nil,
			},
			ParallelismSpec: nil,
			Input:           &pps.Input{Pfs: &pps.PFSInput{Glob: "/", Repo: repoName}},
			OutputBranch:    "master",
			Update:          false,
			Determined: &pps.Determined{
				Workspaces: []string{workspaceName},
			},
		},
	)
	require.NoError(t, err)
	current, err := det.GetUsers(ctx, dc)
	require.NoError(t, err)
	require.Equal(t, len(previous)+1, len(current), "the new pipeline has created an additional service user in Determined")
	// once a pipeline is deleted the users should eventually be cleaned up in determined
	_, err = c.PpsAPIClient.DeletePipeline(c.Ctx(), &pps.DeletePipelineRequest{Pipeline: client.NewPipeline(pfs.DefaultProjectName, pipelineName)})
	require.NoError(t, err)
	previous = current
	require.NoErrorWithinTRetryConstant(t, 2*time.Minute, func() error {
		current, err = det.GetUsers(ctx, dc)
		require.NoError(t, err)
		if len(previous)-1 == len(current) {
			return errors.Errorf("the new pipeline has created an additional service user in Determined")
		}
		return nil
	}, 5*time.Second)
}
