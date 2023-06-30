package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
	"google.golang.org/protobuf/proto"
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
	msg.WriteString("on branch " + e.BranchInfo.Branch.String() + "\n")
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
			msg.WriteString(v.Name + " in repo " + v.Repo.String() + "\n")
		}
	}
	return msg.String()
}

type ErrBranchSubvenanceTransitivity struct {
	BranchInfo        *pfs.BranchInfo
	MissingSubvenance *pfs.Branch
}

func (e ErrBranchSubvenanceTransitivity) Error() string {
	return fmt.Sprintf("consistency error: branch %s is missing branch %s in its subvenance\n", e.BranchInfo.Branch, e.MissingSubvenance)
}

// ErrBranchInfoNotFound Branch info could not be found. Typically because of an incomplete deletion of a branch.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrBranchInfoNotFound struct {
	Branch *pfs.Branch
}

func (e ErrBranchInfoNotFound) Error() string {
	return fmt.Sprintf("consistency error: the branch %v on repo %v could not be found\n", e.Branch.Name, e.Branch.Repo)
}

// ErrCommitInfoNotFound Commit info could not be found. Typically because of an incomplete deletion of a commit.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrCommitInfoNotFound struct {
	Location string
	Commit   *pfs.Commit
}

func (e ErrCommitInfoNotFound) Error() string {
	return fmt.Sprintf("consistency error: the commit %s could not be found while checking %v",
		e.Commit, e.Location)
}

// ErrCommitAncestryBroken indicates that a parent and child commit disagree on their relationship.
// This struct contains all the information that was used to demonstrate that this invariant is not being satisfied.
type ErrCommitAncestryBroken struct {
	Parent *pfs.Commit
	Child  *pfs.Commit
}

func (e ErrCommitAncestryBroken) Error() string {
	return fmt.Sprintf("consistency error: parent commit %s and child commit %s disagree about their parent/child relationship",
		e.Parent, e.Child)
}

// ErrBranchCommitProvenanceMismatch occurs when the head commit of one of the parents of the branch is not found in the direct provenance commits of the head of the branch
type ErrBranchCommitProvenanceMismatch struct {
	Branch       *pfs.Branch
	ParentBranch *pfs.Branch
}

func (e ErrBranchCommitProvenanceMismatch) Error() string {
	return fmt.Sprintf("consistency error: parent branch %s commit is not in direct provenance of head of branch %s", e.ParentBranch, e.Branch)
}

func checkBranchProvenances(bi *pfs.BranchInfo, branchInfos map[string]*pfs.BranchInfo, onError func(error) error) error {
	// we expect the branch's provenance to be the same as the provenances of the branch's direct provenances
	// i.e. followedProvenances(branch, branch.Provenance) = followedProvenances(branch, branch.DirectProvenance, branch.DirectProvenance.Provenance)
	direct := bi.DirectProvenance
	followedProvenances := []*pfs.Branch{bi.Branch}

	for _, directProvenance := range direct {
		directProvenanceInfo := branchInfos[pfsdb.BranchKey(directProvenance)]
		if directProvenanceInfo == nil {
			if err := onError(ErrBranchInfoNotFound{Branch: directProvenance}); err != nil {
				return err
			}
			continue
		}
		followedProvenances = append(followedProvenances, directProvenance)
		followedProvenances = append(followedProvenances, directProvenanceInfo.Provenance...)
	}

	if !equalBranches(append(bi.Provenance, bi.Branch), followedProvenances) {
		if err := onError(ErrBranchProvenanceTransitivity{
			BranchInfo:     bi,
			FullProvenance: followedProvenances,
		}); err != nil {
			return err
		}
	}

	return nil
}

func checkBranchSubvenances(bi *pfs.BranchInfo, branchInfos map[string]*pfs.BranchInfo, onError func(error) error) error {
	for _, provBranch := range bi.Provenance {
		provBranchInfo := branchInfos[pfsdb.BranchKey(provBranch)]
		if !branchInSet(bi.Branch, provBranchInfo.Subvenance) {
			if err := onError(ErrBranchSubvenanceTransitivity{
				BranchInfo:        provBranchInfo,
				MissingSubvenance: bi.Branch,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkBranchCommitProvenance(bi *pfs.BranchInfo, branchInfos map[string]*pfs.BranchInfo, commitInfos map[string]*pfs.CommitInfo, onError func(error) error) error {
	// build set of all commits in branch head direct provenance
	headProvenance := make(map[string]bool)
	headCI, ok := commitInfos[pfsdb.CommitKey(bi.Head)]
	if !ok {
		return onError(ErrCommitInfoNotFound{
			Location: fmt.Sprintf("head commit of %s", bi.Branch),
			Commit:   bi.Head,
		})
	}
	for _, provCommit := range headCI.DirectProvenance {
		headProvenance[pfsdb.CommitKey(provCommit)] = true
	}

	// confirm head of each direct provenant branch is included in the commits of the branch head direct provenance
	for _, provBranch := range bi.DirectProvenance {
		provBranchInfo, ok := branchInfos[pfsdb.BranchKey(provBranch)]
		if !ok {
			if err := onError(ErrBranchInfoNotFound{Branch: provBranch}); err != nil {
				return err
			}
			continue
		}
		if _, ok := headProvenance[pfsdb.CommitKey(provBranchInfo.Head)]; !ok {
			if err := onError(ErrBranchCommitProvenanceMismatch{
				Branch:       bi.Branch,
				ParentBranch: provBranch,
			}); err != nil {
				return err
			}
			continue
		}

	}
	return nil
}

func fsckBranches(branchInfos map[string]*pfs.BranchInfo, commitInfos map[string]*pfs.CommitInfo, onError func(error) error) error {
	// for each branch
	for _, bi := range branchInfos {

		if err := checkBranchProvenances(bi, branchInfos, onError); err != nil {
			return err
		}

		// every provenant branch should have this branch in its subvenance
		if err := checkBranchSubvenances(bi, branchInfos, onError); err != nil {
			return err
		}

		if err := checkBranchCommitProvenance(bi, branchInfos, commitInfos, onError); err != nil {
			return err
		}

	}

	return nil
}

func checkParentCommit(commitInfo *pfs.CommitInfo, commitInfos map[string]*pfs.CommitInfo, onError func(error) error) error {
	if commitInfo.ParentCommit != nil {
		parentCommitInfo, ok := commitInfos[pfsdb.CommitKey(commitInfo.ParentCommit)]
		if !ok {
			if err := onError(ErrCommitInfoNotFound{
				Location: fmt.Sprintf("parent commit of %s", commitInfo.Commit),
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

	return nil
}

func checkChildCommits(commitInfo *pfs.CommitInfo, commitInfos map[string]*pfs.CommitInfo, onError func(error) error) error {
	for _, child := range commitInfo.ChildCommits {
		childCommitInfo, ok := commitInfos[pfsdb.CommitKey(child)]

		if !ok {
			if err := onError(ErrCommitInfoNotFound{
				Location: fmt.Sprintf("child commit of %s", commitInfo.Commit),
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

	return nil
}

func fsckCommits(commitInfos map[string]*pfs.CommitInfo, onError func(error) error) error {
	// For every commit
	for _, commitInfo := range commitInfos {
		// Every parent commit info should exist and point to this as a child
		if err := checkParentCommit(commitInfo, commitInfos, onError); err != nil {
			return err
		}

		// Every child commit info should exist and point to this as their parent
		if err := checkChildCommits(commitInfo, commitInfos, onError); err != nil {
			return err
		}

	}

	return nil
}

func (d *driver) listReferencedCommits(ctx context.Context) (map[string]*pfs.Commit, error) {
	cs := make(map[string]*pfs.Commit)
	if err := dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		var ids []string
		if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_totals`); err != nil {
			return errors.Wrap(err, "select commit ids from pfs.commit_totals")
		}
		for _, id := range ids {
			cs[id] = pfsdb.ParseCommit(id)
		}
		ids = make([]string, 0)
		if err := sqlTx.Select(&ids, `SELECT commit_id from  pfs.commit_diffs`); err != nil {
			return errors.Wrap(err, "select commit ids from pfs.commit_diffs")
		}
		for _, id := range ids {
			cs[id] = pfsdb.ParseCommit(id)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return cs, nil
}

func fsckDanglingCommits(ris map[string]*pfs.RepoInfo, cs map[string]*pfs.Commit) []*pfs.Commit {
	var danglingCommits []*pfs.Commit
	for _, c := range cs {
		if _, ok := ris[pfsdb.RepoKey(c.Repo)]; !ok {
			danglingCommits = append(danglingCommits, c)
		}
	}
	return danglingCommits
}

// fsck verifies that pfs satisfies the following invariants:
// 1. Branch provenance is transitive
// 2. Head commit provenance has heads of branch's branch provenance
// If fix is true it will attempt to fix as many of these issues as it can.
func (d *driver) fsck(ctx context.Context, fix bool, cb func(*pfs.FsckResponse) error) error {
	onError := func(err error) error { return cb(&pfs.FsckResponse{Error: err.Error()}) }

	// TODO(global ids): no fixable fsck issues?
	// onFix := func(fix string) error { return cb(&pfs.FsckResponse{Fix: fix}) }

	// collect all the info for the branches and commits in pfs
	branchInfos := make(map[string]*pfs.BranchInfo)
	commitInfos := make(map[string]*pfs.CommitInfo)
	repoInfos := make(map[string]*pfs.RepoInfo)
	repoInfo := &pfs.RepoInfo{}
	if err := d.repos.ReadOnly(ctx).List(repoInfo, col.DefaultOptions(), func(string) error {
		repoInfos[pfsdb.RepoKey(repoInfo.Repo)] = proto.Clone(repoInfo).(*pfs.RepoInfo)
		commitInfo := &pfs.CommitInfo{}
		if err := d.commits.ReadOnly(ctx).GetByIndex(pfsdb.CommitsRepoIndex, pfsdb.RepoKey(repoInfo.Repo), commitInfo, col.DefaultOptions(), func(string) error {
			commitInfos[pfsdb.CommitKey(commitInfo.Commit)] = proto.Clone(commitInfo).(*pfs.CommitInfo)
			return nil
		}); err != nil {
			return errors.EnsureStack(err)
		}
		branchInfo := &pfs.BranchInfo{}
		err := d.branches.ReadOnly(ctx).GetByIndex(pfsdb.BranchesRepoIndex, pfsdb.RepoKey(repoInfo.Repo), branchInfo, col.DefaultOptions(), func(string) error {
			branchInfos[pfsdb.BranchKey(branchInfo.Branch)] = proto.Clone(branchInfo).(*pfs.BranchInfo)
			return nil
		})
		return errors.EnsureStack(err)
	}); err != nil {
		return errors.EnsureStack(err)
	}
	if err := fsckBranches(branchInfos, commitInfos, onError); err != nil {
		return err
	}
	if err := fsckCommits(commitInfos, onError); err != nil {
		return err
	}
	referencedCommits, err := d.listReferencedCommits(ctx)
	if err != nil {
		return err
	}
	dangCs := fsckDanglingCommits(repoInfos, referencedCommits)
	if !fix && len(dangCs) > 0 {
		var keys []string
		for _, dc := range dangCs {
			keys = append(keys, pfsdb.CommitKey(dc))
		}
		return errors.Errorf("commits with dangling references: %v", keys)
	}
	if fix {
		return dbutil.WithTx(ctx, d.env.DB, func(ctx context.Context, sqlTx *pachsql.Tx) error {
			for _, dc := range dangCs {
				if _, err := sqlTx.ExecContext(ctx, `DELETE FROM pfs.commit_totals WHERE commit_id = $1`, pfsdb.CommitKey(dc)); err != nil {
					return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_totals", pfsdb.CommitKey(dc))
				}
				if _, err := sqlTx.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, pfsdb.CommitKey(dc)); err != nil {
					return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_diffs", pfsdb.CommitKey(dc))
				}
			}
			return nil
		})
	}
	return nil
}

type ErrZombieData struct {
	Commit *pfs.Commit
	ID     string
}

func (e ErrZombieData) Error() string {
	return fmt.Sprintf("commit %v contains output from datum %s which should have been deleted", e.Commit, e.ID)
}

type fileStream struct {
	iterator   *fileset.Iterator
	file       fileset.File
	fromOutput bool
}

func (fs *fileStream) Next() error {
	var err error
	fs.file, err = fs.iterator.Next()
	return err
}

// just match on path, we don't care about datum here
func compare(s1, s2 stream.Stream) int {
	idx1 := s1.(*fileStream).file.Index()
	idx2 := s2.(*fileStream).file.Index()
	return strings.Compare(idx1.Path, idx2.Path)
}

func (d *driver) detectZombie(ctx context.Context, outputCommit *pfs.Commit, cb func(*pfs.FsckResponse) error) error {
	log.Info(ctx, "checking for zombie data", log.Proto("outputCommit", outputCommit))
	// generate fileset that groups output files by datum
	id, err := d.createFileSet(ctx, func(w *fileset.UnorderedWriter) error {
		_, fs, err := d.openCommit(ctx, outputCommit)
		if err != nil {
			return err
		}
		return errors.EnsureStack(fs.Iterate(ctx, func(f fileset.File) error {
			id := f.Index().GetFile().GetDatum()
			// write to same path as meta commit, one line per file
			return errors.EnsureStack(w.Put(ctx, common.MetaFilePath(id), id, true,
				strings.NewReader(f.Index().Path+"\n")))
		}))
	})
	if err != nil {
		return err
	}
	// now merge with the meta commit to look for extra datums in the output commit
	return d.storage.Filesets.WithRenewer(ctx, defaultTTL, func(ctx context.Context, r *fileset.Renewer) error {
		if err := r.Add(ctx, *id); err != nil {
			return err
		}
		datumsFS, err := d.storage.Filesets.Open(ctx, []fileset.ID{*id})
		if err != nil {
			return err
		}
		_, metaFS, err := d.openCommit(ctx, ppsutil.MetaCommit(outputCommit))
		if err != nil {
			return err
		}
		var streams []stream.Stream
		streams = append(streams, &fileStream{
			iterator:   fileset.NewIterator(ctx, datumsFS.Iterate),
			fromOutput: true,
		}, &fileStream{
			iterator: fileset.NewIterator(ctx, metaFS.Iterate, index.WithPrefix("/"+common.MetaPrefix)),
		})
		pq := stream.NewPriorityQueue(streams, compare)
		return errors.EnsureStack(pq.Iterate(func(ss []stream.Stream) error {
			if len(ss) == 2 {
				return nil // datum is present both in output and meta, as expected
			}
			s := ss[0].(*fileStream)
			if !s.fromOutput {
				return nil // datum doesn't have any output, not an error
			}
			// this is zombie data: output files not associated with any current datum
			// report each file back as an error
			id := s.file.Index().File.Datum
			return miscutil.WithPipe(func(w io.Writer) error {
				return errors.EnsureStack(s.file.Content(ctx, w))
			}, func(r io.Reader) error {
				sc := bufio.NewScanner(r)
				for sc.Scan() {
					if err := cb(&pfs.FsckResponse{
						Error: fmt.Sprintf("stale datum %s had file %s", id, sc.Text()),
					}); err != nil {
						return err
					}
				}
				return nil
			})
		}))
	})
}
