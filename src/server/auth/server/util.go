package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// CheckClusterIsAuthorizedInTransaction returns an error if the current user doesn't have
// the permissions in `p` on the cluster
func (a *apiServer) CheckClusterIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, p ...auth.Permission) error {
	return a.checkResourceIsAuthorizedInTransaction(txnCtx, &auth.Resource{Type: auth.ResourceType_CLUSTER}, p...)
}

// CheckProjectIsAuthorizedInTransaction returns an error if the current user doesn't have the permissions in `p` on the project.
// Projects inherit access controls from its parent cluster. Therefore, behind the scene, we check the cluster first, followed by this project.
func (a *apiServer) CheckProjectIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, project *pfs.Project, p ...auth.Permission) error {
	return a.checkResourceIsAuthorizedInTransaction(txnCtx, &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project.Name}, p...)
}

// CheckRepoIsAuthorizedInTransaction is identical to CheckRepoIsAuthorized except that
// it performs reads consistent with the latest state of the STM transaction.
// Repos inherit access controls from its parent project, which in turn inherits from the cluster.
// Therefore, behind the scene, we check the cluster first, then this repo's project, and finally this repo.
func (a *apiServer) CheckRepoIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, p ...auth.Permission) error {
	return a.checkResourceIsAuthorizedInTransaction(txnCtx, repo.AuthResource(), p...)
}

// CheckResourceIsAuthorizedInTransaction returns an error if the subject/user doesn't have permission in `p` on the `resource`
func (a *apiServer) checkResourceIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, resource *auth.Resource, p ...auth.Permission) error {
	me, err := txnCtx.WhoAmI()
	if err != nil {
		if errors.Is(err, auth.ErrNotActivated) {
			return nil
		}
		return err
	}

	req := &auth.AuthorizeRequest{Resource: resource, Permissions: p}
	resp, err := a.AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: resource, Required: p}
	}
	return nil
}

// CheckClusterIsAuthorized returns an error if the current user doesn't have
// the permissions in `p` on the cluster
func (a *apiServer) CheckClusterIsAuthorized(ctx context.Context, p ...auth.Permission) error {
	return a.checkResourceIsAuthorized(ctx, &auth.Resource{Type: auth.ResourceType_CLUSTER}, p...)
}

// CheckProjectIsAuthorized returns an error if the current user doesn't have permissiosn in `p` on `project`
func (a *apiServer) CheckProjectIsAuthorized(ctx context.Context, project *pfs.Project, p ...auth.Permission) error {
	return a.checkResourceIsAuthorized(ctx, &auth.Resource{Type: auth.ResourceType_PROJECT, Name: project.Name}, p...)
}

// CheckRepoIsAuthorized returns an error if the current user doesn't have
// the permissions in `p` on `repo`
func (a *apiServer) CheckRepoIsAuthorized(ctx context.Context, repo *pfs.Repo, p ...auth.Permission) error {
	return a.checkResourceIsAuthorized(ctx, repo.AuthResource(), p...)
}

func (a *apiServer) checkResourceIsAuthorized(ctx context.Context, resource *auth.Resource, p ...auth.Permission) error {
	me, err := a.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if err != nil {
		if errors.Is(err, auth.ErrNotActivated) {
			return nil
		}
		return err
	}

	req := &auth.AuthorizeRequest{Resource: resource, Permissions: p}
	resp, err := a.Authorize(ctx, req)
	if err != nil {
		return err
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Resource: resource, Required: p}
	}
	return nil
}
