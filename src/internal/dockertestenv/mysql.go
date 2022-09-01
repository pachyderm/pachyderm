package dockertestenv

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/sirupsen/logrus"
)

const (
	mysqlPort     = 9100
	MySQLPassword = "root"
	mysqlUser     = "root"
)

func NewEphemeralMySQLDB(t testing.TB) (*pachsql.DB, string) {
	name := testutil.GenerateEphemeralDBName(t)
	return testutil.OpenDBURL(t, newMySQLEphemeralURL(t, name), MySQLPassword), name
}

// NewMySQLURL returns a pachsql.URL to an ephemeral database.
func NewMySQLURL(t testing.TB) pachsql.URL {
	dbName := testutil.GenerateEphemeralDBName(t)
	return newMySQLEphemeralURL(t, dbName)
}

func newMySQLEphemeralURL(t testing.TB, name string) pachsql.URL {
	ctx := context.Background()
	log := logrus.StandardLogger()

	dclient := newDockerClient()
	err := ensureContainer(ctx, dclient, "pach_test_mysql", containerSpec{
		Image: "mysql:latest",
		PortMap: map[uint16]uint16{
			mysqlPort: 3306,
		},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": MySQLPassword,
		},
	})
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
	require.NoError(t, dbutil.WaitUntilReady(ctx, log, db))
	testutil.CreateEphemeralDB(t, db, name)
	u2 := u
	u2.Database = name
	return u2
}
