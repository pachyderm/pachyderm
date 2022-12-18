package v2_5_0

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

var reposTypeIndex = &index{
	Name: "type",
	Extract: func(val proto.Message) string {
		return val.(*pfs.RepoInfo).Repo.Type
	},
}

func reposNameKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name
}

var reposNameIndex = &index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return reposNameKey(val.(*pfs.RepoInfo).Repo)
	},
}

var reposIndexes = []*index{reposNameIndex, reposTypeIndex}

func repoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

func repoKeyCheck(key string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 || len(parts[1]) == 0 {
		return errors.Errorf("repo must have a specified type")
	}
	return nil
}

var commitsRepoIndex = &index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return repoKey(val.(*pfs.CommitInfo).Commit.Branch.Repo)
	},
}

var commitsBranchlessIndex = &index{
	Name: "branchless",
	Extract: func(val proto.Message) string {
		return commitBranchlessKey(val.(*pfs.CommitInfo).Commit)
	},
}

var commitsCommitSetIndex = &index{
	Name: "commitset",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.ID
	},
}

var commitsIndexes = []*index{commitsRepoIndex, commitsBranchlessIndex, commitsCommitSetIndex}

func commitKey(commit *pfs.Commit) string {
	return branchKey(commit.Branch) + "=" + commit.ID
}

func commitBranchlessKey(commit *pfs.Commit) string {
	return repoKey(commit.Branch.Repo) + "@" + commit.ID
}

var branchesRepoIndex = &index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return repoKey(val.(*pfs.BranchInfo).Branch.Repo)
	},
}

var branchesIndexes = []*index{branchesRepoIndex}

func branchKey(branch *pfs.Branch) string {
	return repoKey(branch.Repo) + "@" + branch.Name
}

func pfsCollections() []*postgresCollection {
	return []*postgresCollection{
		newPostgresCollection("projects", nil),
	}
}

func migrateProject(p *pfs.Project) *pfs.Project {
	if p == nil || p.Name == "" {
		return &pfs.Project{Name: "default"}
	}
	return p
}

func migrateRepo(r *pfs.Repo) *pfs.Repo {
	r.Project = migrateProject(r.Project)
	return r
}

func migrateRepoInfo(r *pfs.RepoInfo) *pfs.RepoInfo {
	r.Repo = migrateRepo(r.Repo)
	for i, b := range r.Branches {
		r.Branches[i] = migrateBranch(b)
	}
	return r
}

func migrateBranch(b *pfs.Branch) *pfs.Branch {
	b.Repo = migrateRepo(b.Repo)
	return b
}

func migrateBranchInfo(b *pfs.BranchInfo) *pfs.BranchInfo {
	b.Branch = migrateBranch(b.Branch)
	if b.Head != nil {
		b.Head = migrateCommit(b.Head)
	}
	for i, bb := range b.Provenance {
		b.Provenance[i] = migrateBranch(bb)
	}
	for i, bb := range b.DirectProvenance {
		b.DirectProvenance[i] = migrateBranch(bb)
	}
	for i, bb := range b.Subvenance {
		b.Subvenance[i] = migrateBranch(bb)
	}
	return b
}

func migrateCommit(c *pfs.Commit) *pfs.Commit {
	c.Branch = migrateBranch(c.Branch)
	return c
}

func migrateCommitInfo(c *pfs.CommitInfo) *pfs.CommitInfo {
	c.Commit = migrateCommit(c.Commit)
	if c.ParentCommit != nil {
		c.ParentCommit = migrateCommit(c.ParentCommit)
	}
	for i, cc := range c.ChildCommits {
		c.ChildCommits[i] = migrateCommit(cc)
	}
	for i, bb := range c.DirectProvenance {
		c.DirectProvenance[i] = migrateBranch(bb)
	}
	return c
}

// migratePFSDB migrates PFS to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.
func migratePFSDB(ctx context.Context, tx *pachsql.Tx) error {
	// A pre-projects commit diff/total has a commit_id value which looks
	// like images.user@master=da4016a16f8944cba94038ab5bcc9933; a
	// post-projects commit diff/total has a commit_id which looks like
	// myproject/images.user@master=da4016a16f8944cba94038ab5bcc9933.
	//
	// The regexp below is matches only rows with a pre-projects–style
	// commit_id and does not touch post-projects–style rows.  This is
	// because prior to the default project name changing, commits in the
	// default project were still identified without the project (e.g. as
	// images.user@master=da4016a16f8944cba94038ab5bcc9933 rather than
	// /images.user@master=da4016a16f8944cba94038ab5bcc9933).
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_diffs SET commit_id = regexp_replace(commit_id, '^([-a-zA-Z0-9_]+)', 'default/\1') WHERE commit_id ~ '^[-a-zA-Z0-9_]+\.';`); err != nil {
		return errors.Wrap(err, "could not update")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_totals SET commit_id = regexp_replace(commit_id, '^([-a-zA-Z0-9_]+)', 'default/\1') WHERE commit_id ~ '^[-a-zA-Z0-9_]+\.';`); err != nil {
		return errors.Wrap(err, "could not update")
	}
	var oldRepo = new(pfs.RepoInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "repos", reposIndexes, oldRepo, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldRepo = migrateRepoInfo(oldRepo)
		return repoKey(oldRepo.Repo), oldRepo, nil

	},
		withKeyCheck(repoKeyCheck),
		withKeyGen(func(key interface{}) (string, error) {
			if repo, ok := key.(*pfs.Repo); !ok {
				return "", errors.New("key must be a repo")
			} else {
				return repoKey(repo), nil
			}
		}),
	); err != nil {
		return errors.Wrap(err, "could not migrate repos")
	}
	var oldBranch = new(pfs.BranchInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "branches", branchesIndexes, oldBranch, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldBranch = migrateBranchInfo(oldBranch)
		return branchKey(oldBranch.Branch), oldBranch, nil

	},
		withKeyGen(func(key interface{}) (string, error) {
			if branch, ok := key.(*pfs.Branch); !ok {
				return "", errors.New("key must be a branch")
			} else {
				return branchKey(branch), nil
			}
		}),
		withKeyCheck(func(key string) error {
			keyParts := strings.Split(key, "@")
			if len(keyParts) != 2 {
				return errors.Errorf("branch key %s isn't valid, use BranchKey to generate it", key)
			}
			if uuid.IsUUIDWithoutDashes(keyParts[1]) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return repoKeyCheck(keyParts[0])
		})); err != nil {
		return errors.Wrap(err, "could not migrate branches")
	}
	var oldCommit = new(pfs.CommitInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "commits", commitsIndexes, oldCommit, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldCommit = migrateCommitInfo(oldCommit)
		return commitKey(oldCommit.Commit), oldCommit, nil

	}, withKeyGen(func(key interface{}) (string, error) {
		if commit, ok := key.(*pfs.Commit); !ok {
			return "", errors.New("key must be a commit")
		} else {
			return commitKey(commit), nil
		}
	})); err != nil {
		return errors.Wrap(err, "could not migrate commits")
	}

	return nil
}

func forEachCommit(ctx context.Context, tx *pachsql.Tx, ci *pfs.CommitInfo, f func(string) error) error {
	rr, err := tx.QueryContext(ctx, `SELECT key, proto FROM collections.commits;`)
	if err != nil {
		return errors.Wrap(err, "could not read table")
	}
	defer rr.Close()
	for rr.Next() {
		var (
			oldKey string
			pb     []byte
		)
		if err := rr.Err(); err != nil {
			return errors.Wrap(err, "row error")
		}
		if err := rr.Scan(&oldKey, &pb); err != nil {
			return errors.Wrap(err, "could not scan row")
		}
		if err := proto.Unmarshal(pb, ci); err != nil {
			return errors.Wrapf(err, "could not unmarshal proto")
		}
		if err := f(oldKey); err != nil {
			return errors.Wrapf(err, "could not apply f()")
		}
	}
	return nil
}

func getCommitInfo(ctx context.Context, tx *pachsql.Tx, key string) (*pfs.CommitInfo, error) {
	var pb []byte
	if err := tx.GetContext(ctx, &pb, "SELECT proto FROM collections.commits WHERE key=$1", key); err != nil {
		return nil, errors.Wrapf(err, "get collections.commits with key %q", key)
	}
	ci := &pfs.CommitInfo{}
	if err := proto.Unmarshal(pb, ci); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal proto")
	}
	return ci, nil
}

func updateCommitInfo(ctx context.Context, tx *pachsql.Tx, key string, ci *pfs.CommitInfo) error {
	data, err := proto.Marshal(ci)
	if err != nil {
		return errors.Wrapf(err, "unmarshal commit info proto for commit key %q", key)
	}
	_, err = tx.ExecContext(ctx, `UPDATE collections.commits SET proto=$1  WHERE key=$2;`, data, key)
	return errors.Wrapf(err, "update collections.commits with key %q", key)
}

func updateOldCommit(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit, f func(ci *pfs.CommitInfo)) error {
	k := commitKey(c)
	ci, err := getCommitInfo(ctx, tx, k)
	if err != nil {
		return errors.Wrapf(err, "retrieve old commit %q", k)
	}
	f(ci)
	return updateCommitInfo(ctx, tx, k, ci)
}

func resetOldCommitInfo(ctx context.Context, tx *pachsql.Tx, ci *pfs.CommitInfo) error {
	data, err := proto.Marshal(ci)
	if err != nil {
		return errors.Wrapf(err, "unmarshal old commit %q", commitKey(ci.Commit))
	}
	_, err = tx.ExecContext(ctx, `UPDATE collections.commits SET proto=$1  WHERE key=$2;`, data, commitKey(ci.Commit))
	return errors.Wrapf(err, "reset old commit info %q", commitKey(ci.Commit))
}

func getCommitIntId(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) (int, error) {
	r := tx.QueryRowContext(ctx, "SELECT int_id FROM pfs.commits WHERE commit_id=$1", commitKey(commit))
	var intId int
	if err := r.Scan(&intId); err != nil {
		return 0, errors.Wrapf(err, "scan int_id for commit_id %q", commitKey(commit))
	}
	return intId, nil
}

func forEachBranch(ctx context.Context, tx *pachsql.Tx, f func(key string, bi *pfs.BranchInfo)) error {
	rr, err := tx.QueryContext(ctx, `SELECT key, proto FROM collections.branches`)
	if err != nil {
		return errors.Wrap(err, "could not read table")
	}
	defer rr.Close()
	for rr.Next() {
		var (
			oldKey string
			pb     []byte
		)
		if err := rr.Err(); err != nil {
			return errors.Wrap(err, "row error")
		}
		if err := rr.Scan(&oldKey, &pb); err != nil {
			return errors.Wrap(err, "could not scan row")
		}
		bi := &pfs.BranchInfo{}
		if err := proto.Unmarshal(pb, bi); err != nil {
			return errors.Wrapf(err, "could not unmarshal proto")
		}
		f(oldKey, bi)
	}
	return nil
}

func getBranchInfo(ctx context.Context, tx *pachsql.Tx, key string) (*pfs.BranchInfo, error) {
	var pb []byte
	if err := tx.GetContext(ctx, &pb, "SELECT proto FROM collections.branches WHERE key=$1", key); err != nil {
		return nil, errors.Wrapf(err, "get branch with key %q", key)
	}
	bi := &pfs.BranchInfo{}
	if err := proto.Unmarshal(pb, bi); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal proto")
	}
	return bi, nil
}

func updateBranchInfo(ctx context.Context, tx *pachsql.Tx, key string, bi *pfs.BranchInfo) error {
	data, err := proto.Marshal(bi)
	if err != nil {
		return errors.Wrapf(err, "unmarshal branch: %q, proto: %v", key, bi)
	}
	_, err = tx.ExecContext(ctx, `UPDATE collections.branches SET proto=$1  WHERE key=$2;`, data, key)
	return errors.Wrapf(err, "update collections.branches for key %q", key)
}

func updateOldBranch(ctx context.Context, tx *pachsql.Tx, b *pfs.Branch, f func(bi *pfs.BranchInfo)) error {
	k := branchKey(b)
	bi, err := getBranchInfo(ctx, tx, k)
	if err != nil {
		return errors.Wrapf(err, "get branch info for branch: %q", b)
	}
	f(bi)
	return updateBranchInfo(ctx, tx, k, bi)
}

func deleteCommit(ctx context.Context, tx *pachsql.Tx, ci *pfs.CommitInfo, ancestor *pfs.Commit, restoreLinks bool) error {
	key := commitKey(ci.Commit)
	cId, err := getCommitIntId(ctx, tx, ci.Commit)
	if err != nil {
		return err
	}
	ancestorId, err := getCommitIntId(ctx, tx, ancestor)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_provenance SET from_id=$1 WHERE from_id=$2;`, ancestorId, cId); err != nil {
		return errors.Wrapf(err, "update commit provenance for commit %q", commitKey(ci.Commit))
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_provenance SET to_id=$1 WHERE to_id=$2;`, ancestorId, cId); err != nil {
		return errors.Wrapf(err, "update commit provenance for commit %q", commitKey(ci.Commit))
	}
	// delete commit
	if _, err := tx.ExecContext(ctx, `DELETE FROM pfs.commits WHERE int_id=$1;`, cId); err != nil {
		return errors.Wrapf(err, "delete from pfs.commits for commit %q", commitKey(ci.Commit))
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM collections.commits WHERE key=$1;`, key); err != nil {
		return errors.Wrapf(err, "delete from collections.commits for commit %q", commitKey(ci.Commit))
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
			return errors.Wrapf(err, "update parent commit %q", commitKey(ci.ParentCommit))
		}
		for _, c := range ci.ChildCommits {
			if err := updateOldCommit(ctx, tx, c, func(child *pfs.CommitInfo) {
				child.ParentCommit = ci.ParentCommit
			}); err != nil {
				return errors.Wrapf(err, "update child commit %q", pfsdb.CommitKey(c))
			}
		}
	}
	return nil
}

func addCommit(ctx context.Context, tx *pachsql.Tx, commit *pfs.Commit) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO pfs.commits(commit_id, commit_set_id, branch) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING;`, commitKey(commit), commit.ID, branchKey(commit.GetBranch()))
	return errors.Wrapf(err, "insert commit %q", commitKey(commit))
}

func getFileSetMd(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) (*fileset.Metadata, error) {
	var id string
	if err := tx.GetContext(ctx, &id, `SELECT fileset_id FROM pfs.commit_totals WHERE commit_id = $1`, commitKey(c)); err != nil {
		return nil, err
	}
	var mdData []byte
	if err := tx.GetContext(ctx, &mdData, `SELECT metadata_pb FROM storage.filesets WHERE id = $1`, id); err != nil {
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
	if md1.GetPrimitive() != nil && md2.GetPrimitive() != nil {
		if md1.GetPrimitive().SizeBytes != md2.GetPrimitive().SizeBytes {
			return false, nil
		}
		if len(md1.GetPrimitive().Additive.File.DataRefs) != len(md2.GetPrimitive().Additive.File.DataRefs) {
			return false, nil
		}
		for i, dr := range md1.GetPrimitive().Additive.File.DataRefs {
			if res := bytes.Compare(dr.Hash, md2.GetPrimitive().Additive.File.DataRefs[i].Hash); res != 0 {
				return false, nil
			}
		}
		for i, dr := range md1.GetPrimitive().Deletive.File.DataRefs {
			if res := bytes.Compare(dr.Hash, md2.GetPrimitive().Deletive.File.DataRefs[i].Hash); res != 0 {
				return false, nil
			}
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
	cis := make([]*pfs.CommitInfo, 0)
	ci := &pfs.CommitInfo{}
	// first fill commit provenance to all the existing commit sets
	if err := forEachCommit(ctx, tx, ci, func(string) error {
		cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
		return nil
	}); err != nil {
		return errors.Wrap(err, "populate the pfs.commits table")
	}
	for _, ci := range cis {
		if err := addCommit(ctx, tx, ci.Commit); err != nil {
			return err
		}
	}
	deleteCommits := make(map[string]*pfs.CommitInfo)
	for _, ci := range cis {
		for _, b := range ci.DirectProvenance {
			if err := addCommitProvenance(ctx, tx, ci.Commit, b.NewCommit(ci.Commit.ID)); err != nil {
				return errors.Wrapf(err, "add commit provenance from %q to %q", commitKey(ci.Commit), commitKey(b.NewCommit(ci.Commit.ID)))
			}
		}
		// collect all the ALIAS commits to be deleted
		if ci.Origin.Kind == pfs.OriginKind_ALIAS && ci.ParentCommit != nil {
			deleteCommits[commitKey(ci.Commit)] = proto.Clone(ci).(*pfs.CommitInfo)
		}
	}
	for k, ci := range deleteCommits {
		same, err := sameFileSets(ctx, tx, ci.Commit, ci.ParentCommit)
		if err != nil {
			return err
		}
		// this alias commit shouldn't be deleted, and instead should be converted to a AUTO commit
		if !same {
			delete(deleteCommits, k)
			ci.Origin.Kind = pfs.OriginKind_AUTO
			if err := updateCommitInfo(ctx, tx, commitKey(ci.Commit), ci); err != nil {
				return err
			}
		}
	}
	realAncestors := make(map[string]*pfs.CommitInfo)
	childToNewParents := make(map[*pfs.Commit]*pfs.Commit)
	for _, ci := range deleteCommits {
		anctsr := oldestAncestor(ci, deleteCommits)
		if _, ok := realAncestors[commitKey(anctsr)]; !ok {
			ci, err := getCommitInfo(ctx, tx, commitKey(anctsr))
			if err != nil {
				return errors.Wrapf(err, "get commit info with key %q", commitKey(anctsr))
			}
			realAncestors[commitKey(anctsr)] = ci
		}
		ancstrInfo := realAncestors[commitKey(anctsr)]
		childSet := make(map[string]*pfs.Commit)
		for _, ch := range ancstrInfo.ChildCommits {
			childSet[commitKey(ch)] = ch
		}
		delete(childSet, commitKey(ci.Commit))
		newChildren := traverseToEdges(ci, deleteCommits, func(c *pfs.CommitInfo) []*pfs.Commit {
			return c.ChildCommits
		})
		for _, nc := range newChildren {
			childSet[commitKey(nc)] = nc
			childToNewParents[nc] = anctsr
		}
		ancstrInfo.ChildCommits = nil
		for _, c := range childSet {
			ancstrInfo.ChildCommits = append(ancstrInfo.ChildCommits, c)
		}
	}
	for _, ci := range realAncestors {
		if err := resetOldCommitInfo(ctx, tx, ci); err != nil {
			return errors.Wrapf(err, "reset real ancestor %q", commitKey(ci.Commit))
		}
	}
	for child, parent := range childToNewParents {
		if err := updateOldCommit(ctx, tx, child, func(ci *pfs.CommitInfo) {
			ci.ParentCommit = parent
		}); err != nil {
			return errors.Wrapf(err, "update child commit %q", commitKey(child))
		}
	}
	// now write out realAncestors and childToNewParents
	for _, ci := range deleteCommits {
		ancstr := oldestAncestor(ci, deleteCommits)
		// restoreLinks=false because we're already resetting parent/children relationships above
		if err := deleteCommit(ctx, tx, ci, ancstr, false /* restoreLinks */); err != nil {
			return errors.Wrapf(err, "delete real commit %q with ancestor", commitKey(ci.Commit), commitKey(ancstr))
		}
	}
	// re-point branches if necessary
	headlessBranches := make([]*pfs.BranchInfo, 0)
	if err := forEachBranch(ctx, tx, func(key string, bi *pfs.BranchInfo) {
		if _, ok := deleteCommits[commitKey(bi.Head)]; ok {
			headlessBranches = append(headlessBranches, bi)
		}
	}); err != nil {
		return errors.Wrap(err, "record headless branches")
	}
	for _, bi := range headlessBranches {
		prevHead := deleteCommits[commitKey(bi.Head)]
		if err := updateOldBranch(ctx, tx, bi.Branch, func(bi *pfs.BranchInfo) {
			bi.Head = oldestAncestor(prevHead, deleteCommits)
		}); err != nil {
			return errors.Wrap(err, "update headless branches")
		}
	}
	return nil
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
// returns all of the commits that
func traverseToEdges(startCommit *pfs.CommitInfo, skipSet map[string]*pfs.CommitInfo, traverse func(*pfs.CommitInfo) []*pfs.Commit) []*pfs.Commit {
	cs := []*pfs.Commit{startCommit.Commit}
	result := make([]*pfs.Commit, 0)
	var c *pfs.Commit
	for len(cs) > 0 {
		c, cs = cs[0], cs[1:]
		if c == nil {
			continue
		}
		ci, ok := skipSet[commitKey(c)]
		if ok {
			cs = append(cs, traverse(ci)...)
		} else {
			result = append(result, proto.Clone(c).(*pfs.Commit))
		}
	}
	return result
}

func commitSubvenance(ctx context.Context, tx *pachsql.Tx, c *pfs.Commit) ([]*pfs.Commit, error) {
	key := commitKey(c)
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
		// will there always be an origin branch? In theory the relationship is many to many between branches and commits
		// if branch != "" {
		// 	c.Branch = ParseBranch(branch)
		// }
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
// we can simply delete the deferred branch's commit and repoint it's direct subvenant commits to point to the master commit as their
// direct provenance commits.
//
// 2) assert that we didn't miss any duplicate commits, and error the migration if we did.
// TODO(acohen4): Write test to catch this
//
// 3) commit keys can now be substituted from <project>/<repo>@<branch>=<id> -> <project>/<repo>@<id>
func migrateToBranchlessCommits(ctx context.Context, tx *pachsql.Tx) error {
	// 1) handle deferred branches
	ci := &pfs.CommitInfo{}
	cis := make([]*pfs.CommitInfo, 0)
	if err := forEachCommit(ctx, tx, ci, func(string) error {
		cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
		return nil
	}); err != nil {
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
					return errors.Wrapf(err, "delete commit provenance for %q", commitKey(ci.Commit))
				}
				if err := addCommitProvenance(ctx, tx, sc, ci.ParentCommit); err != nil {
					return errors.Wrapf(err, "add commit provenance for %q", commitKey(ci.ParentCommit))
				}
			}
			if err := deleteCommit(ctx, tx, ci, ci.ParentCommit, true /* restoreLinks */); err != nil {
				return errors.Wrapf(err, "delete commit %q", commitKey(ci.Commit))
			}
			// if ci was the head of a branch, point the head to ci.Parent
			if ci.Commit.Branch != nil {
				bk := branchKey(ci.Commit.Branch)
				bi, err := getBranchInfo(ctx, tx, bk)
				if err != nil {
					return errors.Wrapf(err, "get branch info for %q", bk)
				}
				if bi.Head.ID == ci.Commit.ID {
					bi.Head = ci.ParentCommit
					if err := updateBranchInfo(ctx, tx, bk, bi); err != nil {
						return errors.Wrapf(err, "update branch info for %q", bk)
					}
				}
			}
		}
	}
	// 2) migrate commit keys
	var oldCommit = new(pfs.CommitInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "commits", commitsIndexes, oldCommit, func(oldKey string) (newKey string, newVal proto.Message, err error) {
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
		return commitBranchlessKey(oldCommit.Commit), oldCommit, nil

	}, withKeyGen(func(key interface{}) (string, error) {
		if commit, ok := key.(*pfs.Commit); !ok {
			return "", errors.New("key must be a commit")
		} else {
			return commitBranchlessKey(commit), nil
		}
	})); err != nil {
		return errors.Wrap(err, "could not migrate commits")
	}
	var oldBranch = new(pfs.BranchInfo)
	if err := migratePostgreSQLCollection(ctx, tx, "branches", branchesIndexes, oldBranch, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldBranch.Head.Repo = oldBranch.Branch.Repo
		return branchKey(oldBranch.Branch), oldBranch, nil
	},
		withKeyGen(func(key interface{}) (string, error) {
			if branch, ok := key.(*pfs.Branch); !ok {
				return "", errors.New("key must be a branch")
			} else {
				return branchKey(branch), nil
			}
		}),
		withKeyCheck(func(key string) error {
			keyParts := strings.Split(key, "@")
			if len(keyParts) != 2 {
				return errors.Errorf("branch key %s isn't valid, use BranchKey to generate it", key)
			}
			if uuid.IsUUIDWithoutDashes(keyParts[1]) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return repoKeyCheck(keyParts[0])
		})); err != nil {
		return errors.Wrap(err, "could not migrate branches")
	}
	cis = make([]*pfs.CommitInfo, 0)
	if err := forEachCommit(ctx, tx, ci, func(string) error {
		cis = append(cis, proto.Clone(ci).(*pfs.CommitInfo))
		return nil
	}); err != nil {
		return err
	}
	for _, ci := range cis {
		oldKey := commitKey(ci.Commit)
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
