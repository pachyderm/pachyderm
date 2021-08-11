// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/sirupsen/logrus"
)

const (
	postgresPort  = 30228
	pgBouncerPort = 30229
	maxOpenConns  = 10
)

func postgresHost() string {
	return "127.0.0.1"
	// TODO: this lets you use minikube or docker desktop on macOS, but doesn't work in CI.
	// endpoint, isSet := os.LookupEnv("DOCKER_HOST")
	// if !isSet {
	// 	return "127.0.0.1"
	// }
	// u, err := url.Parse(endpoint)
	// if err != nil {
	// 	panic(err)
	// }
	// host, _, err := net.SplitHostPort(u.Host)
	// if err != nil {
	// 	panic(err)
	// }
	// return host
}

func pgBouncerHost() string {
	return postgresHost()
}

func NewTestDBConfig(t testing.TB) serviceenv.ConfigOption {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(t, ctx))
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(pgBouncerHost(), pgBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	dbName := testutil.CreateEphemeralDB(t, db)
	return func(c *serviceenv.Configuration) {
		// common
		c.PostgresDBName = dbName

		// direct
		c.PostgresHost = postgresHost()
		c.PostgresPort = postgresPort
		// pg_bouncer
		c.PGBouncerHost = pgBouncerHost()
		c.PGBouncerPort = pgBouncerPort

		c.PostgresUser = testutil.DefaultPostgresUser
	}
}

func NewTestDB(t testing.TB) *sqlx.DB {
	opts := NewTestDBOptions(t)
	return testutil.NewTestDB(t, opts)
}

func NewTestDirectDB(t testing.TB) *sqlx.DB {
	opts := NewTestDirectDBOptions(t)
	return testutil.NewTestDB(t, opts)
}

func NewTestDBOptions(t testing.TB) []dbutil.Option {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(t, ctx))
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(pgBouncerHost(), pgBouncerPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(maxOpenConns),
	})
}

func NewTestDirectDBOptions(t testing.TB) []dbutil.Option {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(t, ctx))
	return testutil.NewTestDBOptions(t, []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
		dbutil.WithHostPort(postgresHost(), postgresPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithMaxOpenConns(maxOpenConns),
	})
}

var spawnLock sync.Mutex

// TODO: use the docker client.
func ensureDBEnv(t testing.TB, ctx context.Context) error {
	spawnLock.Lock()
	defer spawnLock.Unlock()
	cmd := exec.CommandContext(ctx, "bash", "-c", `
set -ve

unset DOCKER_HOST
unset DOCKER_CERT_PATH
unset DOCKER_TLS_VERIFY

if ! docker ps | grep -q postgres
then
    echo "starting postgres..."
    postgres_id=$(docker run -d \
    -e POSTGRES_DB=pachyderm \
    -e POSTGRES_USER=pachyderm \
    -e POSTGRES_HOST_AUTH_METHOD=trust \
    -p 30228:5432 \
    postgres:13.0-alpine)

    postgres_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' $postgres_id)

    docker run -d \
    -e AUTH_TYPE=any \
    -e DB_USER="pachyderm" \
    -e DB_PASS="password" \
    -e DB_HOST=$postgres_ip \
    -e DB_PORT=5432 \
	-e MAX_CLIENT_CONN=1000 \
    -e POOL_MODE=transaction \
    -p 30229:5432 \
    edoburu/pgbouncer:1.15.0
else
    echo "postgres already started"
fi
	`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return err
	}
	timeout := 30 * time.Second
	ctx, cf := context.WithTimeout(ctx, timeout)
	defer cf()
	return backoff.RetryUntilCancel(ctx, func() error {
		db, err := dbutil.NewDB(
			dbutil.WithDBName(testutil.DefaultPostgresDatabase),
			dbutil.WithHostPort(pgBouncerHost(), pgBouncerPort),
			dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		)
		if err != nil {
			logrus.Error("error connecting to db:", err)
			return err
		}
		defer db.Close()
		return db.PingContext(ctx)
	}, backoff.RetryEvery(time.Second), func(err error, _ time.Duration) error {
		return nil
	})
}
