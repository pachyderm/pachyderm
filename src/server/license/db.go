package license

import (
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// CreateClustersTable sets up the postgres table which tracks active clusters
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CreateClustersTableV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
CREATE TABLE license.clusters (
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

// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func AddUserContextsToClustersTable(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	ALTER TABLE license.clusters
	ADD COLUMN cluster_deployment_id VARCHAR(4096),
	ADD COLUMN user_address VARCHAR(4096),
	ADD COLUMN is_enterprise_server BOOLEAN
	;`)
	return errors.EnsureStack(err)
}

// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func AddClusterClientIdColumn(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	ALTER TABLE license.clusters
	ADD COLUMN client_id VARCHAR(4096)
	;`)
	return errors.EnsureStack(err)
}
