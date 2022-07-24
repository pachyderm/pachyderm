package pachsql

import (
	"context"
	"reflect"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	collections = "collections"
	commits     = "commits"
	commitDiffs = "commit_diffs"
	mockDBName  = "sqlmock"
	pfs         = "pfs"
	schemaName  = "schemaname"
	tableName   = "tablename"
)

func TestListTables(t *testing.T) {
	t.Run("ListTables() with a valid response", listTablesValidResponse)
	t.Run("ListTables() with an unexpected response", listTablesUnexpectedResponse)
}

func listTablesValidResponse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	expected := []SchemaTable{
		{SchemaName: collections, TableName: commits},
		{SchemaName: pfs, TableName: commitDiffs},
	}
	mockRows := sqlmock.NewRows([]string{schemaName, tableName}).
		AddRow(expected[0].SchemaName, expected[0].TableName).
		AddRow(expected[1].SchemaName, expected[1].TableName)
	mock.ExpectQuery(listTablesQuery).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	schemaTables, err := ListTables(context.Background(), sqlxDB)
	require.NoError(t, err, "test should pass")
	require.Equal(t, true, reflect.DeepEqual(expected, schemaTables),
		"ListTables() structTables should equal expected struct")
}

func listTablesUnexpectedResponse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{"food"}).
		AddRow("sushi").
		AddRow("cheeseburger")
	mock.ExpectQuery(listTablesQuery).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	_, err = ListTables(context.Background(), sqlxDB)
	require.YesError(t, err, "expected an error from ListTables()")
	require.Matches(t, "^list tables:.*", err.Error(), "expected an error from ListTables()")
}
