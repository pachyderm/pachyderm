package metadata

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/coredb"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/metadata"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type Auth interface {
	CheckClusterIsAuthorizedInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, p ...auth.Permission) error
	CheckRepoIsAuthorizedInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, p ...auth.Permission) error
}

// EditMetadataInTransaction transactionally mutates metadata.  All operations are attempted, in order, but if
// any fail, the entire operation fails.
func EditMetadataInTransaction(ctx context.Context, tx *txncontext.TransactionContext, auth Auth, req *metadata.EditMetadataRequest) error {
	var errs error
	for i, edit := range req.GetEdits() {
		if err := editInTx(ctx, tx, auth, edit); err != nil {
			errors.JoinInto(&errs, errors.Wrapf(err, "edit #%d", i))
		}
	}
	if errs != nil {
		return errs
	}
	return nil
}

func editMetadata(edit *metadata.Edit, md *map[string]string) error {
	if *md == nil {
		*md = make(map[string]string)
	}
	switch x := edit.GetOp().(type) {
	case *metadata.Edit_AddKey_:
		k, v := x.AddKey.Key, x.AddKey.Value
		if _, ok := (*md)[k]; ok {
			return errors.Errorf("add_key target key %q already exists; use edit_key instead", k)
		}
		(*md)[k] = v
	case *metadata.Edit_EditKey_:
		k, v := x.EditKey.Key, x.EditKey.Value
		(*md)[k] = v
	case *metadata.Edit_DeleteKey_:
		k := x.DeleteKey.Key
		delete(*md, k)
	case *metadata.Edit_Replace_:
		*md = x.Replace.Replacement
	}
	return nil
}

func editInTx(ctx context.Context, tc *txncontext.TransactionContext, authServer Auth, edit *metadata.Edit) error {
	switch x := edit.GetTarget().(type) {
	case *metadata.Edit_Project:
		p, err := pfsdb.PickProject(ctx, x.Project, tc.SqlTx)
		if err != nil {
			return errors.Wrap(err, "pick project")
		}
		// Auth rules: any authenticated user can edit project metadata; this is the same as
		// the rules for editing the description of a project.
		if err := editMetadata(edit, &p.Metadata); err != nil {
			return errors.Wrapf(err, "edit project %q", p.GetProject().GetName())
		}
		if err := pfsdb.UpdateProject(ctx, tc.SqlTx, p.ID, p.ProjectInfo); err != nil {
			return errors.Wrapf(err, "update project %q", p.GetProject().GetName())
		}
	case *metadata.Edit_Commit:
		c, err := pfsdb.PickCommit(ctx, x.Commit, tc.SqlTx)
		if err != nil {
			return errors.Wrap(err, "pick commit")
		}
		// Auth rules: any authenticated user can edit commit metadata; this is the same as
		// the rules for starting commits, finish commits, etc.
		if err := editMetadata(edit, &c.CommitInfo.Metadata); err != nil {
			return errors.Wrapf(err, "edit commit %q", c.GetCommit().Key())
		}
		if err := pfsdb.UpdateCommit(ctx, tc.SqlTx, c.ID, c.CommitInfo, pfsdb.AncestryOpt{
			SkipChildren: true,
			SkipParent:   true,
		}); err != nil {
			return errors.Wrapf(err, "update commit %q", c.GetCommit().Key())
		}
	case *metadata.Edit_Branch:
		b, err := pfsdb.PickBranch(ctx, x.Branch, tc.SqlTx)
		if err != nil {
			return errors.Wrap(err, "pick branch")
		}
		// Auth rules: users must have REPO_CREATE_BRANCH on the target repo to edit a
		// branch.  This is the same as updating the HEAD of an existing branch.
		if err := authServer.CheckRepoIsAuthorizedInTransaction(ctx, tc, b.GetBranch().GetRepo(), auth.Permission_REPO_CREATE_BRANCH); err != nil {
			return errors.Wrapf(err, "check permissions on branch %q of repo %q", b.GetBranch().Key(), b.GetBranch().GetRepo().Key())
		}
		if err := editMetadata(edit, &b.BranchInfo.Metadata); err != nil {
			return errors.Wrapf(err, "edit branch %q", b.GetBranch().Key())
		}
		if _, err := pfsdb.UpsertBranch(ctx, tc.SqlTx, b.BranchInfo); err != nil {
			return errors.Wrapf(err, "update branch %q", b.GetBranch().Key())
		}
	case *metadata.Edit_Repo:
		r, err := pfsdb.PickRepo(ctx, x.Repo, tc.SqlTx)
		if err != nil {
			return errors.Wrap(err, "pick repo")
		}
		// Auth rules: users must have REPO_WRITE to update metadata.  This is the same rule
		// as editing the description.
		if err := authServer.CheckRepoIsAuthorizedInTransaction(ctx, tc, r.GetRepo(), auth.Permission_REPO_WRITE); err != nil {
			return errors.Wrapf(err, "check permissions on repo %q", r.GetRepo().Key())
		}
		if err := editMetadata(edit, &r.RepoInfo.Metadata); err != nil {
			return errors.Wrapf(err, "edit repo %q", r.GetRepo().Key())
		}
		if _, err := pfsdb.UpsertRepo(ctx, tc.SqlTx, r.RepoInfo); err != nil {
			return errors.Wrapf(err, "update repo %q", r.GetRepo().Key())
		}
	case *metadata.Edit_Cluster:
		if err := authServer.CheckClusterIsAuthorizedInTransaction(ctx, tc, auth.Permission_CLUSTER_EDIT_CLUSTER_METADATA); err != nil {
			return errors.Wrap(err, "check cluster permissions")
		}
		md, err := coredb.GetClusterMetadata(ctx, tc.SqlTx)
		if err != nil {
			return errors.Wrap(err, "get cluster metadata")
		}
		if err := editMetadata(edit, &md); err != nil {
			return errors.Wrap(err, "edit cluster metadata")
		}
		if err := coredb.UpdateClusterMetadata(ctx, tc.SqlTx, md); err != nil {
			return errors.Wrap(err, "update cluster metadata")
		}
	default:
		return errors.Errorf("unknown target %v", edit.GetTarget())
	}
	return nil
}
