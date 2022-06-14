package pachsql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"

	"github.com/DATA-DOG/go-sqlmock"
)

const (
	collections = "collections"
	columnName  = "column_name"
	commits     = "commits"
	commitDiffs = "commit_diffs"
	key         = "key"
	mockDBName  = "sqlmock"
	pfs         = "pfs"
	schemaName  = "schemaname"
	tableName   = "tablename"
)

var (
	commitsTable     = &SchemaTable{collections, commits}
	commitDiffsTable = &SchemaTable{pfs, commitDiffs}
)

func TestListTables(t *testing.T) {
	t.Run("ListTables() with a valid response", listTablesValidResponse)
	t.Run("ListTables() with an unexpected response", listTablesUnexpectedResponse)
	t.Run("ListTables() with an error from the database", listTablesInvalidResponse)
}

func listTablesValidResponse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{schemaName, tableName}).
		AddRow(collections, commits).
		AddRow(pfs, commitDiffs)
	mock.ExpectQuery(listTablesQuery).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	schemaTables, err := ListTables(context.Background(), sqlxDB)
	require.NoError(t, err, "test should pass")
	require.Equal(t, 2, len(schemaTables),
		"ListTables() should return same number of rows as input")
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
	require.Equal(t, true,
		err != nil && strings.HasPrefix(err.Error(), "could not scan tables:"),
		"expected an error from ListTables()")
}

func listTablesInvalidResponse(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	_, err = ListTables(context.Background(), sqlxDB)
	require.Equal(t, true, err != nil && strings.HasPrefix(err.Error(), "could not list tables:"),
		"expected an error from ListTables()")
}

func TestGetScannedRows(t *testing.T) {
	t.Run("GetScannedRows() with a valid response", getScannedRowsValidResponse)
	t.Run("GetScannedRows() with a valid response", getScannedRowsInvalidResponse)
}

func getScannedRowsValidResponse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{schemaName, tableName}).
		AddRow(collections, commits).
		AddRow(pfs, commitDiffs)
	mock.ExpectQuery("").WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	rows, err := GetScannedRows(context.Background(), sqlxDB, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))
	require.Equal(t, rows[0][schemaName], "collections")
}

func getScannedRowsInvalidResponse(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	_, err = GetScannedRows(context.Background(), sqlxDB, "")
	require.Equal(t, true, err != nil && strings.HasPrefix(err.Error(), "failed to query database:"),
		"expected an error from GetScannedRows()")
}

func TestKeyColumnExists(t *testing.T) {
	t.Run("KeyColumnExists() returns true", keyColumnExistsTrue)
	t.Run("KeyColumnExists() returns false", keyColumnExistsFalse)
}

func keyColumnExistsTrue(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{columnName}).AddRow(key)
	mock.ExpectQuery(fmt.Sprintf(keyColumnExistsQueryTemplate, collections, commits)).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	found, err := KeyColumnExists(context.Background(), sqlxDB, commitsTable)
	require.NoError(t, err)
	require.Equal(t, true, found)
}

func keyColumnExistsFalse(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{columnName})
	mock.ExpectQuery(fmt.Sprintf(keyColumnExistsQueryTemplate, pfs, commitDiffs)).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	found, err := KeyColumnExists(context.Background(), sqlxDB, commitDiffsTable)
	require.NoError(t, err)
	require.Equal(t, false, found)
}

func TestGetMainColumn(t *testing.T) {
	t.Run("GetMainColumn() when key exists", getMainColumnKeyExists)
	t.Run("GetMainColumn() when key doesn't exist", getMainColumnKeyDoesntExist)
}

func getMainColumnKeyExists(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockRows := sqlmock.NewRows([]string{columnName}).AddRow(key)
	mock.ExpectQuery(fmt.Sprintf(keyColumnExistsQueryTemplate, collections, commits)).WillReturnRows(mockRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	col, err := GetMainColumn(context.Background(), sqlxDB, commitsTable)
	require.NoError(t, err)
	require.Equal(t, key, col)
}

func getMainColumnKeyDoesntExist(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()
	mockKeyColumnRows := sqlmock.NewRows([]string{columnName})
	mock.ExpectQuery(fmt.Sprintf(keyColumnExistsQueryTemplate, pfs, commitDiffs)).WillReturnRows(mockKeyColumnRows)
	mockOrdinalColumnRows := sqlmock.NewRows([]string{columnName}).AddRow("commit_id")
	mock.ExpectQuery(fmt.Sprintf(getOrdinalColumnQueryTemplate, pfs, commitDiffs)).WillReturnRows(mockOrdinalColumnRows)
	sqlxDB := sqlx.NewDb(mockDB, mockDBName)
	col, err := GetMainColumn(context.Background(), sqlxDB, commitDiffsTable)
	require.NoError(t, err)
	require.Equal(t, "commit_id", col)
}
