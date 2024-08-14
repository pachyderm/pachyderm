package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Propagater is an object that is used to propagate PFS branches at the end of
// a transaction.  The transactionenv package provides the interface for this
// and will call the Run function at the end of a transaction.
type Propagater struct {
	a      *apiServer
	txnCtx *txncontext.TransactionContext

	// Branches that were modified (new commits or head commit was moved to an old commit)
	branches map[string]*pfs.Branch
}

func (a *apiServer) NewPropagater(txnCtx *txncontext.TransactionContext) txncontext.PfsPropagater {
	return &Propagater{
		a:        a,
		txnCtx:   txnCtx,
		branches: map[string]*pfs.Branch{},
	}
}

// PropagateBranch marks a branch as needing propagation once the transaction
// successfully ends.  This will be performed by the Run function.
func (t *Propagater) PropagateBranch(branch *pfs.Branch) error {
	if branch == nil {
		return errors.New("cannot propagate nil branch")
	}
	t.branches[pfsdb.BranchKey(branch)] = branch
	return nil
}

// DeleteBranch removes a branch from the list of those needing propagation
// if present.
func (t *Propagater) DeleteBranch(branch *pfs.Branch) {
	delete(t.branches, pfsdb.BranchKey(branch))
}

// Run performs any final tasks and cleanup tasks in the transaction, such as
// propagating branches
func (t *Propagater) Run(ctx context.Context) error {
	branches := make([]*pfs.Branch, 0, len(t.branches))
	for _, branch := range t.branches {
		branches = append(branches, branch)
	}
	if err := t.a.validateDAGStructure(ctx, t.txnCtx, branches); err != nil {
		return errors.Wrap(err, "validate DAG at end of transaction")
	}
	return t.a.propagateBranches(ctx, t.txnCtx, branches)
}

type RepoValidator struct {
	a      *apiServer
	txnCtx *txncontext.TransactionContext
	repos  map[string]*pfs.Repo
}

func (a *apiServer) NewRepoValidator(txnCtx *txncontext.TransactionContext) txncontext.PfsRepoValidator {
	return &RepoValidator{
		a:      a,
		txnCtx: txnCtx,
		repos:  map[string]*pfs.Repo{},
	}
}

func (rc *RepoValidator) ValidateRepo(repo *pfs.Repo) error {
	if repo == nil {
		return errors.New("cannot check branches in an empty repo")
	}
	rc.repos[pfsdb.RepoKey(repo)] = repo
	return nil
}

func (rc *RepoValidator) Run(ctx context.Context) error {
	for _, repo := range rc.repos {
		if err := rc.a.listBranchInTransaction(ctx, rc.txnCtx, repo, false, func(bi *pfs.BranchInfo) error {
			head, err := pfsdb.GetCommitByKey(ctx, rc.txnCtx.SqlTx, bi.Head)
			if err != nil {
				return errors.Wrap(err, "get commit by key")
			}
			// the branch head should not be a forgotten commit; only a finished commit can be forgotten
			if head.Finished != nil && head.Error == "" {
				_, err := rc.a.commitStore.GetTotalFileSetTx(rc.txnCtx.SqlTx, head)
				if err != nil {
					// a finished commit that has no total file set is forgotten
					if errors.Is(err, errNoTotalFileSet) {
						return errors.New("the branch head cannot be a forgotten commit")
					}
				}
				return errors.Wrap(err, "get total file set")
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
