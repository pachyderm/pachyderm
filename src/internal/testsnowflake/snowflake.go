// package testsnowflake provides convenience functions for creating Snowflake databases for testing.
package testsnowflake

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/sirupsen/logrus"
)

func DSN() string {
	user := os.Getenv("SNOWFLAKE_USER")
	account_identifier := os.Getenv("SNOWFLAKE_ACCOUNT")
	return fmt.Sprintf("snowflake://%s@%s", user, account_identifier)
}

func getURLAndPassword(t testing.TB) (*pachsql.URL, string) {
	password := os.Getenv("SNOWFLAKE_PASSWORD")
	url, err := pachsql.ParseURL(DSN())
	require.NoError(t, err)
	return url, password
}

// NewSnowSQL creates an emphermeral database in a real Snowflake instance.
func NewSnowSQL(t testing.TB) *sqlx.DB {
	return NewEphemeralSnowflakeDB(t, testutil.GenerateEphermeralDBName(t))
}

func NewEphemeralSnowflakeDB(t testing.TB, dbName string) *sqlx.DB {
	url, password := getURLAndPassword(t)
	db := testutil.OpenDBURL(t, *url, password)
	ctx := context.Background()
	log := logrus.StandardLogger()
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	require.NoError(t, dbutil.WaitUntilReady(ctx, log, db))

	testutil.CreateEphemeralDB(t, db, dbName)
	url.Database = dbName
	url.Schema = "public"

	return testutil.OpenDBURL(t, *url, password)
}
