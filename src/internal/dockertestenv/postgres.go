// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"go.uber.org/zap"
)

const (
	postgresPort            = 30228
	PGBouncerPort           = 30229
	maxOpenConns            = 10
	DefaultPostgresPassword = "correcthorsebatterystable"
	DefaultPostgresUser     = "pachyderm"
	DefaultPostgresDatabase = "pachyderm"
)

func postgresHost() string {
	return getDockerHost()
}

func PGBouncerHost() string {
	return postgresHost()
}

type PostgresConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string

	DBName string
}

func (pgc PostgresConfig) DBOptions() []dbutil.Option {
	return []dbutil.Option{
		dbutil.WithMaxOpenConns(10),
		dbutil.WithUserPassword(pgc.User, pgc.Password),
		dbutil.WithHostPort(pgc.Host, int(pgc.Port)),
		dbutil.WithDBName(pgc.DBName),
	}
}

type DBConfig struct {
	Direct    PostgresConfig
	PGBouncer PostgresConfig
	Identity  PostgresConfig
}

func (dbc DBConfig) PachConfigOption(c *pachconfig.Configuration) {
	// common
	c.PostgresDBName = dbc.PGBouncer.DBName
	c.IdentityServerDatabase = dbc.Identity.DBName
	c.PostgresUser = dbc.PGBouncer.User
	c.PostgresPassword = dbc.PGBouncer.Password

	// direct
	c.PostgresHost = dbc.Direct.Host
	c.PostgresPort = int(dbc.Direct.Port)

	// pg_bouncer
	c.PGBouncerHost = dbc.PGBouncer.Host
	c.PGBouncerPort = int(dbc.PGBouncer.Port)
}

// NewTestDBConfig returns a DBConfig for a test environment
// The environment will be torn down at the end of the test.
func NewTestDBConfig(t testing.TB) DBConfig {
	var (
		ctx     = pctx.TestContext(t)
		dbName  = testutil.GenerateEphemeralDBName(t)
		dexName = testutil.UniqueString("dex")
	)
	err := backoff.Retry(func() error {
		return EnsureDBEnv(ctx)
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err, "DB should be created")
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
		dbutil.WithDBName(DefaultPostgresDatabase),
	)
	testutil.CreateEphemeralDB(t, db, dbName)
	testutil.CreateEphemeralDB(t, db, dexName)
	return DBConfig{
		Direct: PostgresConfig{
			Host:     postgresHost(),
			Port:     postgresPort,
			User:     DefaultPostgresUser,
			Password: DefaultPostgresPassword,
			DBName:   dbName,
		},
		PGBouncer: PostgresConfig{
			Host:     PGBouncerHost(),
			Port:     PGBouncerPort,
			User:     DefaultPostgresUser,
			Password: DefaultPostgresPassword,
			DBName:   dbName,
		},
		Identity: PostgresConfig{
			Host:     PGBouncerHost(),
			Port:     PGBouncerPort,
			User:     DefaultPostgresUser,
			Password: DefaultPostgresPassword,
			DBName:   dexName,
		},
	}
}

// NewTestDB creates a new database connection scoped to the test.
func NewTestDB(t testing.TB) *pachsql.DB {
	cfg := NewTestDBConfig(t)
	return testutil.OpenDB(t, cfg.PGBouncer.DBOptions()...)
}

func NewTestDirectDB(t testing.TB) *pachsql.DB {
	cfg := NewTestDBConfig(t)
	return testutil.OpenDB(t, cfg.Direct.DBOptions()...)
}

// NewEphemeralPostgresDB creates a randomly-named new database, returning a
// connection to the new DB and the name itself.
func NewEphemeralPostgresDB(ctx context.Context, t testing.TB) (*pachsql.DB, string) {
	c := NewTestDBConfig(t)
	return testutil.OpenDB(t, c.Direct.DBOptions()...), c.Direct.DBName
}

var spawnLock sync.Mutex

// TODO: use the docker client, instead of the bash script
// TODO: use the bitnami pg_bouncer image
// TODO: look into https://github.com/ory/dockertest
func EnsureDBEnv(ctx context.Context) error {
	// bazel run //src/testing/cmd/dockertestenv creates these for many CI runs.
	if got, want := os.Getenv("SKIP_DOCKER_POSTGRES_CREATE"), "1"; got == want {
		log.Info(ctx, "not attempting to create docker container; SKIP_DOCKER_POSTGRES_CREATE=1")
		return nil
	}

	spawnLock.Lock()
	defer spawnLock.Unlock()
	timeout := 120 * time.Second
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
		Image: "postgres:15-alpine",
		Cmd:   []string{"postgres", "-c", "max_connections=500", "-c", "fsync=off"},
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
			dbutil.WithDBName(DefaultPostgresDatabase),
			dbutil.WithHostPort(PGBouncerHost(), PGBouncerPort),
			dbutil.WithUserPassword(DefaultPostgresUser, DefaultPostgresPassword),
		)
		if err != nil {
			log.Info(ctx, "failed to connect to database; retrying", zap.Error(err))
			return errors.Wrap(err, "connect to db")
		}
		defer db.Close()
		return errors.Wrap(db.PingContext(ctx), "ping db")
	}, backoff.RetryEvery(time.Second), func(err error, _ time.Duration) error {
		return nil
	})
}
