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
	SnowflakeComputingDomain = ".snowflakecomputing.com"
)

// DB is an alias for sqlx.DB which is the standard database type used throughout the project
type DB = sqlx.DB

// Tx is an alias for sqlx.Tx which is the standard transaction type used throughout the project
type Tx = sqlx.Tx

// Stmt is an alias for sqlx.Stmt which is the standard prepared statement type used throught the project
type Stmt = sqlx.Stmt

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
	case ProtocolSnowflake:
		driver = "snowflake"
		dsn, err = snowflakeDSN(u, password)
	default:
		return nil, errors.Errorf("database protocol %q not supported", u.Protocol)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate DSN: %v", u)
	}
	res, err := sqlx.Open(driver, dsn)
	return res, errors.EnsureStack(err)
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

func snowflakeDSN(u URL, password string) (string, error) {
	// The Snowflake driver supports the following syntaxes:
	// * username[:password]@<account_identifier>/dbname/schemaname[?param1=value&...&paramN=valueN]
	// * username[:password]@<account_identifier>/dbname[?param1=value&...&paramN=valueN]
	// * username[:password]@hostname:port/dbname/schemaname?account=<account_identifier>[&param1=value&...&paramN=valueN]

	// A Snowflake account_identifier uniquely identifies a Snowflake account.
	// account_identifer can be embedded in the hostname, e.g. <account_identifier>.snowflakecomputing.com
	// however, the "snowflakecomputing.com" can be left out
	// example: jsmith@my_organization-my_account/mydb/testschema?warehouse=mywh
	// in this case, the account_identifier is my_organization-my_account
	params := make(map[string]*string, len(u.Params))
	for k, v := range u.Params {
		params[k] = &v
	}
	var account, host string
	if u.Port == 0 {
		// note sf.DSN will automatically set port to 443
		account = strings.TrimSuffix(u.Host, SnowflakeComputingDomain)
	} else if u.Host != "" && params["account"] != nil {
		host = u.Host
		account = *params["account"]
		delete(params, "account")
	}

	cfg := &sf.Config{
		Account:  account,
		User:     u.User,
		Password: password,
		Database: u.Database,
		Schema:   u.Schema,
		Host:     host,
		Port:     int(u.Port),
		Params:   params,
	}
	dsn, err := sf.DSN(cfg)
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	return dsn, nil
}

func copyParams(x map[string]string) map[string]string {
	y := make(map[string]string, len(x))
	for k, v := range x {
		y[k] = v
	}
	return y
}
