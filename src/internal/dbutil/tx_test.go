package dbutil_test

import (
	"context"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"go.uber.org/zap"
)

func TestNestingTransactionsPanics(t *testing.T) {
	ctx := pctx.Child(pctx.TestContext(t), "", pctx.WithOptions(zap.Development()))
	db := dockertestenv.NewTestDB(t)
	var withTxErr error
	require.YesPanic(t, func() {
		withTxErr = dbutil.WithTx(ctx, db, func(ctx context.Context, _ *pachsql.Tx) error {
			err := dbutil.WithTx(ctx, db, func(_ context.Context, _ *pachsql.Tx) error {
				return nil
			})
			return errors.Wrap(err, "nested WithTx")
		})
	}, "nesting transactions should panic")
	require.NoError(t, withTxErr, "WithTx should not have errored")
}
