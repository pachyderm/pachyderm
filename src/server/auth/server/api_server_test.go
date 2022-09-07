package server_test

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
)

func buildClusterBindings(s ...string) *auth.RoleBinding {
	return buildBindings(append(s,
		auth.RootUser, auth.ClusterAdminRole,
		auth.InternalPrefix+"auth-server", auth.ClusterAdminRole,
	)...)
}

func buildBindings(s ...string) *auth.RoleBinding {
	var b auth.RoleBinding
	b.Entries = make(map[string]*auth.Roles)
	for i := 0; i < len(s); i += 2 {
		if _, ok := b.Entries[s[i]]; !ok {
			b.Entries[s[i]] = &auth.Roles{Roles: make(map[string]bool)}
		}
		b.Entries[s[i]].Roles[s[i+1]] = true
	}
	return &b
}

func envWithAuth(t *testing.T) *testpachd.RealEnv {
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateLicense(t, env.PachClient, peerPort, time.Now().Add(10*time.Second))
	_, err := env.PachClient.Enterprise.Activate(env.PachClient.Ctx(),
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	_, err = env.AuthServer.Activate(env.PachClient.Ctx(), &auth.ActivateRequest{RootToken: tu.RootToken})
	require.NoError(t, err, "activate server should work")
	env.PachClient.SetAuthToken(tu.RootToken)
	require.NoError(t, config.WritePachTokenToConfig(tu.RootToken, false))
	client := env.PachClient.WithCtx(context.Background())
	_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(t, err, "should be able to activate auth")
	_, err = client.PpsAPIClient.ActivateAuth(client.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(t, err, "should be able to activate auth")
	return env
}

func TestActivateUnit(t *testing.T) {
	env := envWithAuth(t)
	rootClient := tu.AuthenticateClient(t, env.PachClient, auth.RootUser)
	_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
	require.NoError(t, err)
	resp, err := rootClient.AuthAPIClient.Activate(context.Background(), &auth.ActivateRequest{})
	require.NoError(t, err)
	rootClient.SetAuthToken(resp.PachToken)
	defer func() {
		_, err := rootClient.Deactivate(rootClient.Ctx(), &auth.DeactivateRequest{})
		require.NoError(t, err)
	}()

	// Check that the token 'c' received from pachd authenticates them as "pach:root"
	who, err := rootClient.WhoAmI(rootClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, who.Username)

	bindings, err := rootClient.GetClusterRoleBinding()
	require.NoError(t, err)
	require.Equal(t, buildClusterBindings(), bindings)
}
