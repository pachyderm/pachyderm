package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func equalBranches(a, b []*pfs.Branch) bool {
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)
	for _, branch := range a {
		aMap[pfsdb.BranchKey(branch)] = true
	}
	for _, branch := range b {
		bMap[pfsdb.BranchKey(branch)] = true
	}
	if len(aMap) != len(bMap) {
		return false
	}

	for k := range aMap {
		if !bMap[k] {
			return false
		}
	}
	return true
}

func branchInSet(branch *pfs.Branch, set []*pfs.Branch) bool {
	for _, b := range set {
		if proto.Equal(branch, b) {
			return true
		}
	}
	return false
}

// ErrBranchProvenanceTransitivity Branch provenance is not transitively closed.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrBranchProvenanceTransitivity struct {
	BranchInfo     *pfs.BranchInfo
	FullProvenance []*pfs.Branch
}

func (e ErrBranchProvenanceTransitivity) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: branch provenance was not transitive\n")
	msg.WriteString("on branch " + e.BranchInfo.Branch.Name + " in repo " + e.BranchInfo.Branch.Repo.Name + "\n")
	fullMap := make(map[string]*pfs.Branch)
	provMap := make(map[string]*pfs.Branch)
	for _, branch := range e.FullProvenance {
		fullMap[pfsdb.BranchKey(branch)] = branch
	}
	provMap[pfsdb.BranchKey(e.BranchInfo.Branch)] = e.BranchInfo.Branch
	for _, branch := range e.BranchInfo.Provenance {
		provMap[pfsdb.BranchKey(branch)] = branch
	}
	msg.WriteString("the following branches are missing from the provenance:\n")
	for k, v := range fullMap {
		if _, ok := provMap[k]; !ok {
			msg.WriteString(v.Name + " in repo " + v.Repo.Name + "\n")
		}
	}
	return msg.String()
}

type ErrBranchSubvenanceTransitivity struct {
	BranchInfo        *pfs.BranchInfo
	MissingSubvenance *pfs.Branch
}

func (e ErrBranchSubvenanceTransitivity) Error() string {
	return fmt.Sprintf("consistency error: branch %s is missing branch %s in its subvenance\n", pfsdb.BranchKey(e.BranchInfo.Branch), pfsdb.BranchKey(e.MissingSubvenance))
}

// ErrBranchInfoNotFound Branch info could not be found. Typically because of an incomplete deletion of a branch.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrBranchInfoNotFound struct {
	Branch *pfs.Branch
}

func (e ErrBranchInfoNotFound) Error() string {
	return fmt.Sprintf("consistency error: the branch %v on repo %v could not be found\n", e.Branch.Name, e.Branch.Repo.Name)
}

// ErrCommitInfoNotFound Commit info could not be found. Typically because of an incomplete deletion of a commit.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrCommitInfoNotFound struct {
	Location string
	Commit   *pfs.Commit
}

func (e ErrCommitInfoNotFound) Error() string {
	return fmt.Sprintf("consistency error: the commit %s could not be found while checking %v",
		pfsdb.CommitKey(e.Commit), e.Location)
}

// ErrCommitAncestryBroken Commit info could not be found. Typically because of an incomplete deletion of a commit.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrCommitAncestryBroken struct {
	Parent *pfs.Commit
	Child  *pfs.Commit
}

func (e ErrCommitAncestryBroken) Error() string {
	return fmt.Sprintf("consistency error: parent commit %s and child commit %s disagree about their parent/child relationship",
		pfsdb.CommitKey(e.Parent), pfsdb.CommitKey(e.Child))
}

// fsck verifies that pfs satisfies the following invariants:
// 1. Branch provenance is transitive
// 2. Head commit provenance has heads of branch's branch provenance
// If fix is true it will attempt to fix as many of these issues as it can.
func (d *driver) fsck(ctx context.Context, fix bool, cb func(*pfs.FsckResponse) error) error {
	// Check that the user is logged in (user doesn't need any access level to
	// fsck, but they must be authenticated if auth is active)
	pachClient := d.env.GetPachClient(ctx)
	if _, err := pachClient.WhoAmI(ctx, &auth.WhoAmIRequest{}); err != nil {
		if !auth.IsErrNotActivated(err) {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to run fsck)")
		}
	}

	onError := func(err error) error { return cb(&pfs.FsckResponse{Error: err.Error()}) }

	// TODO(global ids): no fixable fsck issues?
	// onFix := func(fix string) error { return cb(&pfs.FsckResponse{Fix: fix}) }

	// collect all the info for the branches and commits in pfs
	branchInfos := make(map[string]*pfs.BranchInfo)
	commitInfos := make(map[string]*pfs.CommitInfo)
	newCommitInfos := make(map[string]*pfs.CommitInfo)
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadOnly(ctx).List(repoInfo, col.DefaultOptions(), func(string) error {
		commitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadOnly(ctx).GetByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repoInfo.Repo), commitInfo, col.DefaultOptions(), func(string) error {
			commitInfos[pfsdb.CommitKey(commitInfo.Commit)] = proto.Clone(commitInfo).(*pfs.CommitInfo)
			return nil
		}); err != nil {
			return err
		}
		branchInfo := &pfs.BranchInfo{}
		return d.branches.ReadOnly(ctx).GetByIndex(pfsdb.BranchesRepoIndex, pfsdb.RepoKey(repoInfo.Repo), branchInfo, col.DefaultOptions(), func(string) error {
			branchInfos[pfsdb.BranchKey(branchInfo.Branch)] = proto.Clone(branchInfo).(*pfs.BranchInfo)
			return nil
		})
	}); err != nil {
		return err
	}

	// for each branch
	for _, bi := range branchInfos {
		// we expect the branch's provenance to equal the union of the provenances of the branch's direct provenances
		// i.e. union(branch, branch.Provenance) = union(branch, branch.DirectProvenance, branch.DirectProvenance.Provenance)
		direct := bi.DirectProvenance
		union := []*pfs.Branch{bi.Branch}
		for _, directProvenance := range direct {
			directProvenanceInfo := branchInfos[pfsdb.BranchKey(directProvenance)]
			union = append(union, directProvenance)
			if directProvenanceInfo != nil {
				union = append(union, directProvenanceInfo.Provenance...)
			}
		}

		if !equalBranches(append(bi.Provenance, bi.Branch), union) {
			if err := onError(ErrBranchProvenanceTransitivity{
				BranchInfo:     bi,
				FullProvenance: union,
			}); err != nil {
				return err
			}
		}

		// every provenant branch should have this branch in its subvenance
		for _, provBranch := range bi.Provenance {
			provBranchInfo := branchInfos[pfsdb.BranchKey(provBranch)]
			if !branchInSet(bi.Branch, provBranchInfo.Subvenance) {
				if !fix {
					if err := onError(ErrBranchSubvenanceTransitivity{
						BranchInfo:        provBranchInfo,
						MissingSubvenance: bi.Branch,
					}); err != nil {
						return err
					}
				} else {
					// TODO(global ids): fix branch subvenance
				}
			}
		}

		// 	if there is a HEAD commit
		if bi.Head != nil {
			// we expect the branch's provenance to equal the HEAD commit's provenance
			// i.e branch.Provenance contains the branch provBranch and provBranch.Head != nil implies branch.Head.Provenance contains provBranch.Head
			// =>
			for _, provBranch := range bi.Provenance {
				provBranchInfo, ok := branchInfos[pfsdb.BranchKey(provBranch)]
				if !ok {
					if err := onError(ErrBranchInfoNotFound{Branch: provBranch}); err != nil {
						return err
					}
					continue
				}
				if provBranchInfo.Head != nil {
					// in this case, the headCommit Provenance should contain provBranch.Head
					if _, ok := commitInfos[pfsdb.CommitKey(bi.Head)]; !ok {
						if err := onError(ErrCommitInfoNotFound{
							Location: "head commit provenance (=>)",
							Commit:   bi.Head,
						}); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// For every commit
	for _, commitInfo := range commitInfos {
		// Every parent commit info should exist and point to this as a child
		if commitInfo.ParentCommit != nil {
			parentCommitInfo, ok := commitInfos[pfsdb.CommitKey(commitInfo.ParentCommit)]
			if !ok {
				if err := onError(ErrCommitInfoNotFound{
					Location: fmt.Sprintf("parent commit of %s", pfsdb.CommitKey(commitInfo.Commit)),
					Commit:   commitInfo.ParentCommit,
				}); err != nil {
					return err
				}
			} else {
				found := false
				for _, child := range parentCommitInfo.ChildCommits {
					if proto.Equal(child, commitInfo.Commit) {
						found = true
						break
					}
				}

				if !found {
					if err := onError(ErrCommitAncestryBroken{
						Parent: parentCommitInfo.Commit,
						Child:  commitInfo.Commit,
					}); err != nil {
						return err
					}
				}
			}
		}

		// Every child commit info should exist and point to this as their parent
		for _, child := range commitInfo.ChildCommits {
			childCommitInfo, ok := commitInfos[pfsdb.CommitKey(child)]

			if !ok {
				if err := onError(ErrCommitInfoNotFound{
					Location: fmt.Sprintf("child commit of %s", pfsdb.CommitKey(commitInfo.Commit)),
					Commit:   child,
				}); err != nil {
					return err
				}
			} else {
				if childCommitInfo.ParentCommit == nil || !proto.Equal(childCommitInfo.ParentCommit, commitInfo.Commit) {
					if err := onError(ErrCommitAncestryBroken{
						Parent: commitInfo.Commit,
						Child:  childCommitInfo.Commit,
					}); err != nil {
						return err
					}
				}
			}
		}
	}

	// TODO(global ids): is there any verification we can do for commitsets?

	if fix {
		return col.NewSQLTx(ctx, d.env.GetDBClient(), func(sqlTx *sqlx.Tx) error {
			for _, ci := range newCommitInfos {
				// We've observed users getting ErrExists from this create,
				// which doesn't make a lot of sense, but we insulate against
				// it anyways so it doesn't prevent the command from working.
				if err := d.commits.ReadWrite(sqlTx).Create(pfsdb.CommitKey(ci.Commit), ci); err != nil && !col.IsErrExists(err) {
					return err
				}
			}
			return nil
		})
	}
	return nil
}
