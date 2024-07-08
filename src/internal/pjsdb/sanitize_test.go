package pjsdb_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pjsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func trySanitize(t *testing.T, d dependencies, parent pjsdb.JobID, mutate func(req *pjsdb.CreateJobRequest)) error {
	_, err := makeReq(t, d, parent, mutate).Sanitize(d.ctx, d.tx, d.s)
	if err != nil {
		return err
	}
	return nil
}

func TestCreateJobRequest_Sanitize(t *testing.T) {
	t.Run("invalid/spec/nil", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			err := trySanitize(t, d, 0, func(req *pjsdb.CreateJobRequest) {
				(*req).Spec = ""
			})
			require.YesError(t, err)
			if !errors.As(err, &pjsdb.InvalidFilesetError{}) {
				t.Fatalf("expected invalid fileset error, got: %s", err)
			}
		})
	})
	t.Run("invalid/spec/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			err := trySanitize(t, d, createRootJob(t, d), func(req *pjsdb.CreateJobRequest) {
				id := [16]byte{}
				_, err := rand.Read(id[:])
				require.NoError(t, err, "should be able to generate fileset id")
				req.Spec = hex.EncodeToString(id[:])
				req.Spec = ""
			})
			require.YesError(t, err)
			if !errors.As(err, &pjsdb.InvalidFilesetError{}) {
				t.Fatalf("expected invalid fileset error, got: %s", err)
			}
		})
	})
	t.Run("valid/input/nil", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			err := trySanitize(t, d, createRootJob(t, d), func(req *pjsdb.CreateJobRequest) {
				req.Input = ""
			})
			require.NoError(t, err)
		})
	})
	t.Run("invalid/input/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			err := trySanitize(t, d, createRootJob(t, d), func(req *pjsdb.CreateJobRequest) {
				id := [16]byte{}
				_, err := rand.Read(id[:])
				require.NoError(t, err, "should be able to generate fileset id")
				(*req).Input = hex.EncodeToString(id[:])
			})
			require.YesError(t, err)
			if !errors.As(err, &pjsdb.InvalidFilesetError{}) {
				t.Fatalf("expected invalid fileset error, got: %s", err)
			}
		})
	})
	t.Run("invalid/parent/not_exists", func(t *testing.T) {
		withDependencies(t, func(d dependencies) {
			err := trySanitize(t, d, pjsdb.JobID(1000), nil)
			require.YesError(t, err)
			if !errors.As(err, pjsdb.ErrParentNotFound) {
				t.Fatalf("expected parent not found error, got: %s", err)
			}
		})
	})

}
