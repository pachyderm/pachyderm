package server

import (
	"context"
	"math"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// CommitEvent is an event that contains a CommitInfo or an error
type CommitEvent struct {
	Err   error
	Value *pfs.CommitInfo
}

// CommitStream is a stream of CommitInfos
type CommitStream interface {
	Stream() <-chan CommitEvent
	Close()
}

func newUserCommitInfo(txnCtx *txncontext.TransactionContext, branch *pfs.Branch) *pfs.CommitInfo {
	log.Info(pctx.TODO(), "creating commit", zap.Stack("stack"), zap.String("username", txnCtx.Username()), zap.Stringer("branch", branch), zap.String("id", txnCtx.CommitSetID))
	return &pfs.CommitInfo{
		Commit: &pfs.Commit{
			Branch: branch,
			Repo:   branch.Repo,
			Id:     txnCtx.CommitSetID,
		},
		Origin:    &pfs.CommitOrigin{Kind: pfs.OriginKind_USER},
		Started:   txnCtx.Timestamp,
		Details:   &pfs.CommitInfo_Details{},
		CreatedBy: txnCtx.Username(),
	}
}

// only transform source repos and spouts get a closed commit
func (a *apiServer) makeEmptyCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, directProvenance []*pfs.Branch, parent *pfs.Commit) (*pfsdb.Commit, error) {
	// Input repos want a closed head commit, so decide if we leave
	// it open by the presence of branch provenance.
	closed := true
	if len(directProvenance) > 0 {
		closed = false
	}
	commitHandle := branch.NewCommit(txnCtx.CommitSetID)
	commitHandle.Repo = branch.Repo
	commitInfo := &pfs.CommitInfo{
		Commit:    commitHandle,
		Origin:    &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
		Started:   txnCtx.Timestamp,
		CreatedBy: txnCtx.Username(),
	}
	if closed {
		commitInfo.Finishing = txnCtx.Timestamp
		commitInfo.Finished = txnCtx.Timestamp
		commitInfo.Details = &pfs.CommitInfo_Details{}
	}
	commitId, err := a.addCommitInfoToDB(ctx, txnCtx, commitInfo, parent, directProvenance, false /* needsFinishedParent */)
	if err != nil {
		return nil, err
	}
	commit := &pfsdb.Commit{
		ID:         commitId,
		CommitInfo: commitInfo,
	}
	if closed {
		total, err := a.storage.Filesets.ComposeTx(txnCtx.SqlTx, nil, defaultTTL)
		if err != nil {
			return nil, err
		}
		if err := a.commitStore.SetTotalFileSetTx(txnCtx.SqlTx, commit, *total); err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	return commit, nil
}

// creates a new commit, and adds both commit ancestry, and commit provenance pointers
//
// NOTE: Requiring source commits to have finishing / finished parents ensures that the commits are not compacted
// in a pathological order (finishing later commits before earlier commits will result with us compacting
// the earlier commits multiple times).
func (a *apiServer) addCommitInfoToDB(ctx context.Context, txnCtx *txncontext.TransactionContext, newCommitInfo *pfs.CommitInfo, parent *pfs.Commit, directProvenance []*pfs.Branch, needsFinishedParent bool) (pfsdb.CommitID, error) {
	if err := a.linkParent(ctx, txnCtx, newCommitInfo, parent, needsFinishedParent); err != nil {
		return 0, err
	}
	for _, prov := range directProvenance {
		b, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, prov)
		if err != nil {
			if pfsdb.IsNotFoundError(err) {
				return 0, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: prov})
			}
			return 0, errors.Wrap(err, "add commit")
		}
		newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, b.BranchInfo.Head)
	}
	commitID, err := pfsdb.CreateCommit(ctx, txnCtx.SqlTx, newCommitInfo)
	if err != nil {
		if errors.As(err, &pfsdb.CommitAlreadyExistsError{}) {
			return 0, errors.Join(err, pfsserver.ErrInconsistentCommit{Commit: newCommitInfo.Commit})
		}
		return 0, errors.EnsureStack(err)
	}
	for _, p := range newCommitInfo.DirectProvenance {
		if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommitInfo.Commit, p); err != nil {
			return 0, err
		}
	}
	return commitID, nil
}

// startCommit makes a new commit in 'branch', with the parent 'parent':
//   - 'parent' may be omitted, in which case the parent commit is inferred
//     from 'branch'.
//   - If 'parent' is set, it determines the parent commit, but 'branch' is
//     still moved to point at the new commit
func (a *apiServer) startCommit(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	parent *pfs.Commit,
	branch *pfs.Branch,
	description string,
) (*pfsdb.Commit, error) {
	if err := a.validateStartCommitArgs(ctx, txnCtx, branch); err != nil {
		return nil, err
	}
	newCommitInfo := newUserCommitInfo(txnCtx, branch)
	newCommitInfo.Description = description
	commitID, branchInfo, err := a.createCommitOnBranch(ctx, txnCtx, newCommitInfo, parent, branch)
	if err != nil {
		return nil, errors.Wrap(err, "create commit")
	}
	// check if this is happening in a spout pipeline, and alias the spec commit
	_, ok1 := os.LookupEnv(client.PPSPipelineNameEnv)
	_, ok2 := os.LookupEnv("PPS_SPEC_COMMIT")
	if !(ok1 && ok2) && len(branchInfo.Provenance) > 0 {
		// Otherwise, we don't allow user code to start commits on output branches
		return nil, pfsserver.ErrCommitOnOutputBranch{Branch: branch}
	}
	// Defer propagation of the commit until the end of the transaction, so we can
	// batch downstream commits together if there are multiple changes.
	if err := txnCtx.PropagateBranch(branch); err != nil {
		return nil, err
	}
	return &pfsdb.Commit{ID: commitID, CommitInfo: newCommitInfo}, nil
}

func (a *apiServer) validateStartCommitArgs(ctx context.Context, txnCtx *txncontext.TransactionContext, branchHandle *pfs.Branch) error {
	// Validate arguments:
	if branchHandle == nil || branchHandle.Name == "" {
		return errors.Errorf("branch must be specified")
	}
	// Check that caller is authorized
	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, branchHandle.Repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.EnsureStack(err)
	}
	// Check if repo exists
	_, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, branchHandle.Repo.Project.Name, branchHandle.Repo.Name, branchHandle.Repo.Type)
	if err != nil {
		if pfsdb.IsErrRepoNotFound(err) {
			return pfsserver.ErrRepoNotFound{Repo: branchHandle.Repo}
		}
		return errors.EnsureStack(err)
	}
	if err := ancestry.ValidateName(branchHandle.Name); err != nil {
		return err
	}
	return nil
}

// createCommitOnBranch creates a commit and also creates a branch if needed.
// it handles updating the commit's 'branch' field after the branch is resolved.
func (a *apiServer) createCommitOnBranch(
	ctx context.Context,
	txnCtx *txncontext.TransactionContext,
	newCommitInfo *pfs.CommitInfo,
	parent *pfs.Commit,
	branchHandle *pfs.Branch,
) (pfsdb.CommitID, *pfs.BranchInfo, error) {
	branchInfo, err := getOrDefaultBranchInfo(ctx, txnCtx, branchHandle)
	if err != nil {
		return 0, nil, errors.Wrap(err, "create commit")
	}
	// If the parent is unspecified, use the current head of the branch.
	if parent == nil {
		parent = branchInfo.Head
	}
	commitID, err := a.addCommitInfoToDB(ctx, txnCtx, newCommitInfo, parent, branchInfo.DirectProvenance, true)
	if err != nil {
		return 0, nil, err
	}
	branchInfo.Head = newCommitInfo.Commit
	if err := upsertBranchAndSyncCommit(ctx, txnCtx.SqlTx, branchInfo, commitID); err != nil {
		return 0, nil, errors.Wrap(err, "create commit")
	}
	return commitID, branchInfo, nil
}

// getOrDefaultBranchInfo attempts to get the branchInfo from the database referenced by branchHandle.
// the purpose of this function is to improve readability for apiServer.addCommitInfoToDB().
func getOrDefaultBranchInfo(ctx context.Context, txnCtx *txncontext.TransactionContext, branchHandle *pfs.Branch) (branchInfo *pfs.BranchInfo, err error) {
	b, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branchHandle)
	if err != nil {
		if !pfsdb.IsNotFoundError(err) {
			return nil, errors.Wrap(err, "get or default branchHandle info")
		}
		branchInfo = &pfs.BranchInfo{
			CreatedBy: txnCtx.Username(),
			Branch:    branchHandle,
		}
	}
	if b != nil && b.BranchInfo != nil { // The branch exists already in postgres.
		branchInfo = b.BranchInfo
	}
	return branchInfo, nil
}

// Set child.ParentCommit (if 'parent' has been determined).
func (a *apiServer) linkParent(ctx context.Context, txnCtx *txncontext.TransactionContext, child *pfs.CommitInfo, parent *pfs.Commit, needsFinishedParent bool) error {
	if parent == nil {
		return nil
	}
	// Resolve 'parent' if it's a branch that isn't 'branch' (which can
	// happen if 'branch' is new and diverges from the existing branch in
	// 'parent').
	parentCommit, err := a.resolveCommitTx(ctx, txnCtx.SqlTx, parent)
	if err != nil {
		return errors.Wrapf(err, "parent commit not found")
	}
	parentCommitInfo := parentCommit.CommitInfo
	// fail if the parent commit has not been finished
	if needsFinishedParent && parentCommitInfo.Finishing == nil {
		return errors.Errorf("parent commit %s has not been finished", parentCommitInfo.Commit)
	}
	child.ParentCommit = parentCommitInfo.Commit
	return nil
}

func (a *apiServer) finishCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commit *pfsdb.Commit, description, commitError string, force bool) error {
	commitInfo := commit.CommitInfo
	if commitInfo.Finishing != nil {
		return pfsserver.ErrCommitFinished{
			Commit: commitInfo.Commit,
		}
	}
	if !force && len(commitInfo.DirectProvenance) > 0 {
		if info, err := a.env.GetPipelineInspector().InspectPipelineInTransaction(ctx, txnCtx, pps.RepoPipeline(commitInfo.Commit.Repo)); err != nil && !errutil.IsNotFoundError(err) {
			return errors.EnsureStack(err)
		} else if err == nil && info.Type == pps.PipelineInfo_PIPELINE_TYPE_TRANSFORM {
			return errors.Errorf("cannot finish a pipeline output or meta commit, use 'stop job' instead")
		}
		// otherwise, this either isn't a pipeline at all, or is a spout or service for which we should allow finishing
	}
	if description != "" {
		if err := pfsdb.UpdateDescription(ctx, txnCtx.SqlTx, commit.ID, description); err != nil {
			return errors.Wrap(err, "finish commit")
		}
	}
	return pfsdb.FinishingCommit(ctx, txnCtx.SqlTx, commit.ID, txnCtx.Timestamp, commitError)
}

func (a *apiServer) findCommits(ctx context.Context, request *pfs.FindCommitsRequest, cb func(response *pfs.FindCommitsResponse) error) error {
	commit := request.Start
	foundCommits, commitsSearched := uint32(0), uint32(0)
	searchDone := false
	var found *pfs.Commit
	makeResp := func(found *pfs.Commit, commitsSearched uint32, lastSearchedCommit *pfs.Commit) *pfs.FindCommitsResponse {
		resp := &pfs.FindCommitsResponse{}
		if found != nil {
			resp.Result = &pfs.FindCommitsResponse_FoundCommit{FoundCommit: found}
		} else {
			resp.Result = &pfs.FindCommitsResponse_LastSearchedCommit{LastSearchedCommit: commit}
		}
		resp.CommitsSearched = commitsSearched
		return resp
	}
	for {
		if searchDone {
			return errors.EnsureStack(cb(makeResp(nil, commitsSearched, commit)))
		}
		logFields := []zap.Field{
			zap.String("commit", commit.Id),
			zap.String("repo", commit.Repo.String()),
			zap.String("target", request.FilePath),
		}
		if commit.Branch != nil {
			logFields = append(logFields, zap.String("branch", commit.Branch.String()))
		}
		if err := log.LogStep(ctx, "searchingCommit", func(ctx context.Context) error {
			inspectCommitResp, err := a.resolveCommitWithAuth(ctx, commit)
			if err != nil {
				return err
			}
			if inspectCommitResp.Finished == nil {
				return pfsserver.ErrCommitNotFinished{Commit: commit}
			}
			commit = inspectCommitResp.Commit
			inCommit, err := a.isPathModifiedInCommit(ctx, inspectCommitResp, request.FilePath)
			if err != nil {
				return err
			}
			if inCommit {
				log.Info(ctx, "found target", zap.String("commit", commit.Id), zap.String("repo", commit.Repo.String()), zap.String("target", request.FilePath))
				found = commit
				foundCommits++
				if err := cb(makeResp(found, commitsSearched, nil)); err != nil {
					return err
				}
			}
			commitsSearched++
			if (foundCommits == request.Limit && request.Limit != 0) || inspectCommitResp.ParentCommit == nil {
				searchDone = true
				return nil
			}
			commit = inspectCommitResp.ParentCommit
			return nil
		}, logFields...); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return errors.EnsureStack(cb(makeResp(nil, commitsSearched, commit)))
			}
			return errors.EnsureStack(err)
		}
	}
}

func (a *apiServer) isPathModifiedInCommit(ctx context.Context, commit *pfsdb.Commit, filePath string) (bool, error) {
	diffID, err := a.getCompactedDiffFileSet(ctx, commit)
	if err != nil {
		return false, err
	}
	diffFileSet, err := a.storage.Filesets.Open(ctx, []fileset.ID{*diffID})
	if err != nil {
		return false, err
	}
	found := false
	pathOption := []index.Option{
		index.WithRange(&index.PathRange{
			Lower: filePath,
		}),
	}
	if err = diffFileSet.Iterate(ctx, func(file fileset.File) error {
		if file.Index().Path == filePath {
			found = true
			return errutil.ErrBreak
		}
		return nil
	}, pathOption...); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return false, err
	}
	// we don't care about the file operation, so if a file was already found, skip iterating over deletive set.
	if found {
		return found, nil
	}
	if err = diffFileSet.IterateDeletes(ctx, func(file fileset.File) error {
		if file.Index().Path == filePath {
			found = true
			return errutil.ErrBreak
		}
		return nil
	}, pathOption...); err != nil && !errors.Is(err, errutil.ErrBreak) {
		return false, err
	}
	return found, nil
}

func (a *apiServer) waitForCommit(ctx context.Context, commitHandle *pfs.Commit, wait pfs.CommitState) (*pfsdb.Commit, error) {
	commit, err := a.resolveCommitWithAuth(ctx, commitHandle)
	if err != nil {
		return nil, err
	}
	if commit.Finished == nil {
		switch wait {
		case pfs.CommitState_STARTED:
		case pfs.CommitState_READY:
			for _, c := range commit.DirectProvenantIDs {
				var provCommit *pfsdb.Commit
				if err := a.txnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
					var err error
					provCommit, err = pfsdb.GetCommit(ctx, txnCtx.SqlTx, c)
					return err
				}); err != nil {
					return nil, errors.Wrap(err, "inspect commit")
				}
				if _, err := a.waitForFinishingOrFinished(ctx, provCommit, pfs.CommitState_FINISHED); err != nil {
					return nil, err
				}
			}
		case pfs.CommitState_FINISHING, pfs.CommitState_FINISHED:
			return a.waitForFinishingOrFinished(ctx, commit, wait)
		}
	}
	return commit, nil
}

// waitForFinishingOrFinished waits for the commit to be FINISHING or FINISHED.
func (a *apiServer) waitForFinishingOrFinished(ctx context.Context, commit *pfsdb.Commit, wait pfs.CommitState) (*pfsdb.Commit, error) {
	// We only cancel the watcher if we detect the commit is the right state.
	expectedErr := errors.New("commit is in the right state")
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	if err := pfsdb.WatchCommit(ctx, a.env.DB, a.env.Listener, commit.ID,
		func(c pfsdb.Commit) error {
			switch wait {
			case pfs.CommitState_FINISHING:
				if c.Finishing != nil {
					commit.CommitInfo = c.CommitInfo
					cancel(expectedErr)
					return nil
				}
			case pfs.CommitState_FINISHED:
				if c.Finished != nil {
					commit.CommitInfo = c.CommitInfo
					cancel(expectedErr)
					return nil
				}
			}
			return nil
		},
		func(id pfsdb.CommitID) error {
			return pfsserver.ErrCommitDeleted{Commit: commit.Commit}
		},
	); err != nil && !errors.Is(context.Cause(ctx), expectedErr) {
		return nil, errors.Wrap(err, "inspect finishing or finished commit")
	}
	return commit, nil
}

func (a *apiServer) resolveCommitWithAuth(ctx context.Context, commitHandle *pfs.Commit) (*pfsdb.Commit, error) {
	if err := a.env.Auth.CheckRepoIsAuthorized(ctx, commitHandle.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return a.resolveCommit(ctx, commitHandle)
}

// resolveCommit creates a transaction, then calls resolveCommitTx in that transaction.
func (a *apiServer) resolveCommit(ctx context.Context, commitHandle *pfs.Commit) (*pfsdb.Commit, error) {
	// Resolve the commit in case it specifies a branch head or commit ancestry
	var commit *pfsdb.Commit
	if err := a.txnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		commit, err = a.resolveCommitTx(ctx, txnCtx.SqlTx, commitHandle)
		return err
	}); err != nil {
		return nil, err
	}
	return commit, nil
}

// resolveCommitTx contains the essential implementation of InspectCommit. it converts 'commit' (which may
// be a commit ID or branch reference, plus '~' and/or '^') to a repo + commit
// ID. It accepts a postgres transaction so that it can be used in a transaction
// and avoids an inconsistent call to a.waitForCommit()
func (a *apiServer) resolveCommitTx(ctx context.Context, sqlTx *pachsql.Tx, commitHandle *pfs.Commit) (*pfsdb.Commit, error) {
	if commitHandle == nil {
		return nil, errors.Errorf("cannot resolve nil commit")
	}
	if commitHandle.Repo == nil {
		return nil, errors.Errorf("cannot resolve commit with no repo")
	}
	if commitHandle.Id == "" && commitHandle.GetBranch().GetName() == "" {
		return nil, errors.Errorf("cannot resolve commit with no ID or branch")
	}
	if commitHandle.AccessRepo().Name == fileSetsRepo {
		cinfo := &pfs.CommitInfo{
			Commit:      commitHandle,
			Description: "FileSet - Virtual Commit",
			Finished:    &timestamppb.Timestamp{}, // it's always been finished. How did you get the id if it wasn't finished?
		}
		return &pfsdb.Commit{
			ID:         0, // this doesn't seem like the right thing to do, but here we are.
			CommitInfo: cinfo,
			Revision:   0,
		}, nil
	}
	commitHandleCopy := proto.Clone(commitHandle).(*pfs.Commit) // back up user commit, for error reporting
	// Extract any ancestor tokens from 'commit.ID' (i.e. ~, ^ and .)
	var ancestryLength int
	var err error
	commitHandleCopy.Id, ancestryLength, err = ancestry.Parse(commitHandleCopy.Id)
	if err != nil {
		return nil, err
	}
	// Now that ancestry has been parsed out, check if the ID is a branch name
	if commitHandleCopy.Id != "" && !uuid.IsUUIDWithoutDashes(commitHandleCopy.Id) {
		if commitHandleCopy.Branch.GetName() != "" {
			return nil, errors.Errorf("invalid commit ID given with a branch (%s): %s\n", commitHandleCopy.Branch, commitHandleCopy.Id)
		}
		commitHandleCopy.Branch = commitHandleCopy.Repo.NewBranch(commitHandleCopy.Id)
		commitHandleCopy.Id = ""
	}
	// If commit.ID is unspecified, get it from the branch head
	if commitHandleCopy.Id == "" {
		branchInfo, err := pfsdb.GetBranch(ctx, sqlTx, commitHandleCopy.Branch)
		if err != nil {
			if pfsdb.IsNotFoundError(err) {
				return nil, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: commitHandleCopy.Branch})
			}
			return nil, errors.Wrap(err, "resolve commit")
		}
		commitHandleCopy.Id = branchInfo.Head.Id
	}
	commit, err := pfsdb.GetCommitByKey(ctx, sqlTx, commitHandleCopy)
	if err != nil {
		if pfsdb.IsNotFoundError(err) {
			// try to resolve to alias if not found
			resolvedCommit, err := pfsdb.ResolveCommitProvenance(sqlTx, commitHandle.Repo, commitHandleCopy.Id)
			if err != nil {
				return nil, err
			}
			commitHandleCopy.Id = resolvedCommit.Id
			commit, err = pfsdb.GetCommitByKey(ctx, sqlTx, commitHandleCopy)
			if err != nil {
				return nil, errors.EnsureStack(err)
			}
		} else {
			return nil, errors.Wrap(err, "resolve commit")
		}
	}
	// Traverse commits' parents until you've reached the right ancestor
	if ancestryLength >= 0 {
		for i := 1; i <= ancestryLength; i++ {
			if commit.CommitInfo.ParentCommit == nil {
				return nil, pfsserver.ErrCommitNotFound{Commit: commitHandle}
			}
			parent := commit.CommitInfo.ParentCommit
			commit, err = pfsdb.GetCommitByKey(ctx, sqlTx, parent)
			if err != nil {
				if pfsdb.IsNotFoundError(err) {
					if i == 0 {
						return nil, errors.Join(err, pfsserver.ErrCommitNotFound{Commit: commitHandle})
					}
					return nil, errors.Join(err, pfsserver.ErrParentCommitNotFound{Commit: commitHandleCopy})
				}
				return nil, errors.EnsureStack(err)
			}
		}
	} else {
		cis := make([]*pfsdb.Commit, ancestryLength*-1)
		for i := 0; ; i++ {
			if commitHandleCopy == nil {
				if i >= len(cis) {
					commit = cis[i%len(cis)]
					break
				}
				return nil, pfsserver.ErrCommitNotFound{Commit: commitHandle}
			}
			cis[i%len(cis)], err = pfsdb.GetCommitByKey(ctx, sqlTx, commitHandleCopy)
			if err != nil {
				if pfsdb.IsNotFoundError(err) {
					if i == 0 {
						return nil, errors.Join(err, pfsserver.ErrCommitNotFound{Commit: commitHandle})
					}
					return nil, errors.Join(err, pfsserver.ErrParentCommitNotFound{Commit: commitHandleCopy})
				}
				return nil, err
			}
			commitHandleCopy = cis[i%len(cis)].CommitInfo.ParentCommit
		}
	}
	return commit, nil
}

// passesCommitOriginFilter is a helper function for listCommit and
// subscribeCommit to apply filtering to the returned commits.  By default
// we allow users to request all the commits with
// 'all', or a specific type of commit with 'originKind'.
func passesCommitOriginFilter(commitInfo *pfs.CommitInfo, all bool, originKind pfs.OriginKind) bool {
	if all {
		return true
	}
	if originKind != pfs.OriginKind_ORIGIN_KIND_UNKNOWN {
		return commitInfo.Origin.Kind == originKind
	}
	return true
}

func (a *apiServer) listCommit(
	ctx context.Context,
	repo *pfs.Repo,
	to *pfs.Commit,
	from *pfs.Commit,
	startTime *timestamppb.Timestamp,
	number int64,
	reverse bool,
	all bool,
	originKind pfs.OriginKind,
	cb func(*pfs.CommitInfo) error,
) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	if err := a.env.Auth.CheckRepoIsAuthorized(ctx, repo, auth.Permission_REPO_LIST_COMMIT); err != nil {
		return errors.EnsureStack(err)
	}
	if from != nil && !proto.Equal(from.Repo, repo) || to != nil && !proto.Equal(to.Repo, repo) {
		return errors.Errorf("`from` and `to` commits need to be from repo %s", repo)
	}
	// Make sure that the repo exists
	if repo.Name != "" {
		if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
			if _, err := pfsdb.GetRepoByName(ctx, tx, repo.Project.Name, repo.Name, repo.Type); err != nil {
				if pfsdb.IsErrRepoNotFound(err) {
					return pfsserver.ErrRepoNotFound{Repo: repo}
				}
				return errors.EnsureStack(err)
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "ensure repo exists")
		}
	}

	// Make sure that both from and to are valid commits
	if from != nil {
		if _, err := a.resolveCommit(ctx, from); err != nil {
			return err
		}
	}
	if to != nil {
		ci, err := a.resolveCommit(ctx, to)
		if err != nil {
			return err
		}
		to = ci.Commit
	}

	// if number is 0, we return all commits that match the criteria
	if number == 0 {
		number = math.MaxInt64
	}

	if from != nil && to == nil {
		return errors.Errorf("cannot use `from` commit without `to` commit")
	} else if from == nil && to == nil {
		// we hold onto a revisions worth of cis so that we can sort them by provenance
		var cis []*pfsdb.Commit
		// sendCis sorts cis and passes them to f
		sendCis := func() error {
			// We don't sort these because there is no provenance between commits
			// within a repo, so there is no topological sort necessary.
			for i, ci := range cis {
				if number == 0 {
					return errutil.ErrBreak
				}
				number--

				if reverse {
					ci = cis[len(cis)-1-i]
				}
				var err error
				ci.SizeBytesUpperBound, err = a.commitSizeUpperBound(ctx, ci)
				if err != nil && !pfsserver.IsBaseCommitNotFinishedErr(err) {
					return err
				}
				if err := cb(ci.CommitInfo); err != nil {
					return err
				}
			}
			cis = nil
			return nil
		}

		// if neither from and to is given, we list all commits in
		// the repo, sorted by revision timestamp.
		var filter *pfs.Commit
		if repo.Name != "" {
			filter = &pfs.Commit{Repo: repo}
		}
		// driver.listCommit should return more recent commits by default, which is the
		// opposite behavior of pfsdb.ForEachCommit.
		order := pfsdb.SortOrderDesc
		if reverse {
			order = pfsdb.SortOrderAsc
		}
		lastRev := int64(-1)
		if err := pfsdb.ForEachCommit(ctx, a.env.DB, filter, func(commit pfsdb.Commit) error {
			ci := commit
			if ci.Revision != lastRev {
				if err := sendCis(); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
				lastRev = ci.Revision
			}
			if passesCommitOriginFilter(ci.CommitInfo, all, originKind) {
				commitInfo := proto.Clone(ci.CommitInfo).(*pfs.CommitInfo)
				if startTime != nil {
					createdAt := time.Unix(ci.Started.GetSeconds(), int64(ci.Started.GetNanos())).UTC()
					fromTime := time.Unix(startTime.GetSeconds(), int64(startTime.GetNanos())).UTC()
					if !reverse && createdAt.Before(fromTime) || reverse && createdAt.After(fromTime) {
						cis = append(cis, &pfsdb.Commit{CommitInfo: commitInfo, Revision: ci.Revision, ID: ci.ID})
					}
					return nil
				}
				cis = append(cis, &pfsdb.Commit{CommitInfo: commitInfo, Revision: ci.Revision, ID: ci.ID})
			}
			return nil
		}, pfsdb.OrderByCommitColumn{Column: pfsdb.CommitColumnID, Order: order}); err != nil {
			return errors.Wrap(err, "list commit")
		}
		// Call sendCis one last time to send whatever's pending in 'cis'
		if err := sendCis(); err != nil && !errors.Is(err, errutil.ErrBreak) {
			return err
		}
	} else {
		if reverse {
			return errors.Errorf("cannot use 'Reverse' while also using 'From' or 'To'")
		}
		cursor := to
		for number != 0 && cursor != nil && (from == nil || cursor.Id != from.Id) {
			var commit *pfsdb.Commit
			var err error
			if err := dbutil.WithTx(ctx, a.env.DB, func(cbCtx context.Context, tx *pachsql.Tx) error {
				commit, err = pfsdb.GetCommitByKey(ctx, tx, cursor)
				return err
			}); err != nil {
				return errors.Wrap(err, "list commit")
			}
			if passesCommitOriginFilter(commit.CommitInfo, all, originKind) {
				var err error
				commit.SizeBytesUpperBound, err = a.commitSizeUpperBound(ctx, commit)
				if err != nil && !pfsserver.IsBaseCommitNotFinishedErr(err) {
					return err
				}
				if err := cb(commit.CommitInfo); err != nil {
					if errors.Is(err, errutil.ErrBreak) {
						return nil
					}
					return err
				}
				number--
			}
			cursor = commit.ParentCommit
		}
	}
	return nil
}

func (a *apiServer) subscribeCommit(
	ctx context.Context,
	repo *pfs.Repo,
	branch string,
	from *pfs.Commit,
	state pfs.CommitState,
	all bool,
	originKind pfs.OriginKind,
	cb func(*pfs.CommitInfo) error,
) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}
	if from != nil && !proto.Equal(from.Repo, repo) {
		return errors.Errorf("the `from` commit needs to be from repo %s", repo)
	}
	// keep track of the commits that have been sent
	seen := make(map[string]bool)
	var repoID pfsdb.RepoID
	var err error
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		repoID, err = pfsdb.GetRepoID(ctx, tx, repo.Project.Name, repo.Name, repo.Type)
		return errors.Wrap(err, "get repo ID")
	}); err != nil {
		return err
	}
	return pfsdb.WatchCommitsInRepo(ctx, a.env.DB, a.env.Listener, repoID,
		func(c pfsdb.Commit) error { // onUpsert
			// if branch is provided, make sure the commit was created on that branch
			if branch != "" && c.Commit.Branch.Name != branch {
				return nil
			}
			// If the origin of the commit doesn't match what we're interested in, skip it
			if !passesCommitOriginFilter(c.CommitInfo, all, originKind) {
				return nil
			}
			// We don't want to include the `from` commit itself
			if !(seen[c.Commit.Id] || (from != nil && from.Id == c.Commit.Id)) {
				// Wait for the commit to enter the right state
				commit, err := a.waitForCommit(ctx, proto.Clone(c.Commit).(*pfs.Commit), state)
				if err != nil {
					return err
				}
				commitInfo := commit.CommitInfo
				if err := cb(commitInfo); err != nil {
					return err
				}
				seen[commitInfo.Commit.Id] = true
			}
			return nil
		},
		func(id pfsdb.CommitID) error { // onDelete
			return nil
		},
	)
}

func (a *apiServer) clearCommit(ctx context.Context, commitHandle *pfs.Commit) error {
	commit, err := a.resolveCommit(ctx, commitHandle)
	if err != nil {
		return err
	}
	if commit.Finishing != nil {
		return errors.Errorf("cannot clear finished commit")
	}
	return errors.EnsureStack(a.commitStore.DropFileSets(ctx, commit))
}

func (a *apiServer) squashCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commitHandle *pfs.Commit, recursive bool) error {
	commit, err := a.resolveCommitTx(ctx, txnCtx.SqlTx, commitHandle)
	if err != nil {
		return err
	}
	commitHandle = commit.Commit
	subvenantCommits, err := pfsdb.GetFullCommitSubvenance(ctx, txnCtx.SqlTx, commitHandle)
	if err != nil {
		return err
	}
	commits := []*pfsdb.Commit{commit}
	for _, c := range subvenantCommits {
		ci, err := pfsdb.GetCommitByKey(ctx, txnCtx.SqlTx, c)
		if err != nil {
			return err
		}
		commits = append(commits, ci)
	}
	if len(commits) > 1 && !recursive {
		return errors.Errorf("cannot squash commit (%v) with subvenance without recursive", commitHandle)
	}
	for _, c := range commits {
		if c.Commit.Branch.Repo.Type == pfs.SpecRepoType && c.Origin.Kind == pfs.OriginKind_USER {
			return errors.Errorf("cannot squash commit %s because it updated a pipeline", c.Commit)
		}
		if len(c.ChildCommits) == 0 {
			return &pfsserver.ErrSquashWithoutChildren{Commit: c.Commit}
		}
	}
	for _, c := range commits {
		if err := a.deleteCommit(ctx, txnCtx, c); err != nil {
			return err
		}
		txnCtx.StopJob(c.Commit)
	}
	return nil
}

func (a *apiServer) dropCommit(ctx context.Context, txnCtx *txncontext.TransactionContext, commitHandle *pfs.Commit, recursive bool) error {
	commit, err := a.resolveCommitTx(ctx, txnCtx.SqlTx, commitHandle)
	if err != nil {
		return err
	}
	commitHandle = commit.Commit
	subvenantCommits, err := pfsdb.GetFullCommitSubvenance(ctx, txnCtx.SqlTx, commitHandle)
	if err != nil {
		return err
	}
	commits := []*pfsdb.Commit{commit}
	for _, subvenantCommit := range subvenantCommits {
		c, err := pfsdb.GetCommitByKey(ctx, txnCtx.SqlTx, subvenantCommit)
		if err != nil {
			return err
		}
		commits = append(commits, c)
	}
	if len(commits) > 1 && !recursive {
		return errors.Errorf("cannot drop commit (%v) with subvenance without recursive", commitHandle)
	}
	for _, ci := range commits {
		if len(ci.ChildCommits) > 0 {
			return &pfsserver.ErrDropWithChildren{Commit: ci.Commit}
		}
	}
	for _, ci := range commits {
		if err := a.deleteCommit(ctx, txnCtx, ci); err != nil {
			return err
		}
		txnCtx.StopJob(ci.Commit)
	}
	return nil
}

func (a *apiServer) walkCommitProvenanceTx(ctx context.Context, txnCtx *txncontext.TransactionContext, request *WalkCommitProvenanceRequest,
	startId pfsdb.CommitID, cb func(commitInfo *pfs.CommitInfo) error) error {
	commits, err := pfsdb.GetProvenantCommits(ctx, txnCtx.SqlTx, startId,
		pfsdb.WithMaxDepth(request.MaxDepth), pfsdb.WithLimit(request.MaxCommits))
	if err != nil {
		return errors.Wrap(err, "walk commit provenance in transaction")
	}
	for _, commit := range commits {
		if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, commit.Commit.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
			return errors.EnsureStack(err)
		}
		if err := cb(commit.CommitInfo); err != nil {
			return errors.Wrap(err, "walk commit provenance in transaction")
		}
	}
	return nil
}

func (a *apiServer) walkCommitSubvenanceTx(ctx context.Context, txnCtx *txncontext.TransactionContext, request *WalkCommitSubvenanceRequest,
	startId pfsdb.CommitID, cb func(commitInfo *pfs.CommitInfo) error) error {
	commits, err := pfsdb.GetSubvenantCommits(ctx, txnCtx.SqlTx, startId,
		pfsdb.WithMaxDepth(request.MaxDepth), pfsdb.WithLimit(request.MaxCommits))
	if err != nil {
		return errors.Wrap(err, "walk commit subvenance in transaction")
	}
	for _, commit := range commits {
		if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, commit.Commit.Repo, auth.Permission_REPO_INSPECT_COMMIT); err != nil {
			return errors.EnsureStack(err)
		}
		if err := cb(commit.CommitInfo); err != nil {
			return errors.Wrap(err, "walk commit subvenance in transaction")
		}
	}
	return nil
}

func (a *apiServer) openCommit(ctx context.Context, commitHandle *pfs.Commit) (*pfs.CommitInfo, fileset.FileSet, error) {
	if commitHandle.AccessRepo() == nil {
		return nil, nil, errors.New("nil repo or branch.repo in commit")
	}
	if commitHandle.AccessRepo().Name == fileSetsRepo {
		fsid, err := fileset.ParseID(commitHandle.Id)
		if err != nil {
			return nil, nil, err
		}
		fs, err := a.storage.Filesets.Open(ctx, []fileset.ID{*fsid})
		if err != nil {
			return nil, nil, err
		}
		return &pfs.CommitInfo{Commit: commitHandle}, fs, nil
	}
	if err := a.env.Auth.CheckRepoIsAuthorized(ctx, commitHandle.Repo, auth.Permission_REPO_READ); err != nil {
		return nil, nil, errors.EnsureStack(err)
	}
	commit, err := a.resolveCommit(ctx, commitHandle)
	if err != nil {
		return nil, nil, err
	}
	if commit.Finishing != nil && commit.Finished == nil {
		_, err := a.waitForCommit(ctx, commitHandle, pfs.CommitState_FINISHED)
		if err != nil {
			return nil, nil, err
		}
	}
	id, err := a.getFileset(ctx, commit)
	if err != nil {
		return nil, nil, err
	}
	fs, err := a.storage.Filesets.Open(ctx, []fileset.ID{*id})
	if err != nil {
		return nil, nil, err
	}
	return commit.CommitInfo, fs, nil
}
