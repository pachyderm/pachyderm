// DO NOT MODIFY THIS STATE
// IT HAS ALREADY SHIPPED IN A RELEASE
package clusterstate

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

var state_2_5_4 migrations.State = state_2_5_0.
	Apply("Lengthen auth token column length for projects", func(ctx context.Context, env migrations.Env) error {
		if _, err := env.Tx.ExecContext(ctx, `ALTER TABLE auth.auth_tokens ALTER COLUMN subject TYPE varchar(128)`); err != nil {
			return errors.Wrap(err, "could not alter column subject in table auth.auth_tokens from varchar(64) to varchar(128)")
		}
		return nil
	}, migrations.Squash)
