package server

import (
	"context"
	"github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/v2/src/internal/cronutil"
	"sort"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
)

// propagateBranches starts commits downstream of 'branches'
// in order to restore the invariant that branch provenance matches HEAD commit
// provenance:
//
//	B.Head is provenant on A.Head <=>
//	branch B is provenant on branch A
//
// The implementation assumes that the invariant already holds for all branches
// upstream of 'branches', but not necessarily for each 'branch' itself.
//
// In other words, propagateBranches scans all branches b_downstream that are
// equal to or downstream of 'branches', and if the HEAD of b_downstream isn't
// provenant on the HEADs of b_downstream's provenance, propagateBranches starts
// a new HEAD commit in b_downstream that is. For example, propagateBranches
// starts downstream output commits (which trigger PPS jobs) when new input
// commits arrive on 'branch', when 'branches's HEAD is deleted, or when
// 'branches' are newly created (i.e. in CreatePipeline).
func (a *apiServer) propagateBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branches []*pfs.Branch) error {
	if len(branches) == 0 {
		return nil
	}
	var propagatedBranches []*pfs.BranchInfo
	seen := make(map[string]*pfs.BranchInfo)
	for _, branchHandle := range branches {
		branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branchHandle)
		if err != nil {
			return errors.Wrapf(err, "propagate branches: get branchHandle %q", pfsdb.BranchKey(branchHandle))
		}
		for _, subvenantBranchHandle := range branch.Subvenance {
			if _, ok := seen[pfsdb.BranchKey(subvenantBranchHandle)]; !ok {
				subvenantBranch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, subvenantBranchHandle)
				if err != nil {
					return errors.Wrapf(err, "propagate branches: get subvenant branchHandle %q", pfsdb.BranchKey(subvenantBranchHandle))
				}
				branchInfo := subvenantBranch.BranchInfo
				seen[pfsdb.BranchKey(subvenantBranchHandle)] = branchInfo
				propagatedBranches = append(propagatedBranches, branchInfo)
			}
		}
	}
	sort.Slice(propagatedBranches, func(i, j int) bool {
		return len(propagatedBranches[i].Provenance) < len(propagatedBranches[j].Provenance)
	})
	// add new commits, set their ancestry + provenance pointers, and advance branch heads
	for _, bi := range propagatedBranches {
		// TODO(acohen4): can we just make calls to addCommitInfoToDB() here?
		// Do not propagate an open commit onto spout output branches (which should
		// only have a single provenance on a spec commit)
		if len(bi.Provenance) == 1 && bi.Provenance[0].Repo.Type == pfs.SpecRepoType {
			continue
		}
		// necessary when upstream and downstream branches are created in the same pachyderm transaction.
		// ex. when the spec and meta repos are created, the meta commit is already created at branch create time.
		if bi.GetHead().GetId() == txnCtx.CommitSetID {
			continue
		}
		newCommit := &pfs.Commit{
			Repo:   bi.Branch.Repo,
			Branch: bi.Branch,
			Id:     txnCtx.CommitSetID,
		}
		newCommitInfo := &pfs.CommitInfo{
			Commit:    newCommit,
			Origin:    &pfs.CommitOrigin{Kind: pfs.OriginKind_AUTO},
			Started:   txnCtx.Timestamp,
			CreatedBy: txnCtx.Username(),
		}
		// enumerate the new commit's provenance
		for _, b := range bi.DirectProvenance {
			var provCommit *pfs.Commit
			if pbi, ok := seen[pfsdb.BranchKey(b)]; ok {
				provCommit = client.NewCommit(pbi.Branch.Repo.Project.Name, pbi.Branch.Repo.Name, pbi.Branch.Name, txnCtx.CommitSetID)
			} else {
				provBranchInfo, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, b)
				if err != nil {
					return errors.Wrapf(err, "get provenant branchHandle %q", pfsdb.BranchKey(b))
				}
				provCommit = provBranchInfo.Head
			}
			newCommitInfo.DirectProvenance = append(newCommitInfo.DirectProvenance, provCommit)
		}
		// Set 'newCommit's ParentCommit, 'branch.Head's ChildCommits and 'branch.Head'
		newCommitInfo.ParentCommit = proto.Clone(bi.Head).(*pfs.Commit)
		bi.Head = newCommit
		// create open 'commit'.
		commitID, err := pfsdb.CreateCommit(ctx, txnCtx.SqlTx, newCommitInfo)
		if err != nil {
			return errors.Wrapf(err, "create new commit %q", pfsdb.CommitKey(newCommit))
		}
		if newCommitInfo.ParentCommit != nil {
			if err := pfsdb.CreateCommitParent(ctx, txnCtx.SqlTx, newCommitInfo.ParentCommit, commitID); err != nil {
				return errors.Wrapf(err, "update parent commit %q with child %q", pfsdb.CommitKey(newCommitInfo.ParentCommit), pfsdb.CommitKey(newCommit))
			}
		}
		// add commit provenance
		for _, c := range newCommitInfo.DirectProvenance {
			if err := pfsdb.AddCommitProvenance(txnCtx.SqlTx, newCommit, c); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q", pfsdb.CommitKey(newCommit), pfsdb.CommitKey(c))
			}
		}
		// update branch head.
		if _, err := pfsdb.UpsertBranch(ctx, txnCtx.SqlTx, bi); err != nil {
			return errors.Wrapf(err, "put branchHandle %q with head %q", pfsdb.BranchKey(bi.Branch), pfsdb.CommitKey(bi.Head))
		}
	}
	txnCtx.PropagateJobs()
	return nil
}

// fillNewBranches helps create the upstream branches on which a branch is provenant, if they don't exist.
// TODO(provenance): consider removing this functionality
func (a *apiServer) fillNewBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branchHandle *pfs.Branch, provenance []*pfs.Branch) error {
	branchHandle.Repo.EnsureProject()
	newRepoCommits := make(map[string]*pfsdb.Commit)
	for _, p := range provenance {
		branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, p)
		if err != nil && !pfsdb.IsNotFoundError(err) {
			return errors.Wrap(err, "get branch info")
		}
		if branch != nil {
			// Skip because the branch already exists.
			continue
		}
		var newHead *pfsdb.Commit
		head, ok := newRepoCommits[p.Repo.Key()]
		if !ok {
			head, err = a.makeEmptyCommit(ctx, txnCtx, p, nil, nil)
			if err != nil {
				return errors.Wrap(err, "create empty commit")
			}
			newHead = head
			newRepoCommits[p.Repo.Key()] = head
		}
		branchInfo := &pfs.BranchInfo{Branch: p, Head: head.CommitInfo.Commit, CreatedBy: txnCtx.Username()}
		if err := upsertBranchAndSyncCommit(ctx, txnCtx.SqlTx, branchInfo, newHead.ID); err != nil {
			return errors.Wrap(err, "fill new branches: upsert branch")
		}
	}
	return nil
}

// upsertBranchAndSyncCommit upserts a branch but also updates the head commit's 'Branch' field if needed.
// the reason why this exists is because, in our data model, a branch must have a head commit, but a commit is also dependent
// on the branch name.
// the 'Branch' field is targeted for deprecation, so this function should make it easy to find all the callers.
func upsertBranchAndSyncCommit(ctx context.Context, tx *pachsql.Tx, branchInfo *pfs.BranchInfo, headID pfsdb.CommitID) error {
	branchID, err := pfsdb.UpsertBranch(ctx, tx, branchInfo)
	if err != nil {
		return errors.Wrap(err, "update head")
	}
	if headID != 0 {
		return errors.Wrap(pfsdb.UpdateCommitBranch(ctx, tx, headID, branchID), "upsert branch and sync commit")
	}
	return nil
}

// createBranch creates a new branch or updates an existing branch (must be one
// or the other). Most importantly, it sets 'branch.DirectProvenance' to
// 'provenance' and then for all (downstream) branches, restores the invariant:
//
//	∀ b . b.Provenance = ∪ b'.Provenance (where b' ∈ b.DirectProvenance)
//
// This invariant is assumed to hold for all branches upstream of 'branch', but not
// for 'branch' itself once 'b.Provenance' has been set.
//
// i.e. up to one branch in a repo can be present within a DAG
func (a *apiServer) createBranch(ctx context.Context, txnCtx *txncontext.TransactionContext, branchHandle *pfs.Branch, commitHandle *pfs.Commit, provenance []*pfs.Branch, trigger *pfs.Trigger) error {
	// Validate arguments
	if branchHandle == nil {
		return errors.New("branch cannot be nil")
	}
	if branchHandle.Repo == nil {
		return errors.New("branch repo cannot be nil")
	}
	if err := a.validateTrigger(ctx, txnCtx, branchHandle, trigger); err != nil {
		return err
	}
	if len(provenance) > 0 && trigger != nil {
		return errors.New("a branch cannot have both provenance and a trigger")
	}
	var err error
	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, branchHandle.Repo, auth.Permission_REPO_CREATE_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	// Validate request
	var allBranches []*pfs.Branch
	allBranches = append(allBranches, provenance...)
	allBranches = append(allBranches, branchHandle)
	for _, b := range allBranches {
		if err := ancestry.ValidateName(b.Name); err != nil {
			return err
		}
	}
	// Create any of the provenance branches that don't exist yet, and give them an empty commit
	if err := a.fillNewBranches(ctx, txnCtx, branchHandle, provenance); err != nil {
		return err
	}
	// if the user passed a commit to point this branch at, resolve it
	if commitHandle != nil {
		ci, err := a.resolveCommitInfo(ctx, txnCtx.SqlTx, commitHandle)
		if err != nil {
			return errors.Wrapf(err, "unable to inspect %s", commitHandle)
		}
		commitHandle = ci.Commit
	}
	// retrieve the current version of this branch and set its head if specified
	var oldProvenance []*pfs.Branch
	propagate := false
	branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branchHandle)
	if err != nil && !pfsdb.IsNotFoundError(err) {
		return errors.Wrap(err, "create branch")
	}
	branchInfo := &pfs.BranchInfo{
		CreatedBy: txnCtx.Username(),
	}
	if branch != nil {
		branchInfo = branch.BranchInfo
	}
	branchInfo.Branch = branchHandle
	oldProvenance = branchInfo.DirectProvenance
	branchInfo.DirectProvenance = nil
	for _, provBranch := range provenance {
		if proto.Equal(provBranch.Repo, branchHandle.Repo) {
			return errors.Errorf("repo %s cannot be in the provenance of its own branch", branchHandle.Repo)
		}
		add(&branchInfo.DirectProvenance, provBranch)
	}
	if commitHandle != nil {
		branchInfo.Head = commitHandle
		propagate = true
	}
	newHead := pfsdb.CommitID(0)
	// if we don't have a branch head, or the provenance has changed, add a new commit to the branch to capture the changed structure
	// the one edge case here, is that it's undesirable to add a commit in the case where provenance is completely removed...
	//
	// TODO(provenance): This sort of hurts Branch Provenance invariant. See if we can re-assess....
	if branchInfo.Head == nil || (!same(oldProvenance, provenance) && len(provenance) != 0) {
		c, err := a.makeEmptyCommit(ctx, txnCtx, branchHandle, provenance, branchInfo.Head)
		if err != nil {
			return err
		}
		branchInfo.Head = c.CommitInfo.Commit
		newHead = c.ID
		propagate = true
	}
	if trigger != nil && trigger.Branch != "" {
		branchInfo.Trigger = trigger
	}
	if err := upsertBranchAndSyncCommit(ctx, txnCtx.SqlTx, branchInfo, newHead); err != nil {
		return errors.Wrap(err, "create branch")
	}
	// propagate the head commit to 'branch'. This may also modify 'branch', by
	// creating a new HEAD commit if 'branch's provenance was changed and its
	// current HEAD commit has old provenance
	if propagate {
		if err := txnCtx.PropagateBranch(branchHandle); err != nil {
			return errors.Wrap(err, "propagate branch")
		}
	}
	if err := txnCtx.ValidateRepo(branchHandle.Repo); err != nil {
		return errors.Wrap(err, "check branches")
	}
	return nil
}

// validateTrigger returns an error if a trigger is invalid
func (a *apiServer) validateTrigger(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch, trigger *pfs.Trigger) error {
	if trigger == nil {
		return nil
	}
	if trigger.Branch == "" {
		return errors.Errorf("triggers must specify a branch to trigger on")
	}
	if err := ancestry.ValidateName(trigger.Branch); err != nil {
		return err
	}
	if _, err := cronutil.ParseCronExpression(trigger.RateLimitSpec); trigger.RateLimitSpec != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger rate limit spec")
	}
	if _, err := units.FromHumanSize(trigger.Size); trigger.Size != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger size")
	}
	if trigger.Commits < 0 {
		return errors.Errorf("can't trigger on a negative number of commits")
	}
	if _, err := cronutil.ParseCronExpression(trigger.CronSpec); trigger.CronSpec != "" && err != nil {
		return errors.Wrapf(err, "invalid trigger cron spec")
	}

	biMaps := make(map[string]*pfs.BranchInfo)
	if err := a.listBranchInTransaction(ctx, txnCtx, branch.Repo, false, func(bi *pfs.BranchInfo) error {
		biMaps[bi.Branch.Name] = proto.Clone(bi).(*pfs.BranchInfo)
		return nil
	}); err != nil {
		return err
	}
	b := trigger.Branch
	for {
		if b == branch.Name {
			return errors.Errorf("triggers cannot form a loop")
		}
		bi, ok := biMaps[b]
		if ok && bi.Trigger != nil {
			b = bi.Trigger.Branch
		} else {
			break
		}
	}
	return nil
}

// delete branches from most provenance to least, that way if one
// branch is provenant on another (which is likely the case when
// multiple repos are provided) we delete them in the right order.
func (a *apiServer) deleteBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, branchInfos []*pfs.BranchInfo, force bool) error {
	sort.Slice(branchInfos, func(i, j int) bool { return len(branchInfos[i].Provenance) > len(branchInfos[j].Provenance) })
	for _, bi := range branchInfos {
		if err := a.deleteBranch(ctx, txnCtx, bi.Branch, force); err != nil {
			return errors.Wrapf(err, "delete branch %s", bi.Branch)
		}
	}
	return nil
}

func (a *apiServer) inspectBranch(ctx context.Context, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	branchInfo := &pfs.BranchInfo{}
	if err := a.txnEnv.WithReadContext(ctx, func(ctx context.Context, txnCtx *txncontext.TransactionContext) error {
		var err error
		branchInfo, err = a.inspectBranchInTransaction(ctx, txnCtx, branch)
		return err
	}); err != nil {
		if branch != nil && branch.Repo != nil && branch.Repo.Project != nil && pfsdb.IsNotFoundError(err) {
			return nil, errors.Join(err, pfsserver.ErrBranchNotFound{Branch: branch})
		}
		return nil, err
	}
	return branchInfo, nil
}

func (a *apiServer) inspectBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, branch *pfs.Branch) (*pfs.BranchInfo, error) {
	// Validate arguments
	if branch == nil {
		return nil, errors.New("branch cannot be nil")
	}
	if branch.Repo == nil {
		return nil, errors.New("branch repo cannot be nil")
	}
	if branch.Repo.Project == nil {
		branch.Repo.Project = &pfs.Project{Name: "default"}
	}

	// Check that the user is logged in, but don't require any access level
	if _, err := txnCtx.WhoAmI(); err != nil {
		if !auth.IsErrNotActivated(err) {
			return nil, errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to inspect a branch)")
		}
	}

	result, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branch)
	if err != nil {
		return nil, errors.Wrap(err, "inspect branch in transaction")
	}
	return result.BranchInfo, nil
}

func (a *apiServer) listBranch(ctx context.Context, reverse bool, cb func(*pfs.BranchInfo) error) error {
	if _, err := a.env.Auth.WhoAmI(ctx, &auth.WhoAmIRequest{}); err != nil && !auth.IsErrNotActivated(err) {
		return errors.EnsureStack(err)
	} else if err == nil {
		return errors.New("Cannot list branches from all repos with auth activated")
	}

	var bis []*pfs.BranchInfo
	sendBis := func() error {
		if !reverse {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) < len(bis[j].Provenance) })
		} else {
			sort.Slice(bis, func(i, j int) bool { return len(bis[i].Provenance) > len(bis[j].Provenance) })
		}
		for i := range bis {
			if err := cb(bis[i]); err != nil {
				return err
			}
		}
		bis = nil
		return nil
	}

	lastRev := int64(-1)
	listCallback := func(branch pfsdb.Branch) error {
		if branch.Revision != lastRev {
			if err := sendBis(); err != nil {
				return errors.EnsureStack(err)
			}
			lastRev = branch.Revision
		}
		bis = append(bis, proto.Clone(branch.BranchInfo).(*pfs.BranchInfo))
		return nil
	}

	order := pfsdb.SortOrderDesc
	if reverse {
		order = pfsdb.SortOrderAsc
	}
	orderBys := []pfsdb.OrderByBranchColumn{
		{Column: pfsdb.BranchColumnCreatedAt, Order: order},
		{Column: pfsdb.BranchColumnID, Order: order},
	}
	if err := dbutil.WithTx(ctx, a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		return pfsdb.ForEachBranch(ctx, tx, nil, listCallback, orderBys...)
	}); err != nil {
		return errors.Wrap(err, "list branches")
	}
	return sendBis()
}

func (a *apiServer) listBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.Repo, reverse bool, cb func(*pfs.BranchInfo) error) error {
	// Validate arguments
	if repo == nil {
		return errors.New("repo cannot be nil")
	}

	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, repo, auth.Permission_REPO_LIST_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	// Make sure that the repo exists
	if repo.Name != "" {
		if _, err := pfsdb.GetRepoByName(ctx, txnCtx.SqlTx, repo.Project.Name, repo.Name, repo.Type); err != nil {
			if pfsdb.IsErrRepoNotFound(err) {
				return pfsserver.ErrRepoNotFound{Repo: repo}
			}
			return errors.EnsureStack(err)
		}
	}
	order := pfsdb.SortOrderDesc
	if reverse {
		order = pfsdb.SortOrderAsc
	}
	orderBys := []pfsdb.OrderByBranchColumn{
		{Column: pfsdb.BranchColumnCreatedAt, Order: order},
		{Column: pfsdb.BranchColumnID, Order: order},
	}
	return errors.Wrap(pfsdb.ForEachBranch(ctx, txnCtx.SqlTx, &pfs.Branch{Repo: repo}, func(branch pfsdb.Branch) error {
		return cb(branch.BranchInfo)
	}, orderBys...), "list branch in transaction")
}

func (a *apiServer) listRepoBranches(ctx context.Context, txnCtx *txncontext.TransactionContext, repo *pfs.RepoInfo) ([]*pfs.BranchInfo, error) {
	var bis []*pfs.BranchInfo
	for _, branch := range repo.Branches {
		bi, err := a.inspectBranchInTransaction(ctx, txnCtx, branch)
		if err != nil {
			return nil, errors.Wrapf(err, "error inspecting branch %s", branch)
		}
		bis = append(bis, bi)
	}
	return bis, nil
}

func (a *apiServer) deleteBranch(ctx context.Context, txnCtx *txncontext.TransactionContext, branchHandle *pfs.Branch, force bool) error {
	if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, branchHandle.Repo, auth.Permission_REPO_DELETE_BRANCH); err != nil {
		return errors.EnsureStack(err)
	}
	branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, branchHandle)
	if err != nil {
		return errors.Wrapf(err, "get branch %q", branchHandle.Key())
	}
	return pfsdb.DeleteBranch(ctx, txnCtx.SqlTx, branch, force)
}

func (a *apiServer) walkBranchProvenanceTx(ctx context.Context, txnCtx *txncontext.TransactionContext, request *WalkBranchProvenanceRequest,
	startId pfsdb.BranchID, cb func(branchInfo *pfs.BranchInfo) error) error {
	branches, err := pfsdb.GetProvenantBranches(ctx, txnCtx.SqlTx, startId,
		pfsdb.WithMaxDepth(request.MaxDepth), pfsdb.WithLimit(request.MaxBranches))
	if err != nil {
		return errors.Wrap(err, "walk branch provenance in transaction")
	}
	for _, branch := range branches {
		if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, branch.Branch.Repo, auth.Permission_REPO_LIST_BRANCH); err != nil {
			return errors.EnsureStack(err)
		}
		if err := cb(branch.BranchInfo); err != nil {
			return errors.Wrap(err, "walk branch provenance in transaction")
		}
	}
	return nil
}

func (a *apiServer) walkBranchSubvenanceTx(ctx context.Context, txnCtx *txncontext.TransactionContext, request *WalkBranchSubvenanceRequest,
	startId pfsdb.BranchID, cb func(branchInfo *pfs.BranchInfo) error) error {
	branches, err := pfsdb.GetSubvenantBranches(ctx, txnCtx.SqlTx, startId,
		pfsdb.WithMaxDepth(request.MaxDepth), pfsdb.WithLimit(request.MaxBranches))
	if err != nil {
		return errors.Wrap(err, "walk branch subvenance in transaction")
	}
	for _, branch := range branches {
		if err := a.env.Auth.CheckRepoIsAuthorizedInTransaction(ctx, txnCtx, branch.Branch.Repo, auth.Permission_REPO_LIST_BRANCH); err != nil {
			return errors.EnsureStack(err)
		}
		if err := cb(branch.BranchInfo); err != nil {
			return errors.Wrap(err, "walk branch subvenance transaction")
		}
	}
	return nil
}

// for a DAG to be valid, it may not have a multiple branches from the same repo
// reachable by traveling edges bidirectionally. The reason is that this would complicate resolving
func (a *apiServer) validateDAGStructure(ctx context.Context, txnCtx *txncontext.TransactionContext, bs []*pfs.Branch) error {
	cache := make(map[string]*pfs.BranchInfo)
	getBranchInfo := func(b *pfs.Branch) (*pfs.BranchInfo, error) {
		if bi, ok := cache[pfsdb.BranchKey(b)]; ok {
			return bi, nil
		}
		branch, err := pfsdb.GetBranch(ctx, txnCtx.SqlTx, b)
		if err != nil {
			return nil, errors.Wrapf(err, "get branch info")
		}
		cache[pfsdb.BranchKey(b)] = branch.BranchInfo
		return branch.BranchInfo, nil
	}
	for _, b := range bs {
		dagBranches := make(map[string]*pfs.BranchInfo)
		bi, err := getBranchInfo(b)
		if err != nil {
			return errors.Wrapf(err, "loading branch info for %q during validation", b.String())
		}
		dagBranches[pfsdb.BranchKey(b)] = bi
		expanded := make(map[string]struct{}) // expanded branches
		for {
			hasExpanded := false
			for _, bi := range dagBranches {
				if _, ok := expanded[pfsdb.BranchKey(bi.Branch)]; ok {
					continue
				}
				hasExpanded = true
				expanded[pfsdb.BranchKey(bi.Branch)] = struct{}{}
				for _, b := range bi.Provenance {
					bi, err := getBranchInfo(b)
					if err != nil {
						return err
					}
					dagBranches[pfsdb.BranchKey(b)] = bi
				}
				for _, b := range bi.Subvenance {
					bi, err := getBranchInfo(b)
					if err != nil {
						return err
					}
					dagBranches[pfsdb.BranchKey(b)] = bi
				}
			}
			if !hasExpanded {
				break
			}
		}
		foundRepos := make(map[string]struct{})
		for _, bi := range dagBranches {
			if _, ok := foundRepos[bi.Branch.Repo.String()]; ok {
				return &pfsserver.ErrInvalidProvenanceStructure{Branch: bi.Branch}
			}
			foundRepos[bi.Branch.Repo.String()] = struct{}{}
		}
	}
	return nil
}

// TODO: Is this really necessary?
type branchSet []*pfs.Branch

func (b *branchSet) search(branch *pfs.Branch) (int, bool) {
	key := pfsdb.BranchKey(branch)
	i := sort.Search(len(*b), func(i int) bool {
		return pfsdb.BranchKey((*b)[i]) >= key
	})
	if i == len(*b) {
		return i, false
	}
	return i, pfsdb.BranchKey((*b)[i]) == pfsdb.BranchKey(branch)
}

func (b *branchSet) add(branch *pfs.Branch) {
	i, ok := b.search(branch)
	if !ok {
		*b = append(*b, nil)
		copy((*b)[i+1:], (*b)[i:])
		(*b)[i] = branch
	}
}

func add(bs *[]*pfs.Branch, branch *pfs.Branch) {
	(*branchSet)(bs).add(branch)
}

func (b *branchSet) has(branch *pfs.Branch) bool {
	_, ok := b.search(branch)
	return ok
}

func has(bs *[]*pfs.Branch, branch *pfs.Branch) bool {
	return (*branchSet)(bs).has(branch)
}

func same(bs []*pfs.Branch, branches []*pfs.Branch) bool {
	if len(bs) != len(branches) {
		return false
	}
	for _, br := range branches {
		if !has(&bs, br) {
			return false
		}
	}
	return true
}
