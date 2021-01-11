package server

import (
	"fmt"
	"path"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

func equalBranches(a, b []*pfs.Branch) bool {
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)
	key := path.Join
	for _, branch := range a {
		aMap[key(branch.Repo.Name, branch.Name)] = true
	}
	for _, branch := range b {
		bMap[key(branch.Repo.Name, branch.Name)] = true
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

func equalCommits(a, b []*pfs.Commit) bool {
	aMap := make(map[string]bool)
	bMap := make(map[string]bool)
	key := path.Join
	for _, commit := range a {
		aMap[key(commit.Repo.Name, commit.ID)] = true
	}
	for _, commit := range b {
		bMap[key(commit.Repo.Name, commit.ID)] = true
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

// ErrBranchProvenanceTransitivity Branch provenance is not transitively closed.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrBranchProvenanceTransitivity struct {
	BranchInfo     *pfs.BranchInfo
	FullProvenance []*pfs.Branch
}

func (e ErrBranchProvenanceTransitivity) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: branch provenance was not transitive\n")
	msg.WriteString("on branch " + e.BranchInfo.Name + " in repo " + e.BranchInfo.Branch.Repo.Name + "\n")
	fullMap := make(map[string]*pfs.Branch)
	provMap := make(map[string]*pfs.Branch)
	key := path.Join
	for _, branch := range e.FullProvenance {
		fullMap[key(branch.Repo.Name, branch.Name)] = branch
	}
	provMap[key(e.BranchInfo.Branch.Repo.Name, e.BranchInfo.Name)] = e.BranchInfo.Branch
	for _, branch := range e.BranchInfo.Provenance {
		provMap[key(branch.Repo.Name, branch.Name)] = branch
	}
	msg.WriteString("the following branches are missing from the provenance:\n")
	for k, v := range fullMap {
		if _, ok := provMap[k]; !ok {
			msg.WriteString(v.Name + " in repo " + v.Repo.Name + "\n")
		}
	}
	return msg.String()
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
	return fmt.Sprintf("consistency error: the commit %v in repo %v could not be found while checking %v",
		e.Commit.ID, e.Commit.Repo.Name, e.Location)
}

// ErrInconsistentCommitProvenance Commit provenance somehow has a branch and commit from different repos.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrInconsistentCommitProvenance struct {
	CommitProvenance *pfs.CommitProvenance
}

func (e ErrInconsistentCommitProvenance) Error() string {
	return fmt.Sprintf("consistency error: the commit provenance has repo %v for the branch but repo %v for the commit",
		e.CommitProvenance.Branch.Repo.Name, e.CommitProvenance.Commit.Repo.Name)
}

// ErrHeadProvenanceInconsistentWithBranch The head provenance of a branch does not match the branch's provenance
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrHeadProvenanceInconsistentWithBranch struct {
	BranchInfo     *pfs.BranchInfo
	ProvBranchInfo *pfs.BranchInfo
	HeadCommitInfo *pfs.CommitInfo
}

func (e ErrHeadProvenanceInconsistentWithBranch) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: head provenance is not consistent with branch provenance\n")
	msg.WriteString("on branch " + e.BranchInfo.Name + " in repo " + e.BranchInfo.Branch.Repo.Name + "\n")
	msg.WriteString("which has head commit " + e.HeadCommitInfo.Commit.ID + "\n")
	msg.WriteString("this branch is provenant on the branch " +
		e.ProvBranchInfo.Name + " in repo " + e.ProvBranchInfo.Branch.Repo.Name + "\n")
	msg.WriteString("which has head commit " + e.ProvBranchInfo.Head.ID + "\n")
	msg.WriteString("but this commit is missing from the head commit provenance\n")
	return msg.String()
}

// ErrProvenanceTransitivity Commit provenance is not transitively closed.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrProvenanceTransitivity struct {
	CommitInfo     *pfs.CommitInfo
	FullProvenance []*pfs.Commit
}

func (e ErrProvenanceTransitivity) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: commit provenance was not transitive\n")
	msg.WriteString("on commit " + e.CommitInfo.Commit.ID + " in repo " + e.CommitInfo.Commit.Repo.Name + "\n")
	fullMap := make(map[string]*pfs.Commit)
	provMap := make(map[string]*pfs.Commit)
	key := path.Join
	for _, prov := range e.FullProvenance {
		fullMap[key(prov.Repo.Name, prov.ID)] = prov
	}
	for _, prov := range e.CommitInfo.Provenance {
		provMap[key(prov.Commit.Repo.Name, prov.Commit.ID)] = prov.Commit
	}
	msg.WriteString("the following commit provenances are missing from the full provenance:\n")
	for k, v := range fullMap {
		if _, ok := provMap[k]; !ok {
			msg.WriteString(v.ID + " in repo " + v.Repo.Name + "\n")
		}
	}
	return msg.String()
}

// ErrNilCommitInSubvenance Commit provenance somehow has a branch and commit from different repos.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrNilCommitInSubvenance struct {
	CommitInfo      *pfs.CommitInfo
	SubvenanceRange *pfs.CommitRange
}

func (e ErrNilCommitInSubvenance) Error() string {
	upper := "<nil>"
	if e.SubvenanceRange.Upper != nil {
		upper = e.SubvenanceRange.Upper.ID
	}
	lower := "<nil>"
	if e.SubvenanceRange.Lower != nil {
		lower = e.SubvenanceRange.Lower.ID
	}
	return fmt.Sprintf("consistency error: the commit %v has nil subvenance in the %v - %v range",
		e.CommitInfo.Commit.ID, lower, upper)
}

// ErrSubvenanceOfProvenance The commit was not found in its provenance's subvenance
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrSubvenanceOfProvenance struct {
	CommitInfo     *pfs.CommitInfo
	ProvCommitInfo *pfs.CommitInfo
}

func (e ErrSubvenanceOfProvenance) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: the commit was not in its provenance's subvenance\n")
	msg.WriteString("commit " + e.CommitInfo.Commit.ID + " in repo " + e.CommitInfo.Commit.Repo.Name + "\n")
	msg.WriteString("provenance commit " + e.ProvCommitInfo.Commit.ID + " in repo " + e.ProvCommitInfo.Commit.Repo.Name + "\n")
	return msg.String()
}

// ErrProvenanceOfSubvenance The commit was not found in its subvenance's provenance
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrProvenanceOfSubvenance struct {
	CommitInfo     *pfs.CommitInfo
	SubvCommitInfo *pfs.CommitInfo
}

func (e ErrProvenanceOfSubvenance) Error() string {
	var msg strings.Builder
	msg.WriteString("consistency error: the commit was not in its subvenance's provenance\n")
	msg.WriteString("commit " + e.CommitInfo.Commit.ID + " in repo " + e.CommitInfo.Commit.Repo.Name + "\n")
	msg.WriteString("subvenance commit " + e.SubvCommitInfo.Commit.ID + " in repo " + e.SubvCommitInfo.Commit.Repo.Name + "\n")
	return msg.String()
}

// fsck verifies that pfs satisfies the following invariants:
// 1. Branch provenance is transitive
// 2. Head commit provenance has heads of branch's branch provenance
// 3. Commit provenance is transitive
// 4. Commit provenance and commit subvenance are dual relations
// If fix is true it will attempt to fix as many of these issues as it can.
func (d *driver) fsck(pachClient *client.APIClient, fix bool, cb func(*pfs.FsckResponse) error) error {
	// Check that the user is logged in (user doesn't need any access level to
	// fsck, but they must be authenticated if auth is active)
	if _, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{}); err != nil {
		if !auth.IsErrNotActivated(err) {
			return errors.Wrapf(grpcutil.ScrubGRPC(err), "error authenticating (must log in to run fsck)")
		}
	}

	ctx := pachClient.Ctx()

	repos := d.repos.ReadOnly(ctx)
	key := path.Join

	onError := func(err error) error { return cb(&pfs.FsckResponse{Error: err.Error()}) }
	onFix := func(fix string) error { return cb(&pfs.FsckResponse{Fix: fix}) }

	// collect all the info for the branches and commits in pfs
	branchInfos := make(map[string]*pfs.BranchInfo)
	commitInfos := make(map[string]*pfs.CommitInfo)
	newCommitInfos := make(map[string]*pfs.CommitInfo)
	repoInfo := &pfs.RepoInfo{}
	if err := repos.List(repoInfo, col.DefaultOptions, func(repoName string) error {
		commits := d.commits(repoName).ReadOnly(ctx)
		commitInfo := &pfs.CommitInfo{}
		if err := commits.List(commitInfo, col.DefaultOptions, func(commitID string) error {
			commitInfos[key(repoName, commitID)] = proto.Clone(commitInfo).(*pfs.CommitInfo)
			return nil
		}); err != nil {
			return err
		}
		branches := d.branches(repoName).ReadOnly(ctx)
		branchInfo := &pfs.BranchInfo{}
		return branches.List(branchInfo, col.DefaultOptions, func(branchName string) error {
			branchInfos[key(repoName, branchName)] = proto.Clone(branchInfo).(*pfs.BranchInfo)
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
			directProvenanceInfo := branchInfos[key(directProvenance.Repo.Name, directProvenance.Name)]
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

		// 	if there is a HEAD commit
		if bi.Head != nil {
			// we expect the branch's provenance to equal the HEAD commit's provenance
			// i.e branch.Provenance contains the branch provBranch and provBranch.Head != nil implies branch.Head.Provenance contains provBranch.Head
			// =>
			for _, provBranch := range bi.Provenance {
				provBranchInfo, ok := branchInfos[key(provBranch.Repo.Name, provBranch.Name)]
				if !ok {
					if err := onError(ErrBranchInfoNotFound{Branch: provBranch}); err != nil {
						return err
					}
					continue
				}
				if provBranchInfo.Head != nil {
					// in this case, the headCommit Provenance should contain provBranch.Head
					headCommitInfo, ok := commitInfos[key(bi.Head.Repo.Name, bi.Head.ID)]
					if !ok {
						if !fix {
							if err := onError(ErrCommitInfoNotFound{
								Location: "head commit provenance (=>)",
								Commit:   bi.Head,
							}); err != nil {
								return err
							}
							continue
						}
						headCommitInfo = &pfs.CommitInfo{
							Commit: bi.Head,
							Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_FSCK},
						}
						commitInfos[key(bi.Head.Repo.Name, bi.Head.ID)] = headCommitInfo
						newCommitInfos[key(bi.Head.Repo.Name, bi.Head.ID)] = headCommitInfo
						if err := onFix(fmt.Sprintf(
							"creating commit %s@%s which was missing, but referenced by %s@%s",
							bi.Head.Repo.Name, bi.Head.ID,
							bi.Branch.Repo.Name, bi.Branch.Name),
						); err != nil {
							return err
						}
					}
					// If this commit was created on an output branch, then we don't expect it to satisfy this invariant
					// due to the nature of the RunPipeline functionality.
					if headCommitInfo.Origin != nil && headCommitInfo.Origin.Kind == pfs.OriginKind_AUTO && len(headCommitInfo.Provenance) > 0 {
						continue
					}
					contains := false
					for _, headProv := range headCommitInfo.Provenance {
						if provBranchInfo.Head.Repo.Name == headProv.Commit.Repo.Name &&
							provBranchInfo.Branch.Repo.Name == headProv.Branch.Repo.Name &&
							provBranchInfo.Name == headProv.Branch.Name &&
							provBranchInfo.Head.ID == headProv.Commit.ID {
							contains = true
						}
					}
					if !contains {
						if err := onError(ErrHeadProvenanceInconsistentWithBranch{
							BranchInfo:     bi,
							ProvBranchInfo: provBranchInfo,
							HeadCommitInfo: headCommitInfo,
						}); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	// for each commit
	for _, ci := range commitInfos {
		// ensure that the provenance is transitive
		directProvenance := make([]*pfs.Commit, 0, len(ci.Provenance))
		transitiveProvenance := make([]*pfs.Commit, 0, len(ci.Provenance))
		for _, prov := range ci.Provenance {
			// not part of the above invariant, but we want to make sure provenance is self-consistent
			if prov.Commit.Repo.Name != prov.Branch.Repo.Name {
				if err := onError(ErrInconsistentCommitProvenance{CommitProvenance: prov}); err != nil {
					return err
				}
			}
			directProvenance = append(directProvenance, prov.Commit)
			transitiveProvenance = append(transitiveProvenance, prov.Commit)
			provCommitInfo, ok := commitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)]
			if !ok {
				if !fix {
					if err := onError(ErrCommitInfoNotFound{
						Location: "provenance transitivity",
						Commit:   prov.Commit,
					}); err != nil {
						return err
					}
					continue
				}
				provCommitInfo = &pfs.CommitInfo{
					Commit: prov.Commit,
					Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_FSCK},
				}
				commitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)] = provCommitInfo
				newCommitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)] = provCommitInfo
				if err := onFix(fmt.Sprintf(
					"creating commit %s@%s which was missing, but referenced by %s@%s",
					prov.Commit.Repo.Name, prov.Commit.ID,
					ci.Commit.Repo.Name, ci.Commit.ID),
				); err != nil {
					return err
				}
			}
			for _, provProv := range provCommitInfo.Provenance {
				transitiveProvenance = append(transitiveProvenance, provProv.Commit)
			}
		}
		if !equalCommits(directProvenance, transitiveProvenance) {
			if err := onError(ErrProvenanceTransitivity{
				CommitInfo:     ci,
				FullProvenance: transitiveProvenance,
			}); err != nil {
				return err
			}
		}
	}

	// for each commit
	for _, ci := range commitInfos {
		// we expect that the commit is in the subvenance of another commit iff the other commit is in our commit's provenance
		// i.e. commit.Provenance contains commit C iff C.Subvenance contains commit or C = commit
		// =>
		for _, prov := range ci.Provenance {
			if prov.Commit.ID == ci.Commit.ID {
				continue
			}
			contains := false
			provCommitInfo, ok := commitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)]
			if !ok {
				if !fix {
					if err := onError(ErrCommitInfoNotFound{
						Location: "provenance for provenance-subvenance duality (=>)",
						Commit:   prov.Commit,
					}); err != nil {
						return err
					}
					continue
				}
				provCommitInfo = &pfs.CommitInfo{
					Commit: prov.Commit,
					Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_FSCK},
				}
				commitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)] = provCommitInfo
				newCommitInfos[key(prov.Commit.Repo.Name, prov.Commit.ID)] = provCommitInfo
				if err := onFix(fmt.Sprintf(
					"creating commit %s@%s which was missing, but referenced by %s@%s",
					prov.Commit.Repo.Name, prov.Commit.ID,
					ci.Commit.Repo.Name, ci.Commit.ID),
				); err != nil {
					return err
				}
			}
			for _, subvRange := range provCommitInfo.Subvenance {
				subvCommit := subvRange.Upper
				// loop through the subvenance range
				for {
					if subvCommit == nil {
						if err := onError(ErrNilCommitInSubvenance{
							CommitInfo:      provCommitInfo,
							SubvenanceRange: subvRange,
						}); err != nil {
							return err
						}
						break // can't continue loop now that subvCommit is nil
					}
					subvCommitInfo, ok := commitInfos[key(subvCommit.Repo.Name, subvCommit.ID)]
					if !ok {
						if !fix {
							if err := onError(ErrCommitInfoNotFound{
								Location: "subvenance for provenance-subvenance duality (=>)",
								Commit:   subvCommit,
							}); err != nil {
								return err
							}
							break // can't continue loop if we can't find this commit
						}
						subvCommitInfo = &pfs.CommitInfo{
							Commit: subvCommit,
							Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_FSCK},
						}
						commitInfos[key(subvCommit.Repo.Name, subvCommit.ID)] = subvCommitInfo
						newCommitInfos[key(subvCommit.Repo.Name, subvCommit.ID)] = subvCommitInfo
						if err := onFix(fmt.Sprintf(
							"creating commit %s@%s which was missing, but referenced by %s@%s",
							subvCommit.Repo.Name, subvCommit.ID,
							ci.Commit.Repo.Name, ci.Commit.ID),
						); err != nil {
							return err
						}
					}
					if ci.Commit.ID == subvCommit.ID {
						contains = true
					}

					if subvCommit.ID == subvRange.Lower.ID {
						break // check at the end of the loop so we fsck 'lower' too (inclusive range)
					}
					subvCommit = subvCommitInfo.ParentCommit
				}
			}
			if !contains {
				if err := onError(ErrSubvenanceOfProvenance{
					CommitInfo:     ci,
					ProvCommitInfo: provCommitInfo,
				}); err != nil {
					return err
				}
			}
		}
		// <=
		for _, subvRange := range ci.Subvenance {
			subvCommit := subvRange.Upper
			// loop through the subvenance range
			for {
				contains := false
				if subvCommit == nil {
					if err := onError(ErrNilCommitInSubvenance{
						CommitInfo:      ci,
						SubvenanceRange: subvRange,
					}); err != nil {
						return err
					}
					break // can't continue loop now that subvCommit is nil
				}
				subvCommitInfo, ok := commitInfos[key(subvCommit.Repo.Name, subvCommit.ID)]
				if !ok {
					if !fix {
						if err := onError(ErrCommitInfoNotFound{
							Location: "subvenance for provenance-subvenance duality (<=)",
							Commit:   subvCommit,
						}); err != nil {
							return err
						}
						break // can't continue loop if we can't find this commit
					}
					subvCommitInfo = &pfs.CommitInfo{
						Commit: subvCommit,
						Origin: &pfs.CommitOrigin{Kind: pfs.OriginKind_FSCK},
					}
					commitInfos[key(subvCommit.Repo.Name, subvCommit.ID)] = subvCommitInfo
					newCommitInfos[key(subvCommit.Repo.Name, subvCommit.ID)] = subvCommitInfo
					if err := onFix(fmt.Sprintf(
						"creating commit %s@%s which was missing, but referenced by %s@%s",
						subvCommit.Repo.Name, subvCommit.ID,
						ci.Commit.Repo.Name, ci.Commit.ID),
					); err != nil {
						return err
					}
				}
				if ci.Commit.ID == subvCommit.ID {
					contains = true
				}
				for _, subvProv := range subvCommitInfo.Provenance {
					if ci.Commit.Repo.Name == subvProv.Commit.Repo.Name &&
						ci.Commit.ID == subvProv.Commit.ID {
						contains = true
					}
				}

				if !contains {
					if err := onError(ErrProvenanceOfSubvenance{
						CommitInfo:     ci,
						SubvCommitInfo: subvCommitInfo,
					}); err != nil {
						return err
					}
				}

				if subvCommit.ID == subvRange.Lower.ID {
					break // check at the end of the loop so we fsck 'lower' too (inclusive range)
				}
				subvCommit = subvCommitInfo.ParentCommit
			}
		}
	}
	if fix {
		_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
			for _, ci := range newCommitInfos {
				// We've observed users getting ErrExists from this create,
				// which doesn't make a lot of sense, but we insulate against
				// it anyways so it doesn't prevent the command from working.
				if err := d.commits(ci.Commit.Repo.Name).ReadWrite(stm).Create(ci.Commit.ID, ci); err != nil && !col.IsErrExists(err) {
					return err
				}
			}
			return nil
		})
		return err
	}
	return nil
}
