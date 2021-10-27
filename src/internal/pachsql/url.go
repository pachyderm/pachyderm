package pachsql

import (
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// URL contains the information needed to connect to a SQL database, except for the password.
type URL struct {
	Protocol string
	User     string
	Host     string
	Port     uint16
	Database string
	Params   map[string]string
}

// ParseURL attempts to parse x into a URL
func ParseURL(x string) (*URL, error) {
	u, err := url.Parse(x)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}
	params := make(map[string]string)
	for k, v := range u.Query() {
		if len(v) > 0 {
			params[k] = v[len(v)-1]
		}
	}
	return &URL{
		Protocol: u.Scheme,
		Host:     u.Hostname(),
		Port:     uint16(port),
		User:     u.User.Username(),
		Database: strings.Trim(u.Path, "/"),
		Params:   params,
	}, nil
}

func (u *URL) String() string {
	return (&url.URL{
		Scheme: u.Protocol,
		Host:   u.Host,
		User:   url.UserPassword(u.User, ""),
		Path:   u.Database,
	}).String()
}

// OpenDBFromURL returns a database connection pool to the database specified by u
// If password != "" then it will be used for authentication.
// This function does not confirm that the database is reachable; callers may be interested in sqlx.DB.Ping()
func OpenDBFromURL(u URL, password string) (*sqlx.DB, error) {
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
