package dbutil

import "time"

// Option configures a DB.
type Option func(*dbConfig)

// WithHostPort sets the host and port for the DB.
func WithHostPort(host string, port int) Option {
	return func(dbc *dbConfig) {
		dbc.host = host
		dbc.port = port
	}
}

// WithUserPassword sets the user and password for the DB.
func WithUserPassword(user, password string) Option {
	return func(dbc *dbConfig) {
		dbc.user = user
		dbc.password = password
	}
}

// WithDBName sets the name for the DB.
func WithDBName(DBName string) Option {
	return func(dbc *dbConfig) {
		dbc.name = DBName
	}
}

// WithMaxOpenConns sets the maximum number of concurrent database connections
// to be allocated before blocking new acquisitions.
func WithMaxOpenConns(n int) Option {
	return func(dbc *dbConfig) {
		dbc.maxOpenConns = n
	}
}

// WithMaxIdleConns sets the maximum number of idle database connections to keep available for
// future queries.
func WithMaxIdleConns(n int) Option {
	return func(dbc *dbConfig) {
		dbc.maxIdleConns = n
	}
}

// WithConnMaxLifetime sets the maximum time a database connection may be reused for.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(dbc *dbConfig) {
		dbc.connMaxLifetime = d
	}
}

// WithConnMaxIdleTime sets the maximum time a database connection may be idle for.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(dbc *dbConfig) {
		dbc.connMaxIdleTime = d
	}
}

const (
	SSLModeDisable = "disable"

	DefaultSSLMode = SSLModeDisable
)

// WithSSLMode sets the SSL mode for connections to the database
func WithSSLMode(mode string) Option {
	return func(dbc *dbConfig) {
		dbc.sslMode = mode
	}
}
