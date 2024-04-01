package coredb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
)

type clusterMetadataRow struct {
	OnlyOneRow bool                  `db:"onlyonerow"`
	Metadata   pgjsontypes.StringMap `db:"metadata"`
}

func GetClusterMetadata(ctx context.Context, tx *pachsql.Tx) (map[string]string, error) {
	var row clusterMetadataRow
	if err := tx.GetContext(ctx, &row, `select metadata from core.cluster_metadata limit 1`); err != nil {
		return nil, errors.Wrap(err, "select cluster metdata")
	}
	return row.Metadata.Data, nil
}

func UpdateClusterMetadata(ctx context.Context, tx *pachsql.Tx, metadata map[string]string) error {
	if _, err := tx.ExecContext(ctx, `update core.cluster_metadata set metadata=$1`, pgjsontypes.StringMap{Data: metadata}); err != nil {
		return errors.Wrap(err, "update cluster metadata")
	}
	return nil
}
