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
	"github.com/pachyderm/pachyderm/v2/src/license"
)

func activateLicenseFor(ctx context.Context, t *testing.T, env *testpachd.RealEnv, time time.Time) {
	code := tu.GetTestEnterpriseCode(t)
	_, err := env.PachClient.License.Activate(ctx,
		&license.ActivateRequest{
			ActivationCode: code,
			Expires:        tu.TSProtoOrDie(t, time),
		})
	require.NoError(t, err)
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	_, err = env.PachClient.License.AddCluster(ctx, &license.AddClusterRequest{
		Id:                  "localhost",
		Address:             "grpc://localhost:" + peerPort,
		UserAddress:         "grpc://localhost:" + peerPort,
		Secret:              "localhost",
		ClusterDeploymentId: "test",
		EnterpriseServer:    true,
	})
	require.NoError(t, err)
}

func TestEnterpriseDeactivate(t *testing.T) {
	ctx := context.Background()
	env := testpachd.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	resp, err := env.PachClient.Enterprise.GetState(ctx, &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)
	activateLicenseFor(ctx, t, env, time.Now().Add(5*time.Second))
	_, err = env.PachClient.Enterprise.Activate(ctx,
		&enterprise.ActivateRequest{
			LicenseServer: "127.0.0.1:" + peerPort,
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
	activateLicenseFor(ctx, t, env, time.Now().Add(2*time.Second))
	_, err := env.PachClient.Enterprise.Activate(ctx,
		&enterprise.ActivateRequest{
			LicenseServer: "127.0.0.1:" + peerPort,
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
