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

const (
	SnowflakeCoreDomain = ".snowflakecomputing.com"
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
	case ProtocolSnowflake:
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
	// A Snowflake account_identifier uniquely identifies a Snowflake account.
	// account_identifer is embedded in the hostname, eg <account_identifier>.snowflakecomputing.com
	// however, the "snowflakecomputing.com" can be left out
	// example: jsmith@my_organization-my_account/mydb/testschema?warehouse=mywh
	// in this case, the account_identifier is my_organization-my_account

	// The only required fields are account, user, password
	// note: host and port are deprecated by Snowflake
	cfg := &sf.Config{
		Account:   strings.TrimSuffix(u.Host, SnowflakeCoreDomain),
		User:      u.User,
		Password:  password,
		Database:  u.Database,
		Schema:    u.Schema,
		Warehouse: u.Params["warehouse"],
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
