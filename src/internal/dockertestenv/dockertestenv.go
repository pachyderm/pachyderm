// package dockertestenv provides test environment where service dependencies are docker containers
package dockertestenv

import (
	"context"
	"os/exec"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	postgresPort  = 30228
	pgBouncerPort = 30229
	postgresHost  = "127.0.0.1"
	pgBouncerHost = postgresHost
	pgUsername    = "pachyderm"
)

func NewTestDB(t testing.TB) *sqlx.DB {
	ctx := context.Background()
	require.NoError(t, ensureDBEnv(ctx))
	opts := []dbutil.Option{
		dbutil.WithDBName(pgUsername),
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
    -e POOL_MODE=transaction \
    -p 30229:5432 \
    edoburu/pgbouncer:1.15.0
else
    echo "postgres already started"
fi
	`)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
