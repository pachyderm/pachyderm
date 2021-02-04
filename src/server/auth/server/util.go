package server

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
)

// CheckIsAuthorizedInTransaction is identicalto CheckIsAuthorized except that
// it performs reads consistent with the latest state of the STM transaction.
func CheckIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, r *auth.Resource, p []auth.Permission) error {
	me, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: r, Permissions: p}
	resp, err := txnCtx.Auth().AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: *r, Required: p}
	}
	return nil
}

// CheckIsAuthorized returns an error if the current user (in 'pachClient') has
// authorization scope 's' for resource 'r'
func CheckIsAuthorized(pachClient *client.APIClient, r *auth.Resource, p []auth.Permission) error {
	ctx := pachClient.Ctx()
	me, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: r, Permissions: p}
	resp, err := pachClient.AuthAPIClient.Authorize(ctx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: *r, Required: p}
	}
	return nil
}
