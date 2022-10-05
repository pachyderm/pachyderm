//go:build unit_test

package server_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"

	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/collection"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/license"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	lc "github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

const year = 365 * 24 * time.Hour

func realEnvWithLicense(t *testing.T, expireTime ...time.Time) (*realenv.RealEnv, string) {
	ctx := context.Background()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	testutil.ActivateLicense(t, env.PachClient, peerPort, expireTime...)
	_, err := env.PachClient.Enterprise.Activate(ctx,
		&enterprise.ActivateRequest{
			LicenseServer: "grpc://localhost:" + peerPort,
			Id:            "localhost",
			Secret:        "localhost",
		})
	require.NoError(t, err, "should be able to activate")
	return env, peerPort
}

func TestValidateActivationCode(t *testing.T) {
	t.Parallel()
	_, err := license.Validate(testutil.GetTestEnterpriseCode(t))
	require.NoError(t, err)
}

func TestGetState(t *testing.T) {
	t.Parallel()
	env, peerPort := realEnvWithLicense(t, time.Now().Add(year+time.Hour*24))
	client := env.PachClient

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	expires, err := types.TimestampFromProto(resp.Info.Expires)
	require.NoError(t, err)
	untilExpires := time.Until(expires)
	require.True(t, untilExpires >= year)

	activationCode, err := license.Unmarshal(resp.ActivationCode)
	require.NoError(t, err)

	require.Equal(t, "", activationCode.Signature)

	// Make current enterprise token expire
	expires = time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)

	_, err = client.License.Activate(client.Ctx(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)

	_, err = client.Enterprise.Activate(client.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "grpc://localhost:" + peerPort,
		})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_EXPIRED)

	activationCode, err = license.Unmarshal(resp.ActivationCode)
	require.NoError(t, err)

	require.Equal(t, "", activationCode.Signature)
}

func TestGetActivationCode(t *testing.T) {
	t.Parallel()
	env, peerPort := realEnvWithLicense(t, time.Now().Add(year+time.Hour*24))
	client := env.PachClient

	resp, err := client.Enterprise.GetActivationCode(client.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, testutil.GetTestEnterpriseCode(t), resp.ActivationCode)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	// Make current enterprise token expire
	expires := time.Now().Add(-30 * time.Second)
	expiresProto, err := types.TimestampProto(expires)
	require.NoError(t, err)
	_, err = client.License.Activate(client.Ctx(),
		&lc.ActivateRequest{
			ActivationCode: testutil.GetTestEnterpriseCode(t),
			Expires:        expiresProto,
		})
	require.NoError(t, err)
	_, err = client.Enterprise.Activate(client.Ctx(),
		&enterprise.ActivateRequest{
			Id:            "localhost",
			Secret:        "localhost",
			LicenseServer: "grpc://localhost:" + peerPort,
		})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetActivationCode(client.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_EXPIRED)
	require.Equal(t, testutil.GetTestEnterpriseCode(t), resp.ActivationCode)
}

func TestDeactivate(t *testing.T) {
	t.Parallel()
	env, _ := realEnvWithLicense(t)
	client := env.PachClient

	// Activate Pachyderm Enterprise and make sure the state is ACTIVE
	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_ACTIVE)

	// Deactivate cluster and make sure its state is NONE
	_, err = client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)
}

// TestDoubleDeactivate makes sure calling Deactivate() when there is no
// enterprise token works. Fixes
// https://github.com/pachyderm/pachyderm/v2/issues/3013
func TestDoubleDeactivate(t *testing.T) {
	t.Parallel()
	env, _ := realEnvWithLicense(t, time.Now().Add(year+time.Hour*24))
	client := env.PachClient

	// Deactivate cluster and make sure its state is NONE (enterprise might be
	// active at the start of this test?)
	_, err := client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.State, enterprise.State_NONE)

	// Deactivate the cluster again to make sure deactivation with no token works
	_, err = client.Enterprise.Deactivate(client.Ctx(),
		&enterprise.DeactivateRequest{})
	require.NoError(t, err)
	resp, err = client.Enterprise.GetState(client.Ctx(),
		&enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_NONE, resp.State)
}

func TestGetActivationCodeNotAdmin(t *testing.T) {
	t.Parallel()
	env := realenv.NewRealEnv(t, dockertestenv.NewTestDBConfig(t))
	peerPort := strconv.Itoa(int(env.ServiceEnv.Config().PeerPort))
	c := env.PachClient
	aliceClient := testutil.AuthenticatedPachClient(t, c, "robot:alice", peerPort)
	_, err := aliceClient.Enterprise.GetActivationCode(aliceClient.Ctx(), &enterprise.GetActivationCodeRequest{})
	require.YesError(t, err)
	require.Matches(t, "not authorized", err.Error())
}

// TestHeartbeatDeleted tests that heartbeating fails if the pachd has been
// deleted from the license server
func TestHeartbeatDeleted(t *testing.T) {
	t.Parallel()
	env, peerPort := realEnvWithLicense(t, time.Now().Add(year+time.Hour*24))
	client := env.PachClient

	resp, err := client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
	require.NoError(t, err)
	require.Equal(t, enterprise.State_ACTIVE, resp.State)

	// Delete this pachd from the license server
	_, err = client.License.DeleteCluster(client.Ctx(), &lc.DeleteClusterRequest{Id: "localhost"})
	require.NoError(t, err)

	// Trigger a heartbeat and confirm the cluster is no longer active
	_, err = client.Enterprise.Heartbeat(client.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	require.NoError(t, backoff.Retry(func() error {
		resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if resp.State != enterprise.State_HEARTBEAT_FAILED {
			return errors.Errorf("expected enterprise state to be HEARTBEAT_FAILED but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

	// Re-add the pachd and heartbeat successfully
	_, err = client.License.AddCluster(client.Ctx(),
		&lc.AddClusterRequest{
			Id:      "localhost",
			Secret:  "localhost",
			Address: "grpc://localhost:" + peerPort,
		})
	require.NoError(t, err)

	_, err = client.Enterprise.Heartbeat(client.Ctx(), &enterprise.HeartbeatRequest{})
	require.NoError(t, err)

	require.NoError(t, backoff.Retry(func() error {
		resp, err = client.Enterprise.GetState(client.Ctx(), &enterprise.GetStateRequest{})
		if err != nil {
			return errors.EnsureStack(err)
		}
		if resp.State != enterprise.State_ACTIVE {
			return errors.Errorf("expected enterprise state to be ACTIVE but was %v", resp.State)
		}
		return nil
	}, backoff.NewTestingBackOff()))

}

func TestEnterpriseConfigMigration(t *testing.T) {
	t.Parallel()
	db := dockertestenv.NewTestDB(t)
	etcd := testetcd.NewEnv(t).EtcdClient

	config := &enterprise.EnterpriseConfig{
		Id:            "id",
		LicenseServer: "server",
		Secret:        "secret",
	}

	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &enterprise.EnterpriseConfig{}, nil, nil)
	_, err := col.NewSTM(context.Background(), etcd, func(stm col.STM) error {
		return errors.EnsureStack(etcdConfigCol.ReadWrite(stm).Put("config", config))
	})
	require.NoError(t, err)

	env := migrations.Env{EtcdClient: etcd}
	// create enterprise config record in etcd
	state := migrations.InitialState().
		// the following two state changes were shipped in v2.0.0
		Apply("create collections schema", func(ctx context.Context, env migrations.Env) error {
			return collection.CreatePostgresSchema(ctx, env.Tx)
		}).
		Apply("create collections trigger functions", func(ctx context.Context, env migrations.Env) error {
			return collection.SetupPostgresV0(ctx, env.Tx)
		}).
		// the following two state changes are shipped in v2.1.0 to migrate EnterpriseConfig from etcd -> postgres
		Apply("Move EnterpriseConfig from etcd -> postgres", func(ctx context.Context, env migrations.Env) error {
			return server.EnterpriseConfigPostgresMigration(ctx, env.Tx, env.EtcdClient)
		}).
		Apply("Remove old EnterpriseConfig record from etcd", func(ctx context.Context, env migrations.Env) error {
			return server.DeleteEnterpriseConfigFromEtcd(ctx, env.EtcdClient)
		})
	// run the migration
	err = migrations.ApplyMigrations(context.Background(), db, env, state)
	require.NoError(t, err)
	err = migrations.BlockUntil(context.Background(), db, state)
	require.NoError(t, err)

	pgCol := server.EnterpriseConfigCollection(db, nil)
	result := &enterprise.EnterpriseConfig{}
	require.NoError(t, pgCol.ReadOnly(context.Background()).Get("config", result))
	require.Equal(t, config.Id, result.Id)
	require.Equal(t, config.LicenseServer, result.LicenseServer)
	require.Equal(t, config.Secret, result.Secret)

	err = etcdConfigCol.ReadOnly(context.Background()).Get("config", &enterprise.EnterpriseConfig{})
	require.YesError(t, err)
	require.True(t, collection.IsErrNotFound(err))
}
