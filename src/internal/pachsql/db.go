package pachsql

import (
	"context"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"go.uber.org/zap"
)

const (
	ProtocolPostgres = "postgres"
	ProtocolMySQL    = "mysql"
)

const (
	listTablesQuery = `
		SELECT schemaname, tablename
		FROM pg_catalog.pg_tables
		WHERE schemaname != 'pg_catalog'
		  AND schemaname != 'information_schema'
		ORDER BY schemaname, tablename;
	`
)

var fixMysqlLoggerOnce sync.Once

// DB is an alias for sqlx.DB which is the standard database type used throughout the project
type DB = sqlx.DB

// Tx is an alias for sqlx.Tx which is the standard transaction type used throughout the project
type Tx = sqlx.Tx

// Stmt is an alias for sqlx.Stmt which is the standard prepared statement type used throught the project
type Stmt = sqlx.Stmt

// SchemaTable stores a given table's name and schema.
type SchemaTable struct {
	SchemaName string `json:"schemaname"`
	TableName  string `json:"tablename"`
}

// RowMap is an alias for map[string]interface{} which is the type used by sqlx.MapScan()
type RowMap = map[string]interface{}

// OpenURL returns a database connection pool to the database specified by u
// If password != "" then it will be used for authentication.
// This function does not confirm that the database is reachable; callers may be interested in pachsql.DB.Ping()
func OpenURL(u URL, password string) (*DB, error) {
	var err error
	var driver string
	var dsn string
	switch u.Protocol {
	case ProtocolPostgres, "postgresql":
		driver = "pgx"
		dsn, err = postgresDSN(u, password)
	case ProtocolMySQL:
		driver = "mysql"
		dsn, err = mySQLDSN(u, password)
		fixMysqlLoggerOnce.Do(func() {
			l := zap.NewStdLog(zap.L().Named("mysql"))
			l.Println("enabled global mysql logger")
			mysql.SetLogger(l) //nolint:errcheck
		})
	default:
		return nil, errors.Errorf("database protocol %q not supported", u.Protocol)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate DSN: %v", u)
	}
	res, err := sqlx.Open(driver, dsn)
	return res, errors.EnsureStack(err)
}

// ListTables returns an array of SchemaTable structs that represent the tables.
func ListTables(ctx context.Context, db *DB) ([]SchemaTable, error) {
	var tables []SchemaTable
	if err := sqlx.SelectContext(ctx, db, &tables, listTablesQuery); err != nil {
		return nil, errors.Wrap(err, "list tables")
	}
	return tables, nil
}

func postgresDSN(u URL, password string) (string, error) {
	if u.Schema != "" {
		return "", errors.New("postgres DSN should not contain schema name")
	}
	port := u.Port
	if port == 0 {
		port = 5432
	}
	fields := map[string]string{
		"user":   u.User,
		"host":   u.Host,
		"port":   strconv.Itoa(int(port)),
		"dbname": u.Database,
	}
	if password != "" {
		fields["password"] = password
	}
	for k, v := range u.Params {
		fields[k] = v
	}
	var dsnParts []string
	for k, v := range fields {
		dsnParts = append(dsnParts, k+"="+v)
	}
	return strings.Join(dsnParts, " "), nil
}

func mySQLDSN(u URL, password string) (string, error) {
	if u.Schema != "" {
		return "", errors.New("mysql DSN should not contain schema name")
	}
	port := u.Port
	if port == 0 {
		port = 3306
	}
	params := copyParams(u.Params)
	params["parseTime"] = "true"
	config := mysql.Config{
		User:                 u.User,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(u.Host, strconv.Itoa(int(port))),
		DBName:               u.Database,
		Params:               params,
		AllowNativePasswords: true,
	}
	return config.FormatDSN(), nil
}

func copyParams(x map[string]string) map[string]string {
	y := make(map[string]string, len(x))
	for k, v := range x {
		y[k] = v
	}
	return y
}

// ConnStringFromConn returns a postgres commandline -compatible connection string from the provided
// DB.  If is not a pgx connection, an error will be returned.
func ConnStringFromConn(ctx context.Context, db *DB) (result string, retErr error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return "", errors.Wrap(err, "create new connection")
	}
	defer errors.Close(&retErr, conn, "close tmp connection")
	if err := conn.Raw(func(dc any) error {
		sc, ok := dc.(*stdlib.Conn)
		if !ok {
			return errors.New("not a stdlib.Conn; this method only works on pgx connections")
		}
		var b strings.Builder
		for _, p := range strings.Split(sc.Conn().Config().ConnString(), " ") {
			// Remove non-Postgres options.  See pgx-specific options here:
			// https://pkg.go.dev/github.com/jackc/pgx/v4#ParseConfig
			switch {
			case strings.HasPrefix(p, "statement_cache_capacity="):
			case strings.HasPrefix(p, "statement_cache_mode="):
			case strings.HasPrefix(p, "prefer_simple_protocol="):
			default:
				b.WriteString(p)
				b.WriteRune(' ')
			}
		}
		result = strings.TrimSpace(b.String())
		return nil
	}); err != nil {
		return "", errors.Wrap(err, "attempt to extract connection string from raw connection")
	}
	return
}
