package pachsql

import (
	"net"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	sf "github.com/snowflakedb/gosnowflake"
)

const (
	ProtocolPostgres  = "postgres"
	ProtocolMySQL     = "mysql"
	ProtocolSnowflake = "snowflake"
)

// DB is an alias for sqlx.DB which is the standard database type used throughout the project
type DB = sqlx.DB

// Tx is an alias for sqlx.Tx which is the standard transaction type used throughout the project
type Tx = sqlx.Tx

// OpenURL returns a database connection pool to the database specified by u
// If password != "" then it will be used for authentication.
// This function does not confirm that the database is reachable; callers may be interested in pachsql.DB.Ping()
func OpenURL(u URL, password string) (*DB, error) {
	var driver string
	var dsn string
	switch u.Protocol {
	case ProtocolPostgres, "postgresql":
		driver = "pgx"
		dsn = postgresDSN(u, password)
	case ProtocolMySQL:
		driver = "mysql"
		dsn = mySQLDSN(u, password)
	case ProtocolSnowflake, "sf":
		driver = "snowflake"
		dsn = snowflakeDSN(u, password)
	default:
		return nil, errors.Errorf("database protocol %q not supported", u.Protocol)
	}
	res, err := sqlx.Open(driver, dsn)
	return res, errors.EnsureStack(err)
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
	params := copyParams(u.Params)
	params["parseTime"] = "true"
	config := mysql.Config{
		User:                 u.User,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 net.JoinHostPort(u.Host, strconv.Itoa(int(u.Port))),
		DBName:               u.Database,
		Params:               params,
		AllowNativePasswords: true,
	}
	return config.FormatDSN()
}

func snowflakeDSN(u URL, password string) string {
	// Snowflake account name is usually embedded in the host name
	// e.g <account_identifier>.snowflakecomputing.com
	var account string
	if strings.HasSuffix(u.Host, "snowflakecomputing.com") {
		account = strings.Split(u.Host, ".")[0]
	} else {
		account = u.Params["account"]
	}

	cfg := &sf.Config{
		Account:  account,
		User:     u.User,
		Password: password,
		Database: u.Database,
		Host:     u.Host,
		Port:     int(u.Port),
	}
	dsn, _ := sf.DSN(cfg)
	return dsn
}

func copyParams(x map[string]string) map[string]string {
	y := make(map[string]string, len(x))
	for k, v := range x {
		y[k] = v
	}
	return y
}
