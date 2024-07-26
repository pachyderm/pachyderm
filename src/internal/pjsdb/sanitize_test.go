package pjsdb_test

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestCreateJobRequest_IsSanitized(t *testing.T) {
	t.Run("invalid/program/nil", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			req := pjsdb.CreateJobRequest{}
			err := req.IsSanitized(d.ctx, d.tx)
			require.YesError(t, err)
			require.ErrorContains(t, err, "program cannot be nil")
		})
	})
	t.Run("invalid/parent/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			req := pjsdb.CreateJobRequest{Parent: pjsdb.JobID(1000), Program: mockFileset(t, d, "/program", "")}
			err := req.IsSanitized(d.ctx, d.tx)
			require.YesError(t, err)
			if !errors.As(err, pjsdb.ErrParentNotFound) {
				t.Fatalf("expected parent not found error, got: %s", err)
			}
		})
	})
}
