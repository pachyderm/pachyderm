package server

import (
	"fmt"
	"path"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
)

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
				if err := onFix(fmt.Sprintf("adding %s to subvenance of %s", ci.Commit.ID, provCommitInfo.Commit.ID)); err != nil {
					return err
				}
				newProvCommitInfo := proto.Clone(provCommitInfo).(*pfs.CommitInfo)
				newProvCommitInfo.Subvenance = append(newProvCommitInfo.Subvenance, &pfs.CommitRange{
					Lower: ci.Commit,
					Upper: ci.Commit,
				})
				newCommitInfos[newProvCommitInfo.Commit.ID] = newProvCommitInfo
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
					if err := onFix(fmt.Sprintf("adding %s to provenence of %s", ci.Commit.ID, subvCommitInfo.Commit.ID)); err != nil {
						return err
					}
					newSubvCommitInfo := proto.Clone(subvCommitInfo).(*pfs.CommitInfo)
					newSubvCommitInfo.Provenance = append(subvCommitInfo.Provenance, &pfs.CommitProvenance{
						Branch: ci.Branch,
						Commit: ci.Commit,
					})
					newCommitInfos[subvCommit.ID] = newSubvCommitInfo
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
