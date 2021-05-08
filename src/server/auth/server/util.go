package server

import (
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
)

// CheckClusterIsAuthorizedInTransaction returns an error if the current user doesn't have
// the permissions in `p` on the cluster
func CheckClusterIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, p ...auth.Permission) error {
	me, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Permissions: p}
	resp, err := txnCtx.Auth().AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on cluster")
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_CLUSTER}, Required: p}
	}
	return nil
}

// CheckRepoIsAuthorizedInTransaction is identical to CheckRepoIsAuthorized except that
// it performs reads consistent with the latest state of the STM transaction.
func CheckRepoIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, r string, p ...auth.Permission) error {
	me, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Permissions: p}
	resp, err := txnCtx.Auth().AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on repo \"%s\"", r)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Required: p}
	}
	return nil
}

// CheckRepoIsAuthorized returns an error if the current user doesn't have
// the permissions in `p` on the repo `r`
func CheckRepoIsAuthorized(pachClient *client.APIClient, r string, p ...auth.Permission) error {
	ctx := pachClient.Ctx()
	me, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Permissions: p}
	resp, err := pachClient.AuthAPIClient.Authorize(ctx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on repo \"%s\"", r)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Required: p}
	}
	return nil
}
