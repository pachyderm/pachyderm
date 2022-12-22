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
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func DSN() (string, error) {
	user := os.Getenv("SNOWSQL_USER")
	if user == "" {
		return "", errors.EnsureStack(errors.New("empty SNOWSQL_USER"))
	}
	accountID := os.Getenv("SNOWSQL_ACCOUNT")
	if accountID == "" {
		return "", errors.EnsureStack(errors.New("empty SNOWSQL_ACCOUNT"))
	}
	return fmt.Sprintf("snowflake://%s@%s", user, accountID), nil
}

func getURLAndPassword(t testing.TB) (*pachsql.URL, string) {
	password := os.Getenv("SNOWSQL_PWD")
	if password == "" {
		t.Fatal("empty SNOWSQL_PWD")
	}
	dsn, err := DSN()
	if err != nil {
		t.Fatal(err)
	}
	url, err := pachsql.ParseURL(dsn)
	if err != nil {
		t.Fatal(err)
	}
	return url, password
}

func NewEphemeralSnowflakeDB(ctx context.Context, t testing.TB) (*sqlx.DB, string) {
	name := testutil.GenerateEphemeralDBName(t)
	url, password := getURLAndPassword(t)
	db := testutil.OpenDBURL(t, *url, password)
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	require.NoError(t, dbutil.WaitUntilReady(ctx, db))

	testutil.CreateEphemeralDB(t, db, name)
	url.Database = name
	url.Schema = "public"

	return testutil.OpenDBURL(t, *url, password), name
}
