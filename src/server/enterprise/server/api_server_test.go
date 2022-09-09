package server_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestEnterpriseDeactivate(t *testing.T) {
	ctx := context.Background()
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	resp, err := env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)
	tu.ActivateLicense(t, env.PachClient, peerPort, time.Now().Add(10*time.Second))
	_, err = env.PachClient.Enterprise.Activate(ctx,
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err)
	_, err = env.PachClient.Enterprise.Heartbeat(ctx, &enterprise.HeartbeatRequest{})
	require.NoError(t, err)
	resp, err = env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)
	_, err = env.PachClient.Enterprise.Deactivate(ctx, &enterprise.DeactivateRequest{})
	require.NoError(t, err)
	resp, err = env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)
}

func TestEnterpriseExpired(t *testing.T) {
	ctx := context.Background()
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	tu.ActivateLicense(t, env.PachClient, peerPort, time.Now().Add(2*time.Second))
	_, err := env.PachClient.Enterprise.Activate(ctx,
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err)
	resp, err := env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)
	time.Sleep(time.Second * 3)
	resp, err = env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_EXPIRED)
	codeResp, err := env.PachClient.Enterprise.GetActivationCode(ctx, &enterprise.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, codeResp.State, enterprise.State_EXPIRED)
}
