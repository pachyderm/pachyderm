package v2_6_0

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	idxProto "github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

func updateOldCommit(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit, f func(ci *pfs.CommitInfo)) error {
	k := oldCommitKey(c)
	ci := &pfs.CommitInfo{}
	if err := getCollectionProto(ctx, tx, "commits", k, ci); err != nil {
		return errors.Wrapf(err, "retrieve old commit %q", k)
	}
	f(ci)
	return updateCollectionProto(ctx, tx, "commits", k, k, ci)
}

func updateOldBranch(ctx context.Context, tx *pachsql.Tx, b *pfs.Branch, f func(bi *pfs.BranchInfo)) error {
	k := branchKey(b)
	bi := &pfs.BranchInfo{}
	if err := getCollectionProto(ctx, tx, "branches", k, bi); err != nil {
		return errors.Wrapf(err, "get branch info for branch: %q", b)
	}
	f(bi)
	return updateCollectionProto(ctx, tx, "branches", k, k, bi)
}

func resetOldCommitInfo(ctx context.Context, tx *pachsql.Tx, ci *pfs.CommitInfo) error {
	data, err := proto.Marshal(ci)
	if err != nil {
		return errors.Wrapf(err, "unmarshal old commit %q", oldCommitKey(ci.Commit))
	}
	_, err = tx.ExecContext(ctx, `UPDATE collections.commits SET proto=$1  WHERE key=$2;`, data, oldCommitKey(ci.Commit))
	return errors.Wrapf(err, "reset old commit info %q", oldCommitKey(ci.Commit))
}

func getCommitIntId(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (int, error) {
	r := tx.QueryRowContext(ctx, "SELECT int_id FROM pfs.commits WHERE commit_id=$1", oldCommitKey(commit))
	var intId int
	if err := r.Scan(&intId); err != nil {
		return 0, errors.Wrapf(err, "scan int_id for commit_id %q", oldCommitKey(commit))
	}
	return intId, nil
}

func deleteCommit(ctx context.Context, tx *pachsql.Tx, ci *pfs.CommitInfo, ancestor *pfs.Commit, restoreLinks bool) error {
	key := oldCommitKey(ci.Commit)
	cId, err := getCommitIntId(ctx, tx, ci.Commit)
	if err != nil {
		return err
	}
	ancestorId, err := getCommitIntId(ctx, tx, ancestor)
	if err != nil {
		return err
	}
	// update provenance pointers
	if _, err := tx.ExecContext(ctx, `INSERT INTO pfs.commit_provenance (from_id, to_id) 
                                          SELECT $1, to_id FROM pfs.commit_provenance WHERE from_id=$2 ON CONFLICT DO NOTHING;`,
		ancestorId, cId); err != nil {
		return errors.Wrapf(err, "update commit 'from' provenance for commit %q", oldCommitKey(ci.Commit))
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO pfs.commit_provenance (from_id, to_id)                                       
                                          SELECT from_id, $1 FROM pfs.commit_provenance WHERE to_id=$2 ON CONFLICT DO NOTHING;`,
		ancestorId, cId); err != nil {
		return errors.Wrapf(err, "update commit 'to' provenance for commit %q", oldCommitKey(ci.Commit))
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commit_provenance WHERE from_id=$1 OR to_id=$1;`, cId); err != nil {
		return errors.Wrapf(err, "update commit 'to' provenance for commit %q", oldCommitKey(ci.Commit))
	}
	// delete commit
	if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commits WHERE int_id=$1;`, cId); err != nil {
		return errors.Wrapf(err, "delete from pfs.commits for commit %q", oldCommitKey(ci.Commit))
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM collections.commits WHERE key=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from collections.commits for commit %q", oldCommitKey(ci.Commit))
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from pfs.commit_diffs for commit_id %q", key)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM pfs.commit_totals WHERE commit_id=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from pfs.commit_totals for commit_id %q", key)
	}
	if restoreLinks {
		// update parent and child commit links
		if err := updateOldCommit(ctx, tx, ci.ParentCommit, func(parent *pfs.CommitInfo) {
			parent.ChildCommits = append(parent.ChildCommits, ci.ChildCommits...)
		}); err != nil {
			return errors.Wrapf(err, "update parent commit %q", oldCommitKey(ci.ParentCommit))
		}
		for _, c := range ci.ChildCommits {
			if err := updateOldCommit(ctx, tx, c, func(child *pfs.CommitInfo) {
				child.ParentCommit = ci.ParentCommit
			}); err != nil {
				return errors.Wrapf(err, "update child commit %q", oldCommitKey(c))
			}
		}
	}
	return nil
}

func addCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, oldCommitKey(commit), commit.ID)
	if err != nil {
		return errors.Wrapf(err, "insert commit %q", oldCommitKey(commit))
	}
	var commitIntId int
	query := `SELECT int_id FROM pfs.commits WHERE commit_id = $1;`
	err = tx.GetContext(ctx, &commitIntId, query, oldCommitKey(commit))
	return errors.Wrapf(err, "get int_id of commit %q", oldCommitKey(commit))
}

func getFileSetMd(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) (*fileset.Metadata, error) {
	var mdData []byte
	if err := tx.GetContext(ctx, &mdData, `SELECT metadata_pb FROM storage.filesets JOIN pfs.commit_totals ON id = fileset_id WHERE commit_id = $1`, oldCommitKey(c)); err != nil {
		return nil, errors.EnsureStack(err)
	}
	md := &fileset.Metadata{}
	if err := proto.Unmarshal(mdData, md); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return md, nil
}

func sameFileSets(ctx context.Context, tx *pachsql.Tx, c1 *pfs.Commit, c2 *pfs.Commit) (bool, error) {
	md1, err := getFileSetMd(ctx, tx, c1)
	if err != nil {
		return false, err
	}
	md2, err := getFileSetMd(ctx, tx, c1)
	if err != nil {
		return false, err
	}

	sameByteSlice := func(a []byte, b []byte) bool {
		return true
	}
	sameIndex := func(a *idxProto.Index, b *idxProto.Index) bool {
		if a.File != nil && b.File != nil && len(a.File.DataRefs) == len(b.File.DataRefs) {
			for i, dr := range a.File.DataRefs {
				if res := bytes.Compare(dr.Hash, b.File.DataRefs[i].Hash); res != 0 {
					return false
				}
			}
			return true
		}
		if a.Range != nil && b.Range != nil && sameByteSlice(a.Range.ChunkRef.Hash, b.Range.ChunkRef.Hash) {
			return true
		}
		return false
	}
	// TODO(acohen4): add another utility ref that builds a composite hash
	if md1.GetPrimitive() != nil && md2.GetPrimitive() != nil {
		if !sameIndex(md1.GetPrimitive().GetAdditive(), md2.GetPrimitive().GetAdditive()) || !sameIndex(md1.GetPrimitive().GetDeletive(), md2.GetPrimitive().GetDeletive()) {
			return false, nil
		}
		return true, nil
	} else if md1.GetComposite() != nil && md2.GetComposite() != nil {
		comp1Layers := md1.GetComposite().Layers
		comp2Layers := md2.GetComposite().Layers
		if len(comp1Layers) != len(comp2Layers) {
			return false, nil
		}
		for i, l := range comp1Layers {
			if l != comp2Layers[i] {
				return false, nil
			}
		}
		return true, nil
	} else {
		return false, nil
	}
}

func migrateAliasCommits(ctx context.Context, tx *pachsql.Tx) error {
	var cis []*pfs.CommitInfo
	var err error
	// first fill commit provenance to all the existing commit sets
	if cis, err = listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{}); err != nil {
		return errors.Wrap(err, "populate the pfs.commits table")
	}
	for _, ci := range cis {
		if err := addCommit(ctx, tx, ci.Commit); err != nil {
			return err
		}
	}
	deleteCommits := make(map[string]*pfs.CommitInfo)
	for _, ci := range cis {
		for _, b := range ci.OldDirectProvenance {
			if err := addCommitProvenance(ctx, tx, ci.Commit, b.NewCommit(ci.Commit.ID)); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q", oldCommitKey(ci.Commit), oldCommitKey(b.NewCommit(ci.Commit.ID)))
			}
		}
		// collect all the ALIAS commits to be deleted
		if ci.Origin.Kind == pfs.OriginKind_ALIAS && ci.ParentCommit != nil {
			deleteCommits[oldCommitKey(ci.Commit)] = proto.Clone(ci).(*pfs.CommitInfo)
		}
	}
	for _, ci := range deleteCommits {
		same, err := sameFileSets(ctx, tx, ci.Commit, ci.ParentCommit)
		if err != nil {
			return err
		}
		if !same {
			return errors.Errorf("commit %q is listed as ALIAS but has a different ID than it's first real ancestor.", oldCommitKey(ci.Commit))
		}
	}
	realAncestors := make(map[string]*pfs.CommitInfo)
	childToNewParents := make(map[*pfs.Commit]*pfs.Commit)
	for _, ci := range deleteCommits {
		anctsr := oldestAncestor(ci, deleteCommits)
		if _, ok := realAncestors[oldCommitKey(anctsr)]; !ok {
			ci := &pfs.CommitInfo{}
			if err := getCollectionProto(ctx, tx, "commits", oldCommitKey(anctsr), ci); err != nil {
				return errors.Wrapf(err, "get commit info with key %q", oldCommitKey(anctsr))
			}
			realAncestors[oldCommitKey(anctsr)] = ci
		}
		ancstrInfo := realAncestors[oldCommitKey(anctsr)]
		childSet := make(map[string]*pfs.Commit)
		for _, ch := range ancstrInfo.ChildCommits {
			childSet[oldCommitKey(ch)] = ch
		}
		delete(childSet, oldCommitKey(ci.Commit))
		newChildren := traverseToEdges(ci, deleteCommits, func(c *pfs.CommitInfo) []*pfs.Commit {
			return c.ChildCommits
		})
		for _, nc := range newChildren {
			childSet[oldCommitKey(nc)] = nc
			childToNewParents[nc] = anctsr
		}
		ancstrInfo.ChildCommits = nil
		for _, c := range childSet {
			ancstrInfo.ChildCommits = append(ancstrInfo.ChildCommits, c)
		}
	}
	for _, ci := range realAncestors {
		if err := resetOldCommitInfo(ctx, tx, ci); err != nil {
			return errors.Wrapf(err, "reset real ancestor %q", oldCommitKey(ci.Commit))
		}
	}
	for child, parent := range childToNewParents {
		if err := updateOldCommit(ctx, tx, child, func(ci *pfs.CommitInfo) {
			ci.ParentCommit = parent
		}); err != nil {
			return errors.Wrapf(err, "update child commit %q", oldCommitKey(child))
		}
	}
	// now write out realAncestors and childToNewParents
	for _, ci := range deleteCommits {
		ancstr := oldestAncestor(ci, deleteCommits)
		// restoreLinks=false because we're already resetting parent/children relationships above
		if err := deleteCommit(ctx, tx, ci, ancstr, false /* restoreLinks */); err != nil {
			return errors.Wrapf(err, "delete real commit %q with ancestor", oldCommitKey(ci.Commit), oldCommitKey(ancstr))
		}
	}
	// now set ci.DirectProvenance for each commit
	if cis, err = listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{}); err != nil {
		return errors.Wrap(err, "recollect commits")
	}
	for _, ci := range cis {
		var err error
		ci.DirectProvenance, err = commitProvenance(tx, ci.Commit.Repo, ci.Commit.Branch, ci.Commit.ID)
		if err != nil {
			return err
		}
		if err := updateCollectionProto(ctx, tx, "commits", oldCommitKey(ci.Commit), oldCommitKey(ci.Commit), ci); err != nil {
			return errors.Wrapf(err, "update DirectProvenance for %q", oldCommitKey(ci.Commit))
		}
	}
	// re-point branches if necessary
	headlessBranches := make([]*pfs.BranchInfo, 0)
	if err := forEachCollectionProtos(ctx, tx, "branches", &pfs.BranchInfo{}, func(bi *pfs.BranchInfo) {
		if _, ok := deleteCommits[oldCommitKey(bi.Head)]; ok {
			headlessBranches = append(headlessBranches, proto.Clone(bi).(*pfs.BranchInfo))
		}
	}); err != nil {
		return errors.Wrap(err, "record headless branches")
	}
	for _, bi := range headlessBranches {
		prevHead := deleteCommits[oldCommitKey(bi.Head)]
		if err := updateOldBranch(ctx, tx, bi.Branch, func(bi *pfs.BranchInfo) {
			bi.Head = oldestAncestor(prevHead, deleteCommits)
		}); err != nil {
			return errors.Wrap(err, "update headless branches")
		}
	}
	return nil
}

func commitProvenance(tx *pachsql.Tx, repo *pfs.Repo, branch *pfs.Branch, commitSet string) ([]*pfs.Commit, error) {
	commitKey := oldCommitKey(&pfs.Commit{
		Repo:   repo,
		Branch: branch,
		ID:     commitSet,
	})
	query := `SELECT commit_id FROM pfs.commits 
                  WHERE int_id IN (       
                      SELECT to_id FROM pfs.commits JOIN pfs.commit_provenance 
                        ON int_id = from_id 
                      WHERE commit_id = $1
                  );`
	rows, err := tx.Queryx(query, commitKey)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	commitProvenance := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commitId string
		if err := rows.Scan(&commitId); err != nil {
			return nil, errors.EnsureStack(err)
		}
		c := &pfs.Commit{
			Repo: &pfs.Repo{
				Name: strings.Split(strings.Split(commitId, "/")[1], "@")[0],
				Project: &pfs.Project{
					Name: strings.Split(commitId, "/")[0],
				},
			},
			ID: strings.Split(commitId, "=")[1],
		}
		commitProvenance = append(commitProvenance, c)
	}
	return commitProvenance, nil
}

func oldestAncestor(startCommit *pfs.CommitInfo, skipSet map[string]*pfs.CommitInfo) *pfs.Commit {
	oldest := traverseToEdges(startCommit, skipSet, func(commitInfo *pfs.CommitInfo) []*pfs.Commit {
		return []*pfs.Commit{commitInfo.ParentCommit}
	})
	if len(oldest) == 0 {
		return nil
	}
	return oldest[0]
}

// traverseToEdges does a breadth first search using a traverse function.
// returns all of the commits that are at the edge of "skipSet"
func traverseToEdges(startCommit *pfs.CommitInfo, skipSet map[string]*pfs.CommitInfo, traverse func(*pfs.CommitInfo) []*pfs.Commit) []*pfs.Commit {
	cs := []*pfs.Commit{startCommit.Commit}
	result := make([]*pfs.Commit, 0)
	var c *pfs.Commit
	for len(cs) > 0 {
		c, cs = cs[0], cs[1:]
		if c == nil {
			continue
		}
		ci, ok := skipSet[oldCommitKey(c)]
		if ok {
			cs = append(cs, traverse(ci)...)
		} else {
			result = append(result, proto.Clone(c).(*pfs.Commit))
		}
	}
	return result
}

func commitSubvenance(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) ([]*pfs.Commit, error) {
	key := oldCommitKey(c)
	query := `SELECT commit_id, branch FROM pfs.commits WHERE int_id IN (       
            SELECT to_id FROM pfs.commits JOIN pfs.commit_provenance ON int_id = to_id WHERE commit_id = $1
        );`
	rows, err := tx.QueryxContext(ctx, query, key)
	if err != nil {
		return nil, errors.Wrapf(err, "query commit subvenance for commit %q", c)
	}
	commitSubvenance := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commitId, branch string
		if err := rows.Scan(&commitId, &branch); err != nil {
			return nil, errors.EnsureStack(err)
		}
		project, rest, _ := strings.Cut(commitId, "/")
		repoName, rest, _ := strings.Cut(rest, ".")
		repoType, rest, _ := strings.Cut(rest, "@")
		branch, id, _ := strings.Cut(rest, "=")
		repo := &pfs.Repo{
			Name: repoName,
			Type: repoType,
			Project: &pfs.Project{
				Name: project,
			},
		}
		c := &pfs.Commit{
			Repo:   repo,
			Branch: &pfs.Branch{Name: branch, Repo: repo},
			ID:     id,
		}
		commitSubvenance = append(commitSubvenance, c)
	}
	return commitSubvenance, nil
}

// since the provenance relationship points from downstream to upstream, 'from' represents the commit subvenant to the 'to' commit
func deleteCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to *pfs.Commit) error {
	fromId, err := getCommitIntId(ctx, tx, from)
	if err != nil {
		return errors.Wrapf(err, "get commit int id for 'from' commit %q", from)
	}
	toId, err := getCommitIntId(ctx, tx, to)
	if err != nil {
		return errors.Wrapf(err, "get commit int id for 'to' commit %q", to)
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM pfs.commit_provenance WHERE from_id=$1 AND to_id=$2;`, fromId, toId)
	return errors.EnsureStack(err)
}

// since the provenance relationship points from downstream to upstream, 'from' represents the commit subvenant to the 'to' commit
func addCommitProvenance(ctx context.Context, tx *pachsql.Tx, from, to *pfs.Commit) error {
	fromId, err := getCommitIntId(ctx, tx, from)
	if err != nil {
		return err
	}
	toId, err := getCommitIntId(ctx, tx, to)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO pfs.commit_provenance(from_id, to_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, fromId, toId)
	return errors.EnsureStack(err)
}

// the goal of this migration is to migrate commits to be unqiuely indentified by a (repo, UUID), instead of by (repo, branch, UUID)
//
// 1) commits with equal (repo, UUID) pairs must be de-duplicated. This is primarily expected to happen in deferred processing cases.
// In such cases, the deferred branch's commit is expected to be a child commit of the leading branch's commit. So to make the substitution,
// we can simply delete the deferred branch's commit and re-point it's direct subvenant commits to point to the master commit as their
// direct provenance commits.
//
// 2) assert that we didn't miss any duplicate commits, and error the migration if we did.
// TODO(acohen4): Write test to catch this
//
// 3) commit keys can now be substituted from <project>/<repo>@<branch>=<id> -> <project>/<repo>@<id>
func migrateToBranchlessCommits(ctx context.Context, tx *pachsql.Tx) error {
	// 1) handle deferred branches
	var cis []*pfs.CommitInfo
	var err error
	if cis, err = listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{}); err != nil {
		return err
	}
	for _, ci := range cis {
		// TODO(acohen4): improve this condition check?
		if ci.ParentCommit != nil && branchKey(ci.ParentCommit.Branch) != branchKey(ci.Commit.Branch) {
			subv, err := commitSubvenance(ctx, tx, ci.Commit)
			if err != nil {
				return err
			}
			for _, sc := range subv {
				if err := deleteCommitProvenance(ctx, tx, sc, ci.Commit); err != nil {
					return errors.Wrapf(err, "delete commit provenance for %q", oldCommitKey(ci.Commit))
				}
				if err := addCommitProvenance(ctx, tx, sc, ci.ParentCommit); err != nil {
					return errors.Wrapf(err, "add commit provenance for %q", oldCommitKey(ci.ParentCommit))
				}
			}
			if err := deleteCommit(ctx, tx, ci, ci.ParentCommit, true /* restoreLinks */); err != nil {
				return errors.Wrapf(err, "delete commit %q", oldCommitKey(ci.Commit))
			}
			// if ci was the head of a branch, point the head to ci.Parent
			if ci.Commit.Branch != nil {
				bk := branchKey(ci.Commit.Branch)
				bi := &pfs.BranchInfo{}
				if err := getCollectionProto(ctx, tx, "branches", bk, bi); err != nil {
					return errors.Wrapf(err, "get branch info for %q", bk)
				}
				if bi.Head.ID == ci.Commit.ID {
					bi.Head = ci.ParentCommit
					if err := updateCollectionProto(ctx, tx, "branches", bk, bk, bi); err != nil {
						return errors.Wrapf(err, "update branch info for %q", bk)
					}
				}
			}
		}
	}
	// 2) migrate commit keys
	oldCommits, err := listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{})
	if err != nil {
		return err
	}
	for _, oldCommit := range oldCommits {
		if oldCommit.Commit.Repo == nil {
			oldCommit.Commit.Repo = oldCommit.Commit.Branch.Repo
		}
		if oldCommit.ParentCommit != nil && oldCommit.ParentCommit.Repo == nil {
			oldCommit.ParentCommit.Repo = oldCommit.ParentCommit.Branch.Repo
		}
		for _, child := range oldCommit.ChildCommits {
			if child.Repo == nil {
				child.Repo = child.Branch.Repo
			}
		}
		data, err := proto.Marshal(oldCommit)
		if err != nil {
			return errors.EnsureStack(err)
		}
		params := make(map[string]interface{})
		params["oldKey"] = oldCommitKey(oldCommit.Commit)
		params["proto"] = data
		params["newKey"] = commitBranchlessKey(oldCommit.Commit)
		stmt := fmt.Sprintf("update pfs.commits set commit_id = :newKey where commit_id = :oldKey")
		_, err = tx.NamedExecContext(ctx, stmt, params)
		if err != nil {
			return errors.Wrapf(err, "update pfs.commits")
		}
		stmt = fmt.Sprintf("update collections.commits set key = :newKey, proto = :proto where key = :oldKey")
		_, err = tx.NamedExecContext(ctx, stmt, params)
		if err != nil {
			return errors.Wrapf(err, "update collections.commits")
		}
	}
	updatedBis := make([]*pfs.BranchInfo, 0)
	if err := forEachCollectionProtos(ctx, tx, "branches", &pfs.BranchInfo{}, func(bi *pfs.BranchInfo) {
		bi.Head.Repo = bi.Branch.Repo
		updatedBis = append(updatedBis, proto.Clone(bi).(*pfs.BranchInfo))
	}); err != nil {
		return errors.Wrap(err, "record headless branches")
	}
	for _, bi := range updatedBis {
		if err := updateCollectionProto(ctx, tx, "branches", branchKey(bi.Branch), branchKey(bi.Branch), bi); err != nil {
			return errors.Wrapf(err, "could not migrate branch %q", branchKey(bi.Branch))
		}
	}
	if cis, err = listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{}); err != nil {
		return err
	}
	for _, ci := range cis {
		oldKey := oldCommitKey(ci.Commit)
		newKey := commitBranchlessKey(ci.Commit)
		if _, err := tx.ExecContext(ctx, `UPDATE pfs.commits SET commit_id=$1 WHERE commit_id=$2;`, newKey, oldKey); err != nil {
			return errors.Wrapf(err, "update pfs.commits for old key %q and new key %q", oldKey, newKey)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_totals SET commit_id=$1 WHERE commit_id=$2;`, newKey, oldKey); err != nil {
			return errors.Wrapf(err, "update pfs.commit_totals for old key %q and new key %q", oldKey, newKey)
		}
		if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_diffs SET commit_id=$1 WHERE commit_id=$2;`, newKey, oldKey); err != nil {
			return errors.Wrapf(err, "update pfs.commit_diffs for old key %q and new key %q", oldKey, newKey)
		}
		fromPattern := fmt.Sprintf("^commit/%v/(.*)", oldKey)
		toPattern := fmt.Sprintf("commit/%v/\\1", newKey)
		if _, err := tx.ExecContext(ctx, `UPDATE storage.tracker_objects SET str_id = regexp_replace(str_id, $1, $2);`, fromPattern, toPattern); err != nil {
			return errors.Wrapf(err, "update storage.tracker_objects for old key %q and new key %q", oldKey, newKey)
		}
	}
	return nil
}
