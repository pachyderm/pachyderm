package server

import (
	"context"
	"database/sql"

	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validatable interface {
	ValidatePFS(pfs.LiveValidator) error
}

// ValidateRequest performs live validation on PFS messages.  This indirects through newValidator so
// that RPC methods don't have to care about the somewhat-complicated implementation.  (The
// complexity comes from avoiding an apiServer and txncontext depedency inside the PFS protos.)
func (a *apiServer) ValidateRequest(ctx context.Context, txnCtx *txncontext.TransactionContext, req validatable) error {
	var tx *pachsql.Tx
	if txnCtx != nil {
		tx = txnCtx.SqlTx
	} else {
		// Not everything that needs to validate quite follows the convention of having X
		// call XInTransaction; work around that.
		var err error
		tx, err = a.env.DB.BeginTxx(ctx, &sql.TxOptions{
			ReadOnly: true,
		})
		if err != nil {
			return errors.Wrap(err, "validate request: StartTxx")
		}
		defer tx.Rollback()
	}
	err := req.ValidatePFS(a.newValidator(ctx, tx))
	if err == nil {
		return nil
	}

	// Accumulate all the field violations.
	var fieldViolations []*errdetails.BadRequest_FieldViolation
	var visit func(err error)
	visit = func(err error) {
		var lvErr *pfs.LiveValidationError
		if errors.As(err, &lvErr) {
			fieldViolations = append(fieldViolations, &errdetails.BadRequest_FieldViolation{
				Field:       lvErr.Path,
				Description: lvErr.Err.Error(),
			})
			return
		}
		if x, ok := err.(interface{ Unwrap() error }); ok {
			visit(x.Unwrap())
			return
		}
		if x, ok := err.(interface{ Unwrap() []error }); ok {
			for _, err := range x.Unwrap() {
				visit(err)
			}
		}
	}
	visit(err)

	// Finally return a detailed RPC error.
	s := status.New(codes.FailedPrecondition, "validate request: "+err.Error())
	s, _ = s.WithDetails(&errdetails.BadRequest{
		FieldViolations: fieldViolations,
	})
	return s.Err()
}

// validator remembers the context and transaction context so the PFS package doesn't have to depend
// on txncontext.
type validator struct {
	ctx context.Context
	tx  *pachsql.Tx
	d   *driver
}

// Ensure that a validator is a pfs.Validator.
var _ pfs.LiveValidator = new(validator)

// ValidateRepoExists implements pfs.LiveValidator.
func (v *validator) ValidateRepoExists(repo *pfs.Repo) error {
	if err := v.ValidateProjectExists(repo.GetProject()); err != nil {
		return errors.EnsureStack(err)
	}

	if err := pfsdb.RepoExistsByName(v.ctx, v.tx, repo.GetProject().GetName(), repo.GetName(), repo.GetType()); err != nil {
		return err
	}

	return nil
}

// ValidateProjectExists implements pfs.LiveValidator.
func (v *validator) ValidateProjectExists(project *pfs.Project) error {
	if _, err := coredb.GetProjectByName(v.ctx, v.tx, project.GetName()); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

// newValidator creates a validator.  We use indirect though ValidateRequest so that code in this
// package can't easily hold on to a validator{} and abuse the embedded context.
//
// txnCtx may be nil if a non-transactional collections read is acceptable.
func (a *apiServer) newValidator(ctx context.Context, tx *pachsql.Tx) *validator {
	return &validator{ctx: pctx.Child(ctx, "validate"), tx: tx, d: a.driver}
}
