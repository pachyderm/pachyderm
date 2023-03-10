// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"sync"
	"testing"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	postgresPort  = 30228
	pgBouncerPort = 30229
	maxOpenConns  = 10
)

func postgresHost(dclient docker.APIClient) string {
	return getDockerHost(dclient)
}

func pgBouncerHost(dclient docker.APIClient) string {
	return postgresHost(dclient)
}

func PGBouncerHostPort() (string, int, error) {
	dclient, err := newDockerClient(context.TODO())
	if err != nil {
		return "", 0, err
	}
	defer dclient.Close()
	return getDockerHost(dclient), pgBouncerPort, nil
}

func NewTestDBConfig(t testing.TB) serviceenv.ConfigOption {
	var (
		ctx     = context.Background()
		dbName  = testutil.GenerateEphemeralDBName(t)
		dexName = testutil.UniqueString("dex")
	)
	dclient, err := newDockerClient(context.TODO())
	require.NoError(t, err)
	err = backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(pgBouncerHost(dclient), pgBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	testutil.CreateEphemeralDB(t, db, dbName)
	testutil.CreateEphemeralDB(t, db, dexName)
	return func(c *serviceenv.Configuration) {
		// common
		c.PostgresDBName = dbName
		c.IdentityServerDatabase = dexName

		// direct
		c.PostgresHost = postgresHost(dclient)
		c.PostgresPort = postgresPort
		// pg_bouncer
		c.PGBouncerHost = pgBouncerHost(dclient)
		c.PGBouncerPort = pgBouncerPort

		c.PostgresUser = testutil.DefaultPostgresUser
	}
}

func NewTestDB(t testing.TB) *pachsql.DB {
	return testutil.OpenDB(t, NewTestDBOptions(t)...)
}

// NewEphemeralPostgresDB creates a randomly-named new database, returning a
// connection to the new DB and the name itself.
func NewEphemeralPostgresDB(ctx context.Context, t testing.TB) (*pachsql.DB, string) {
	var (
		name = testutil.GenerateEphemeralDBName(t)
	)
	err := backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	dclient, err := newDockerClient(ctx)
	require.NoError(t, err)
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(pgBouncerHost(dclient), pgBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	testutil.CreateEphemeralDB(t, db, name)
	return testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(pgBouncerHost(dclient), pgBouncerPort),
		dbutil.WithDBName(name),
	), name
}

func NewTestDBOptions(t testing.TB) []dbutil.Option {
	ctx := pctx.TODO()
	err := backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	dclient, err := newDockerClient(ctx)
	require.NoError(t, err)
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(pgBouncerHost(dclient), pgBouncerPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(maxOpenConns),
	})
}

func NewTestDirectDBOptions(t testing.TB) []dbutil.Option {
	ctx := context.Background()
	err := backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	dclient, err := newDockerClient(ctx)
	require.NoError(t, err)
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(postgresHost(dclient), postgresPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(maxOpenConns),
	})
}

var spawnLock sync.Mutex

// TODO: use the docker client, instead of the bash script
// TODO: use the bitnami pg_bouncer image
// TODO: look into https://github.com/ory/dockertest
func ensureDBEnv(t testing.TB, ctx context.Context) error {
	spawnLock.Lock()
	defer spawnLock.Unlock()
	timeout := 30 * time.Second
	ctx, cf := context.WithTimeout(ctx, timeout)
	defer cf()

	dclient, err := newDockerClient(ctx)
	if err != nil {
		return err
	}
	defer dclient.Close()
	log.Info(ctx, "created docker client")
	err = ensureContainer(ctx, dclient, "pach_test_postgres", containerSpec{
		Env: map[string]string{
			"POSTGRES_DB":               "pachyderm",
			"POSTGRES_USER":             "pachyderm",
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		PortMap: map[uint16]uint16{
			30228: 5432,
		},
		Image: "postgres:13.0-alpine",
	})
	if err != nil {
		return errors.EnsureStack(err)
	}

	containerJSON, err := dclient.ContainerInspect(ctx, "pach_test_postgres")
	if err != nil {
		return errors.EnsureStack(err)
	}

	postgresIP := containerJSON.NetworkSettings.IPAddress

	err = ensureContainer(ctx, dclient, "pach_test_pgbouncer", containerSpec{
		Env: map[string]string{
			"AUTH_TYPE":       "any",
			"DB_USER":         "pachyderm",
			"DB_PASS":         "password",
			"DB_HOST":         postgresIP,
			"DB_PORT":         "5432",
			"MAX_CLIENT_CONN": "1000",
			"POOL_MODE":       "transaction",
		},
		PortMap: map[uint16]uint16{
			30229: 5432,
		},
		Image: "edoburu/pgbouncer:1.15.0",
	})
	if err != nil {
		return errors.EnsureStack(err)
	}

	return backoff.RetryUntilCancel(ctx, func() error {
		db, err := dbutil.NewDB(
			dbutil.WithDBName(testutil.DefaultPostgresDatabase),
			dbutil.WithHostPort(pgBouncerHost(dclient), pgBouncerPort),
			dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		)
		if err != nil {
			t.Logf("error connecting to db: %v", err)
			return err
		}
		defer db.Close()
		return errors.EnsureStack(db.PingContext(ctx))
	}, backoff.RetryEvery(time.Second), func(err error, _ time.Duration) error {
		return nil
	})
}
