package setupenv

import (
	"time"

	"github.com/dlmiddlecote/sqlstats"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/prometheus/client_golang/prometheus"
)

func openDirectDB(config pachconfig.PostgresConfiguration) (*pachsql.DB, error) {
	db, err := dbutil.NewDB(
		dbutil.WithHostPort(config.PostgresHost, config.PostgresPort),
		dbutil.WithDBName(config.PostgresDBName),
		dbutil.WithUserPassword(config.PostgresUser, config.PostgresPassword),
		dbutil.WithMaxOpenConns(config.PostgresMaxOpenConns),
		dbutil.WithMaxIdleConns(config.PostgresMaxIdleConns),
		dbutil.WithConnMaxLifetime(time.Duration(config.PostgresConnMaxLifetimeSeconds)*time.Second),
		dbutil.WithConnMaxIdleTime(time.Duration(config.PostgresConnMaxIdleSeconds)*time.Second),
		dbutil.WithSSLMode(config.PostgresSSL),
		dbutil.WithQueryLog(config.PostgresQueryLogging, "pgx.direct"),
	)
	if err != nil {
		return nil, err
	}
	if err := prometheus.Register(sqlstats.NewStatsCollector("direct", db)); err != nil {
		return nil, errors.Errorf("problem registering stats collector for direct db client: %w", err)
	}
	return db, nil
}
