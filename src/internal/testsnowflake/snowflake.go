// package testsnowflake provides convenience functions for creating Snowflake databases for testing.
package testsnowflake

import (
	"context"
	"errors"
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

func DSN() (string, error) {
	user := os.Getenv("SNOWFLAKE_USER")
	if user == "" {
		return "", errors.New("empty SNOWFLAKE_USER")
	}
	accountID := os.Getenv("SNOWFLAKE_ACCOUNT")
	if accountID == "" {
		return "", errors.New("empty SNOWFLAKE_ACCOUNT")
	}
	return fmt.Sprintf("snowflake://%s@%s", user, accountID), nil
}

func getURLAndPassword(t testing.TB) (*pachsql.URL, string) {
	password := os.Getenv("SNOWFLAKE_PASSWORD")
	if password == "" {
		t.Fatal("empty SNOWFLAKE_PASSWORD")
	}
	dsn, err := DSN()
	if err != nil {
		t.Fatal(err)
	}
	url, err := pachsql.ParseURL(dsn)
	require.NoError(t, err)
	return url, password
}

// NewSnowSQL creates an emphermeral database in a real Snowflake instance.
func NewSnowSQL(t testing.TB) *sqlx.DB {
	db, _ := NewEphemeralSnowflakeDB(t)
	return db
}

func NewEphemeralSnowflakeDB(t testing.TB) (*sqlx.DB, string) {
	name := testutil.GenerateEphemeralDBName(t)
	url, password := getURLAndPassword(t)
	db := testutil.OpenDBURL(t, *url, password)
	ctx := context.Background()
	log := logrus.StandardLogger()
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	require.NoError(t, dbutil.WaitUntilReady(ctx, log, db))

	testutil.CreateEphemeralDB(t, db, name)
	url.Database = name
	url.Schema = "public"

	return testutil.OpenDBURL(t, *url, password), name
}
