package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

// CheckClusterIsAuthorizedInTransaction returns an error if the current user doesn't have
// the permissions in `p` on the cluster
func (a *apiServer) CheckClusterIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, p ...auth.Permission) error {
	me, err := a.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Permissions: p}
	resp, err := a.AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_CLUSTER}, Required: p}
	}
	return nil
}

// CheckRepoIsAuthorizedInTransaction is identical to CheckRepoIsAuthorized except that
// it performs reads consistent with the latest state of the STM transaction.
func (a *apiServer) CheckRepoIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, r string, p ...auth.Permission) error {
	me, err := a.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Permissions: p}
	resp, err := a.AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Required: p}
	}
	return nil
}

// CheckRepoIsAuthorized returns an error if the current user doesn't have
// the permissions in `p` on the repo `r`
func (a *apiServer) CheckRepoIsAuthorized(ctx context.Context, r string, p ...auth.Permission) error {
	me, err := a.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Permissions: p}
	resp, err := a.Authorize(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_REPO, Name: r}, Required: p}
	}
	return nil
}

// CheckClusterIsAuthorized returns an error if the current user doesn't have
// the permissions in `p` on the cluster
func (a *apiServer) CheckClusterIsAuthorized(ctx context.Context, p ...auth.Permission) error {
	me, err := a.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Resource: &auth.Resource{Type: auth.ResourceType_CLUSTER}, Permissions: p}
	resp, err := a.Authorize(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: auth.Resource{Type: auth.ResourceType_CLUSTER}, Required: p}
	}
	return nil
}
