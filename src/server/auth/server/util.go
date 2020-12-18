package server

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// CheckIsAuthorizedInTransaction is identicalto CheckIsAuthorized except that
// it performs reads consistent with the latest state of the STM transaction.
func CheckIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, r *pfs.Repo, s auth.Scope) error {
	fmt.Println("cheque")

	me, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Repo: r.Name, Scope: s}
	resp, err := txnCtx.Auth().AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}

// CheckIsAuthorized returns an error if the current user (in 'pachClient') has
// authorization scope 's' for repo 'r'
func CheckIsAuthorized(pachClient *client.APIClient, r *pfs.Repo, s auth.Scope) error {
	fmt.Println("check")
	ctx := pachClient.Ctx()
	me, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Repo: r.Name, Scope: s}
	resp, err := pachClient.AuthAPIClient.Authorize(ctx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}
