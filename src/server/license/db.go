package license

import (
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// CreateClustersTable sets up the postgres table which tracks active clusters
func CreateClustersTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS license.clusters (
	id VARCHAR(4096) PRIMARY KEY,
	address VARCHAR(4096) NOT NULL,
	secret VARCHAR(64) NOT NULL,
	version VARCHAR(64) NOT NULL,
	auth_enabled BOOL NOT NULL,
	last_heartbeat TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`)
	return errors.EnsureStack(err)
}

func AddUserContextsToClustersTable(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
	ALTER TABLE license.clusters
	ADD COLUMN cluster_deployment_id VARCHAR(4096),
	ADD COLUMN user_address VARCHAR(4096),
	ADD COLUMN is_enterprise_server BOOLEAN
	;`)
	return err
}
