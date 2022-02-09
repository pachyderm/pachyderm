package main

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestDeployEnterprise(t *testing.T) {
	k := testutil.GetKubeClient(t)
	minikubetestenv.DeleteRelease(t, context.Background(), k) // cleanup cluster
	c := minikubetestenv.InstallReleaseEnterprise(t, context.Background(), k, auth.RootUser)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
}
