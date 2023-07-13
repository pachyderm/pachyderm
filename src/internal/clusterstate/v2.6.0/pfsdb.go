package v2_6_0

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	v2_5_0 "github.com/pachyderm/pachyderm/v2/src/internal/clusterstate/v2.5.0"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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

func updateOldCommit_V2_5(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit, f func(ci *v2_5_0.CommitInfo)) error {
	k := oldCommitKey(c)
	ci := &v2_5_0.CommitInfo{}
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

func resetOldCommitInfo(ctx context.Context, tx *pachsql.Tx, ci *v2_5_0.CommitInfo) error {
	data, err := proto.Marshal(ci)
	if err != nil {
		return errors.Wrapf(err, "unmarshal old commit %q", oldCommitKey(ci.Commit))
	}
	_, err = tx.ExecContext(ctx, `UPDATE collections.commits SET proto=$1  WHERE key=$2;`, data, oldCommitKey(ci.Commit))
	return errors.Wrapf(err, "reset old commit info %q", oldCommitKey(ci.Commit))
}

func convertCommitInfoToV2_6_0(ci *v2_5_0.CommitInfo) *pfs.CommitInfo {
	ci.Commit.Repo = ci.Commit.Branch.Repo
	return &pfs.CommitInfo{
		Commit:              ci.Commit,
		Origin:              ci.Origin,
		Description:         ci.Description,
		ParentCommit:        ci.ParentCommit,
		ChildCommits:        ci.ChildCommits,
		Started:             ci.Started,
		Finishing:           ci.Finishing,
		Finished:            ci.Finished,
		Error:               ci.Error,
		SizeBytesUpperBound: ci.SizeBytesUpperBound,
		Details:             ci.Details,
	}
}

func validateExistingDAGs(cis []*v2_5_0.CommitInfo) error {
	// group duplicate commits by branchless key
	duplicates := make(map[string]map[string]*v2_5_0.CommitInfo) // branchless commit key -> { old commit key ->  commit info }
	for _, ci := range cis {
		if _, ok := duplicates[commitBranchlessKey(ci.Commit)]; !ok {
			duplicates[commitBranchlessKey(ci.Commit)] = make(map[string]*v2_5_0.CommitInfo)
		}
		duplicates[commitBranchlessKey(ci.Commit)][oldCommitKey(ci.Commit)] = ci
	}
	// the only duplicate commits we allow and handle in the migration are those connected through direct ancestry.
	// the allowed duplicate commits are expected to arise from branch triggers / deferred processing.
	var badCommitSets []string
	for _, dups := range duplicates {
		if len(dups) <= 1 {
			continue
		}
		seen := make(map[string]struct{})
		var commitSet string
		for _, d := range dups {
			commitSet = d.Commit.Id
			if d.ParentCommit != nil {
				if _, ok := dups[oldCommitKey(d.ParentCommit)]; ok {
					seen[oldCommitKey(d.Commit)] = struct{}{}
				}
			}
		}
		if len(dups)-len(seen) > 1 {
			badCommitSets = append(badCommitSets, commitSet)
		}
	}
	if badCommitSets != nil {
		return errors.Errorf("invalid commit sets detected: %v", badCommitSets)
	}
	return nil
}

func removeAliasCommits(ctx context.Context, tx *pachsql.Tx) error {
	var oldCIs []*v2_5_0.CommitInfo
	var err error
	// TODO(provenance): this function is doing more than just removing aliases
	// first fill commit provenance to all the existing commit sets
	if oldCIs, err = listCollectionProtos(ctx, tx, "commits", &v2_5_0.CommitInfo{}); err != nil {
		return errors.Wrap(err, "populate the pfs.commits table")
	}
	for i, ci := range oldCIs {
		log.Info(ctx, "add old commit to pfs.commits",
			zap.String("commit", fmt.Sprintf("%s@%s=%s",
				repoKey(ci.Commit.Branch.Repo),
				ci.Commit.Branch.GetName(),
				ci.Commit.Id)),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(oldCIs))),
		)
		if err := addCommit(ctx, tx, ci.Commit); err != nil {
			return err
		}
	}
	oldCIMap := make(map[string]struct{})
	for _, ci := range oldCIs {
		oldCIMap[oldCommitKey(ci.Commit)] = struct{}{}
	}
	deleteCommits := make(map[string]*v2_5_0.CommitInfo)
	for i, ci := range oldCIs {
		log.Info(ctx, "branch provenance for commit",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.Any("branch provenance", ci.DirectProvenance),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(oldCIs))),
		)
		for j, b := range ci.DirectProvenance {
			log.Info(ctx, "add commit provenance",
				zap.String("from", oldCommitKey(ci.Commit)),
				zap.String("to", oldCommitKey(b.NewCommit(ci.Commit.Id))),
				zap.String("progress", fmt.Sprintf("%v/%v", j, len(ci.DirectProvenance))),
			)
			// if the provenant commit's repo was deleted, it's reference may
			// still exist in the old provenance, so we skip mapping it over to the new model
			if _, ok := oldCIMap[oldCommitKey(b.NewCommit(ci.Commit.Id))]; !ok {
				continue
			}
			if err := addCommitProvenance(ctx, tx, ci.Commit, b.NewCommit(ci.Commit.Id)); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q",
					oldCommitKey(ci.Commit),
					oldCommitKey(b.NewCommit(ci.Commit.Id)))
			}
		}
		// collect all the ALIAS commits to be deleted
		// before 2.6, pfs.OriginKind_ALIAS = 4
		if ci.Origin.Kind == 4 {
			deleteCommits[oldCommitKey(ci.Commit)] = proto.Clone(ci).(*v2_5_0.CommitInfo)
		}
	}
	var count int
	for _, ci := range deleteCommits {
		log.Info(ctx, "validating deleted alias commit",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.String("progress", fmt.Sprintf("%v/%v", count, len(deleteCommits))),
		)
		same, err := sameFileSets(ctx, tx, ci.Commit, ci.ParentCommit)
		if err != nil {
			return err
		}
		if !same {
			return errors.Errorf("commit %q is listed as ALIAS but has a different ID than its first real ancestor",
				oldCommitKey(ci.Commit))
		}
		count++
	}
	realAncestors := make(map[string]*v2_5_0.CommitInfo)
	childToNewParents := make(map[*pfs.Commit]*pfs.Commit)
	count = 0
	for _, ci := range deleteCommits {
		log.Info(ctx, "computing parent / children after deleted alias commit",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.String("progress", fmt.Sprintf("%v/%v", count, len(deleteCommits))),
		)
		anctsr := oldestAncestor(ci, deleteCommits)
		if _, ok := realAncestors[oldCommitKey(anctsr)]; !ok {
			ci := &v2_5_0.CommitInfo{}
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
		newChildren := traverseToEdges(ci, deleteCommits, func(c *v2_5_0.CommitInfo) []*pfs.Commit {
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
		count++
	}
	count = 0
	for _, ci := range realAncestors {
		log.Info(ctx, "updating children after deleted alias commit",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.String("progress", fmt.Sprintf("%v/%v", count, len(realAncestors))),
		)
		if err := resetOldCommitInfo(ctx, tx, ci); err != nil {
			return errors.Wrapf(err, "reset real ancestor %q", oldCommitKey(ci.Commit))
		}
		count++
	}
	count = 0
	for child, parent := range childToNewParents {
		log.Info(ctx, "updating parent after deleted alias commit",
			zap.String("commit", oldCommitKey(child)),
			zap.String("progress", fmt.Sprintf("%v/%v", count, len(childToNewParents))),
		)
		if err := updateOldCommit_V2_5(ctx, tx, child, func(ci *v2_5_0.CommitInfo) {
			ci.ParentCommit = parent
		}); err != nil {
			return errors.Wrapf(err, "update child commit %q", oldCommitKey(child))
		}
		count++
	}
	// now write out realAncestors and childToNewParents
	count = 0
	for _, ci := range deleteCommits {
		log.Info(ctx, "deleting alias commit",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.String("progress", fmt.Sprintf("%v/%v", count, len(deleteCommits))),
		)
		ancstr := oldestAncestor(ci, deleteCommits)
		// passing restoreLinks=false because we're already resetting parent/children relationships above
		if err := deleteCommit(ctx, tx, ci.Commit, ancstr); err != nil {
			return errors.Wrapf(err, "delete real commit %q with ancestor", oldCommitKey(ci.Commit), oldCommitKey(ancstr))
		}
		count++
	}
	// now set ci.DirectProvenance for each commit
	if oldCIs, err = listCollectionProtos(ctx, tx, "commits", &v2_5_0.CommitInfo{}); err != nil {
		return errors.Wrap(err, "recollect commits")
	}
	// re-write V2_5 commits to V2_6
	for i, oldCI := range oldCIs {
		var err error
		ci := convertCommitInfoToV2_6_0(oldCI)
		ci.DirectProvenance, err = commitProvenance(tx, ci.Commit.Repo, ci.Commit.Branch, ci.Commit.Id)
		log.Info(ctx, "collect commit provenance",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.Any("provenance", ci.DirectProvenance),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(oldCIs))),
		)
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

func deleteCommit(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit, ancestor *pfs.Commit) error {
	key := oldCommitKey(c)
	cId, err := getCommitIntId(ctx, tx, c)
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
		return errors.Wrapf(err, "update commit 'from' provenance for commit %q", key)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO pfs.commit_provenance (from_id, to_id)
                                          SELECT from_id, $1 FROM pfs.commit_provenance WHERE to_id=$2 ON CONFLICT DO NOTHING;`,
		ancestorId, cId); err != nil {
		return errors.Wrapf(err, "update commit 'to' provenance for commit %q", key)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commit_provenance WHERE from_id=$1 OR to_id=$1;`, cId); err != nil {
		return errors.Wrapf(err, "update commit 'to' provenance for commit %q", key)
	}
	// delete commit
	if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commits WHERE int_id=$1;`, cId); err != nil {
		return errors.Wrapf(err, "delete from pfs.commits for commit %q", key)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM collections.commits WHERE key=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from collections.commits for commit %q", key)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM pfs.commit_diffs WHERE commit_id=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from pfs.commit_diffs for commit_id %q", key)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM pfs.commit_totals WHERE commit_id=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from pfs.commit_totals for commit_id %q", key)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM storage.tracker_objects WHERE str_id LIKE $1;`, "commit/"+key+"%"); err != nil {
		return errors.Wrapf(err, "delete from pfs.commit_totals for commit_id %q", key)
	}
	return nil
}

func getCommitIntId(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (int, error) {
	rows, err := tx.QueryxContext(ctx, "SELECT int_id FROM pfs.commits WHERE commit_id=$1", oldCommitKey(commit))
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	defer rows.Close()
	var id int
	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return 0, errors.Wrapf(err, "scan int_id for commit_id %q", oldCommitKey(commit))
		}
	}
	if err := rows.Err(); err != nil {
		return 0, errors.EnsureStack(err)
	}
	return id, nil
}

func sameFileSets(ctx context.Context, tx *pachsql.Tx, c1 *pfs.Commit, c2 *pfs.Commit) (bool, error) {
	md1, err := getFileSetMd(ctx, tx, c1)
	if err != nil {
		return false, err
	}
	md2, err := getFileSetMd(ctx, tx, c2)
	if err != nil {
		return false, err
	}
	if md1 == nil || md2 == nil {
		// the semantics here are a little odd - if either commit doesn't have a fileset yet,
		// we assume that they aren't different and are therefore the same.
		return true, nil
	}
	if md1.GetPrimitive() != nil && md2.GetPrimitive() != nil {
		ser1, err := proto.Marshal(md1.GetPrimitive())
		if err != nil {
			return false, errors.Wrap(err, "marshal first primitive fileset")
		}
		ser2, err := proto.Marshal(md2.GetPrimitive())
		if err != nil {
			return false, errors.Wrap(err, "marshal second primitive fileset")
		}
		return bytes.Equal(ser1, ser2), nil
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

func getFileSetMd(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) (*fileset.Metadata, error) {
	var mdData []byte
	if err := tx.GetContext(ctx, &mdData, `SELECT metadata_pb FROM storage.filesets JOIN pfs.commit_totals ON id = fileset_id WHERE commit_id = $1`, oldCommitKey(c)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, errors.EnsureStack(err)
	}
	md := &fileset.Metadata{}
	if err := proto.Unmarshal(mdData, md); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return md, nil
}

func commitProvenance(tx *pachsql.Tx, repo *pfs.Repo, branch *pfs.Branch, commitSet string) ([]*pfs.Commit, error) {
	commitKey := oldCommitKey(&pfs.Commit{
		Repo:   repo,
		Branch: branch,
		Id:     commitSet,
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
	defer rows.Close()
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
			Id: strings.Split(commitId, "=")[1],
		}
		commitProvenance = append(commitProvenance, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return commitProvenance, nil
}

func addCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO pfs.commits(commit_id, commit_set_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;`, oldCommitKey(commit), commit.Id)
	if err != nil {
		return errors.Wrapf(err, "insert commit %q", oldCommitKey(commit))
	}
	var commitIntId int
	query := `SELECT int_id FROM pfs.commits WHERE commit_id = $1;`
	err = tx.GetContext(ctx, &commitIntId, query, oldCommitKey(commit))
	return errors.Wrapf(err, "get int_id of commit %q", oldCommitKey(commit))
}

func oldestAncestor(startCommit *v2_5_0.CommitInfo, skipSet map[string]*v2_5_0.CommitInfo) *pfs.Commit {
	oldest := traverseToEdges(startCommit, skipSet, func(commitInfo *v2_5_0.CommitInfo) []*pfs.Commit {
		return []*pfs.Commit{commitInfo.ParentCommit}
	})
	if len(oldest) == 0 {
		return nil
	}
	return oldest[0]
}

// traverseToEdges does a breadth first search using a traverse function.
// returns all of the commits that are at the edge of "skipSet"
func traverseToEdges(startCommit *v2_5_0.CommitInfo, skipSet map[string]*v2_5_0.CommitInfo, traverse func(*v2_5_0.CommitInfo) []*pfs.Commit) []*pfs.Commit {
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

// the goal of this migration is to migrate commits to be unqiuely indentified by a (repo, UUID), instead of by (repo, branch, UUID)
//
// 1) commits with equal (repo, UUID) pairs must be de-duplicated. This is primarily expected to happen in deferred processing cases.
// In such cases, the deferred branch's commit is expected to be a child commit of the leading branch's commit. So to make the substitution,
// we can simply delete the deferred branch's commit and re-point it's direct subvenant commits to point to the master commit as their
// direct provenance commits.
//
// 2) assert that we didn't miss any duplicate commits, and error the migration if we did.
//
// 3) commit keys can now be substituted from <project>/<repo>@<branch>=<id> -> <project>/<repo>@<id>
func branchlessCommitsPFS(ctx context.Context, tx *pachsql.Tx) error {
	// 1) handle deferred branches
	var cis []*pfs.CommitInfo
	var err error
	if cis, err = listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{}); err != nil {
		return err
	}
	for i, ci := range cis {
		log.Info(ctx, "view commit provenance",
			zap.String("commit", oldCommitKey(ci.Commit)),
			zap.Any("branch provenance", ci.DirectProvenance),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(cis))),
		)
		// TODO(provenance): is this deferred/trigger handling correct?
		if ci.ParentCommit != nil && branchKey(ci.ParentCommit.Branch) != branchKey(ci.Commit.Branch) {
			log.Info(ctx, fmt.Sprintf("deduplicate old deferred processing commit: %s", oldCommitKey(ci.Commit)),
				zap.String("commit branch", branchKey(ci.Commit.Branch)),
				zap.String("parent branch", branchKey(ci.ParentCommit.Branch)))
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
			if err := deleteCommit(ctx, tx, ci.Commit, ci.ParentCommit); err != nil {
				return errors.Wrapf(err, "delete commit %q", oldCommitKey(ci.Commit))
			}
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
			// if ci was the head of a branch, point the head to ci.Parent
			if ci.Commit.Branch != nil {
				bk := branchKey(ci.Commit.Branch)
				bi := &pfs.BranchInfo{}
				if err := getCollectionProto(ctx, tx, "branches", bk, bi); err != nil {
					return errors.Wrapf(err, "get branch info for %q", bk)
				}
				if bi.Head.Id == ci.Commit.Id {
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
	for i, oldCommit := range oldCommits {
		log.Info(ctx, "removing branch from commit",
			zap.String("commit", oldCommitKey(oldCommit.Commit)),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(oldCommits))),
		)
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
		stmt := "update pfs.commits set commit_id = :newKey where commit_id = :oldKey"
		_, err = tx.NamedExecContext(ctx, stmt, params)
		if err != nil {
			return errors.Wrapf(err, "update pfs.commits")
		}
		stmt = "update collections.commits set key = :newKey, proto = :proto where key = :oldKey"
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
	return branchlessCommitKeysPFS(ctx, tx)
}

// map <project>/<repo>@<branch>=<id> -> <project>/<repo>@<id>; basically replace '@[-a-zA-Z0-9_]+)=' -> '@'
func branchlessCommitKeysPFS(ctx context.Context, tx *pachsql.Tx) (retErr error) {
	ctx, end := log.SpanContext(ctx, "branchlessCommitKeysPFS")
	defer end(log.Errorp(&retErr))
	log.Info(ctx, "removing branch from pfs.commits")
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commits SET commit_id = regexp_replace(commit_id, '@(.+)=', '@');`); err != nil {
		return errors.Wrapf(err, "update pfs.commits to branchless commit ids")
	}
	log.Info(ctx, "removing branch from pfs.commit_totals")
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_totals SET commit_id = regexp_replace(commit_id, '@(.+)=', '@');`); err != nil {
		return errors.Wrap(err, "update pfs.commit_totals to branchless commit ids")
	}
	log.Info(ctx, "removing branch from pfs.commit_diffs")
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_diffs SET commit_id = regexp_replace(commit_id, '@(.+)=', '@');`); err != nil {
		return errors.Wrap(err, "update pfs.commit_diffs to branchless commits ids")
	}
	log.Info(ctx, "removing branch from storage.tracker_objects")
	if _, err := tx.ExecContext(ctx, `UPDATE storage.tracker_objects SET str_id = regexp_replace(str_id, '@(.+)=', '@') WHERE str_id LIKE 'commit/%';`); err != nil {
		return errors.Wrap(err, "update storage.tracker_objects to branchless commit ids")
	}
	return nil
}

func commitSubvenance(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) ([]*pfs.Commit, error) {
	key := oldCommitKey(c)
	query := `SELECT commit_id FROM pfs.commits WHERE int_id IN (
            SELECT to_id FROM pfs.commits JOIN pfs.commit_provenance ON int_id = to_id WHERE commit_id = $1
        );`
	rows, err := tx.QueryxContext(ctx, query, key)
	if err != nil {
		return nil, errors.Wrapf(err, "query commit subvenance for commit %q", c)
	}
	defer rows.Close()
	commitSubvenance := make([]*pfs.Commit, 0)
	for rows.Next() {
		var commitId string
		if err := rows.Scan(&commitId); err != nil {
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
			Id:     id,
		}
		commitSubvenance = append(commitSubvenance, c)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.EnsureStack(err)
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

func deleteDanglingCommitRefs(ctx context.Context, tx *pachsql.Tx) error {
	log.Info(ctx, "checking for dangling commits")
	parseRepo := func(key string) *pfs.Repo {
		slashSplit := strings.Split(key, "/")
		dotSplit := strings.Split(slashSplit[1], ".")
		return &pfs.Repo{
			Project: &pfs.Project{Name: slashSplit[0]},
			Name:    dotSplit[0],
			Type:    dotSplit[1],
		}
	}
	parseBranch := func(key string) *pfs.Branch {
		split := strings.Split(key, "@")
		return &pfs.Branch{
			Repo: parseRepo(split[0]),
			Name: split[1],
		}
	}
	listRepoKeys := func(tx *pachsql.Tx) (map[string]struct{}, error) {
		var keys []string
		if err := tx.Select(&keys, `SELECT key FROM collections.repos`); err != nil {
			return nil, errors.Wrap(err, "select keys from collections.repos")
		}
		rs := make(map[string]struct{})
		for _, k := range keys {
			rs[k] = struct{}{}
		}
		return rs, nil
	}
	parseCommit_2_5 := func(key string) (*pfs.Commit, error) {
		split := strings.Split(key, "=")
		if len(split) != 2 {
			return nil, errors.Errorf("parsing commit key with 2.6.x+ structure %q", key)
		}
		b := parseBranch(split[0])
		return &pfs.Commit{
			Repo:   b.Repo,
			Branch: b,
			Id:     split[1],
		}, nil
	}
	listReferencedCommits := func(tx *pachsql.Tx) (map[string]*pfs.Commit, error) {
		cs := make(map[string]*pfs.Commit)
		var err error
		var ids []string
		if err := tx.Select(&ids, `SELECT commit_id from  pfs.commit_totals`); err != nil {
			return nil, errors.Wrap(err, "select commit ids from pfs.commit_totals")
		}
		for _, id := range ids {
			cs[id], err = parseCommit_2_5(id)
			if err != nil {
				return nil, err
			}
		}
		ids = make([]string, 0)
		if err := tx.Select(&ids, `SELECT commit_id from  pfs.commit_diffs`); err != nil {
			return nil, errors.Wrap(err, "select commit ids from pfs.commit_diffs")
		}
		for _, id := range ids {
			cs[id], err = parseCommit_2_5(id)
			if err != nil {
				return nil, err
			}
		}
		return cs, nil
	}
	cs, err := listReferencedCommits(tx)
	if err != nil {
		return errors.Wrap(err, "list referenced commits")
	}
	rs, err := listRepoKeys(tx)
	if err != nil {
		return errors.Wrap(err, "list repos")
	}
	var dangCommitKeys []string
	for _, c := range cs {
		if _, ok := rs[repoKey(c.Repo)]; !ok {
			dangCommitKeys = append(dangCommitKeys, oldCommitKey(c))
		}
	}
	if len(dangCommitKeys) > 0 {
		log.Info(ctx, "detected dangling commit references", zap.Any("references", dangCommitKeys))
	}
	for i, id := range dangCommitKeys {
		log.Info(ctx, "deleting dangling commit",
			zap.String("commit", id),
			zap.String("progress", fmt.Sprintf("%v/%v", i, len(dangCommitKeys))),
		)
		if _, err := tx.Exec(`DELETE FROM pfs.commit_totals WHERE commit_id = $1`, id); err != nil {
			return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_totals", id)
		}
		if _, err := tx.Exec(`DELETE FROM pfs.commit_diffs WHERE commit_id = $1`, id); err != nil {
			return errors.Wrapf(err, "delete dangling commit reference %q from pfs.commit_diffs", id)
		}
	}
	return nil
}
