// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	postgresPort  = 30228
	pgBouncerPort = 30229
	postgresHost  = "127.0.0.1"
	pgBouncerHost = postgresHost
)

func NewTestDBConfig(t testing.TB) serviceenv.ConfigOption {
	db := testutil.OpenDB(t,
		dbutil.WithMaxOpenConns(1),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
		dbutil.WithHostPort(pgBouncerHost, pgBouncerPort),
		dbutil.WithDBName(testutil.DefaultPostgresDatabase),
	)
	dbName := testutil.CreateEphemeralDB(t, db)
	return func(c *serviceenv.Configuration) {
		// common
		c.PostgresDBName = dbName

		// direct
		c.PostgresHost = postgresHost
		c.PostgresPort = postgresPort
		// pg_bouncer
		c.PGBouncerHost = pgBouncerHost
		c.PGBouncerPort = pgBouncerPort

		c.PostgresUser = testutil.DefaultPostgresUser
	}
}

func NewTestDB(t testing.TB) *sqlx.DB {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(ctx))
	opts := []dbutil.Option{
		dbutil.WithDBName(testutil.DefaultPostgresUser),
		dbutil.WithHostPort(pgBouncerHost, pgBouncerPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
	}
	return testutil.NewTestDB(t, opts...)
}

func NewTestDirectDB(t testing.TB) *sqlx.DB {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(ctx))
	opts := []dbutil.Option{
		dbutil.WithDBName("pachyderm"),
		dbutil.WithHostPort(postgresHost, postgresPort),
		dbutil.WithUserPassword(testutil.DefaultPostgresUser, testutil.DefaultPostgresPassword),
	}
	return testutil.NewTestDirectDB(t, opts...)
}

// TODO: use the docker client.
func ensureDBEnv(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "bash", "-c", `
set -ve

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
	-e DB_MAX_CONN=1000 \
    -e POOL_MODE=transaction \
    -p 30229:5432 \
    edoburu/pgbouncer:1.15.0
else
    echo "postgres already started"
fi
	`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	if ec := cmd.ProcessState.ExitCode(); ec != 0 {
		fmt.Println(string(output))
		return errors.Errorf("dockertestenv: non zero exit code during setup %v", ec)
	}
	return nil
}
