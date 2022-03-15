package pachsql

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// URL contains the information needed to connect to a SQL database, except for the password.
type URL struct {
	Protocol string
	User     string
	Host     string
	Port     uint16
	Database string
	Schema   string
	Params   map[string]string
}

// ParseURL attempts to parse x into a URL
func ParseURL(x string) (*URL, error) {
	u, err := url.Parse(x)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	var port int
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return nil, errors.EnsureStack(errors.Wrapf(err, "parsing url port: %w"))
		}
	}

	params := make(map[string]string)
	for k, v := range u.Query() {
		if len(v) > 0 {
			params[k] = v[len(v)-1]
		}
	}

	// parse database name and schema
	// assume /dbname/schemaname
	var dbname, schemaname string
	parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 2)
	dbname = parts[0]
	if len(parts) == 2 {
		schemaname = parts[1]
	}
	return &URL{
		Protocol: u.Scheme,
		Host:     u.Hostname(),
		Port:     uint16(port),
		User:     u.User.Username(),
		Database: dbname,
		Schema:   schemaname,
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
