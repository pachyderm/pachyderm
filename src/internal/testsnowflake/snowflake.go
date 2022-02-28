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

// NewSnowSQL creates an emphermeral database in a real Snowflake instance.
func NewSnowSQL(t testing.TB) *sqlx.DB {
	ctx := context.Background()
	log := logrus.StandardLogger()

	user := os.Getenv("SNOWFLAKE_USER")
	password := os.Getenv("SNOWFLAKE_PASSWORD")
	account_identifier := os.Getenv("SNOWFLAKE_ACCOUNT")
	dsn := fmt.Sprintf("snowflake://%s@%s", user, account_identifier)

	url, err := pachsql.ParseURL(dsn)
	require.NoError(t, err)
	db := testutil.OpenDBURL(t, *url, password)

	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()

	require.NoError(t, dbutil.WaitUntilReady(ctx, log, db))
	dbname := testutil.CreateEphemeralDB(t, db)
	url.Database = dbname
	url.Schema = "public"

	return testutil.OpenDBURL(t, *url, password)
}
