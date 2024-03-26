package dockertestenv

import (
	"context"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

const (
	mysqlPort     = 9100
	MySQLPassword = "root"
	mysqlUser     = "root"
)

func NewEphemeralMySQLDB(ctx context.Context, t testing.TB) (*pachsql.DB, string) {
	name := testutil.GenerateEphemeralDBName()
	return testutil.OpenDBURL(t, newMySQLEphemeralURL(ctx, t, name), MySQLPassword), name
}

// NewMySQLURL returns a pachsql.URL to an ephemeral database.
func NewMySQLURL(ctx context.Context, t testing.TB) pachsql.URL {
	dbName := testutil.GenerateEphemeralDBName()
	return newMySQLEphemeralURL(ctx, t, dbName)
}

func newMySQLEphemeralURL(ctx context.Context, t testing.TB, name string) pachsql.URL {
	dclient := newDockerClient()
	err := backoff.Retry(func() error {
		return ensureContainer(ctx, dclient, "pach_test_mysql", containerSpec{
			Image: "mysql:latest",
			PortMap: map[uint16]uint16{
				mysqlPort: 3306,
			},
			Env: map[string]string{
				"MYSQL_ROOT_PASSWORD": MySQLPassword,
			},
		})
	}, backoff.NewConstantBackOff(time.Second*3))
	require.NoError(t, err)
	u := pachsql.URL{
		Protocol: pachsql.ProtocolMySQL,
		User:     mysqlUser,
		Host:     getDockerHost(),
		Port:     mysqlPort,
		Database: "",
	}
	db := testutil.OpenDBURL(t, u, MySQLPassword)
	ctx, cf := context.WithTimeout(ctx, 30*time.Second)
	defer cf()
	mysql.SetLogger(log.NewStdLogAt(pctx.Child(pctx.TODO(), "ephemeral-mysql"), log.DebugLevel)) //nolint:errcheck
	require.NoError(t, dbutil.WaitUntilReady(ctx, db))
	testutil.CreateEphemeralDB(t, db, name)
	u2 := u
	u2.Database = name
	return u2
}
