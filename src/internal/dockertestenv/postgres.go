// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	postgresPort  = 30228
	PGBouncerPort = 30229
	maxOpenConns  = 10
)

func postgresHost() string {
	return getDockerHost()
}

func PGBouncerHost() string {
	return postgresHost()
}

func NewTestDBConfig(t testing.TB) pachconfig.ConfigOption {
	var (
		ctx     = pctx.Background("testDB")
		dbName  = testutil.GenerateEphemeralDBName(t)
		dexName = testutil.UniqueString("dex")
	)
	err := backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	testutil.CreateEphemeralDB(t, db, dbName)
	testutil.CreateEphemeralDB(t, db, dexName)
	return func(c *pachconfig.Configuration) {
		// common
		c.PostgresDBName = dbName
		c.IdentityServerDatabase = dexName

		// direct
		c.PostgresHost = postgresHost()
		c.PostgresPort = postgresPort
		// pg_bouncer
		c.PGBouncerHost = PGBouncerHost()
		c.PGBouncerPort = PGBouncerPort

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
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	testutil.CreateEphemeralDB(t, db, name)
	return testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
		dbutil.WithDBName(name),
	), name
}

func NewTestDBOptions(t testing.TB) []dbutil.Option {
	ctx := context.Background()
	err := backoff.Retry(func() error {
		return ensureDBEnv(t, ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
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
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(postgresHost(), postgresPort),
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

	dclient := newDockerClient()
	defer dclient.Close()
	if err := ensureContainer(ctx, dclient, "pach_test_postgres", containerSpec{
		Env: map[string]string{
			"POSTGRES_DB":               "pachyderm",
			"POSTGRES_USER":             "pachyderm",
			"POSTGRES_HOST_AUTH_METHOD": "trust",
		},
		PortMap: map[uint16]uint16{
			30228: 5432,
		},
		Image: "postgres:13.0-alpine",
		Cmd:   []string{"postgres", "-c", "max_connections=500"},
	}); err != nil {
		return errors.EnsureStack(err)
	}

	containerJSON, err := dclient.ContainerInspect(ctx, "pach_test_postgres")
	if err != nil {
		return errors.EnsureStack(err)
	}

	postgresIP := containerJSON.NetworkSettings.IPAddress

	if err := ensureContainer(ctx, dclient, "pach_test_pgbouncer", containerSpec{
		Env: map[string]string{
			"PGBOUNCER_AUTH_TYPE":                 "any",
			"POSTGRESQL_USERNAME":                 "pachyderm",
			"POSTGRESQL_PASSWORD":                 "password",
			"POSTGRESQL_HOST":                     postgresIP,
			"POSTGRESQL_PORT":                     "5432",
			"PGBOUNCER_MAX_CLIENT_CONN":           "100000",
			"PGBOUNCER_POOL_MODE":                 "transaction",
			"PGBOUNCER_IGNORE_STARTUP_PARAMETERS": "extra_float_digits",
		},
		PortMap: map[uint16]uint16{
			30229: 5432,
		},
		Image: "pachyderm/pgbouncer:1.16.2",
	}); err != nil {
		return errors.EnsureStack(err)
	}

	return backoff.RetryUntilCancel(ctx, func() error {
		db, err := dbutil.NewDB(
			dbutil.WithDBName(testutil.DefaultPostgresDatabase),
			dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
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
