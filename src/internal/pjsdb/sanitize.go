package pjsdb

import (
	"context"
	"database/sql"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

// sanitize.go contains a sanitization pattern to decouple the CRUD API from *fileset.Storage and validate
// CreateJob() parameters.

// CreateJobRequest is a bundle of related fields required for a CreateJob() invocation.
// In pfsdb, we used a pattern of forcing the caller to convert their resources to a _Info struct, but its problematic since
// each database function only really needs a subset of the fields and it is not clear to the caller which fields are required.
// CreateJobRequest cannot be used directly, it has to be first sanitized via Sanitize()
type CreateJobRequest struct {
	Parent      JobID
	Input, Spec string
}

type createJobRequest struct {
	parent      sql.NullInt64
	input, spec string
	specHash    []byte
}

// Sanitize should be invoked prior to calling CreateJob() to ensure that a job can be properly created.
// this functionally is separate from the CreateJob() so that CreateJob() doesn't require access to a *fileset.Storage
func (req CreateJobRequest) Sanitize(ctx context.Context, tx *pachsql.Tx, s *fileset.Storage) (emptyReq createJobRequest, err error) {
	ctx = pctx.Child(ctx, "sanitize")
	specFileSet, err := validateFileset(ctx, s, req.Spec, true)
	if err != nil {
		return emptyReq, err
	}
	inputFileset := &validFileset{}
	// input can be nil.
	if req.Input != "" {
		inputFileset, err = validateFileset(ctx, s, req.Input, false)
		if err != nil {
			return emptyReq, err
		}
	}
	parent := sql.NullInt64{Valid: false}
	if req.Parent != 0 {
		parent.Int64 = int64(req.Parent)
		if _, err := GetJob(ctx, tx, req.Parent); err != nil {
			if errors.As(err, &JobNotFoundError{}) {
				return emptyReq, errors.Join(ErrParentNotFound, errors.Wrap(err, "sanitize: parent"))
			}
			return emptyReq, errors.Wrap(err, "sanitize: parent")
		}
		parent.Valid = true
	}
	return createJobRequest{
		input:    inputFileset.handle,
		spec:     specFileSet.handle,
		specHash: specFileSet.hash,
		parent:   parent,
	}, nil
}
