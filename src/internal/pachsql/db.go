package pachsql

import (
	"net"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// DB is an alias for pachsql.DB which is the standard database type used throughout the project
type DB = sqlx.DB

// OpenURL returns a database connection pool to the database specified by u
// If password != "" then it will be used for authentication.
// This function does not confirm that the database is reachable; callers may be interested in pachsql.DB.Ping()
func OpenURL(u URL, password string) (*DB, error) {
	var driver string
	var dsn string
	switch u.Protocol {
	case "postgresql", "postgres":
		driver = "pgx"
		dsn = postgresDSN(u, password)
	case "mysql":
		driver = "mysql"
		dsn = mySQLDSN(u, password)
	default:
		return nil, errors.Errorf("database protocol %q not supported", u.Protocol)
	}
	return sqlx.Open(driver, dsn)
}

func postgresDSN(u URL, password string) string {
	fields := map[string]string{
		"user":   u.User,
		"host":   u.Host,
		"port":   strconv.Itoa(int(u.Port)),
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
	return strings.Join(dsnParts, " ")
}

func mySQLDSN(u URL, password string) string {
	config := mysql.Config{
		User:   u.User,
		Passwd: password,
		Addr:   net.JoinHostPort(u.Host, strconv.Itoa(int(u.Port))),
		DBName: u.Database,
		Params: u.Params,
	}
	return config.FormatDSN()
}
