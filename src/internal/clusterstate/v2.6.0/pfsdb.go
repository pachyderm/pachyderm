package v2_6_0

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
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

var schemaUpdate = `
       ALTER TABLE pfs.commits ADD CONSTRAINT fk_col_commit
                  FOREIGN KEY(commit_id)
                  REFERENCES collections.commits(key)
                  ON DELETE CASCADE
`

func setupCommitProvenanceV01(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, schemaUpdate)
	return errors.EnsureStack(err)
}

func SetupCommitProvenanceV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}

// TODO(acohen4): verify how postgres behaves when does altering a table affect foreign key constraints?
var schema = `
	CREATE TABLE pfs.commits (
		int_id BIGSERIAL PRIMARY KEY,
		commit_id VARCHAR(4096) UNIQUE,
                commit_set_id VARCHAR(4096) NOT NULL,
                created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE pfs.commit_provenance (
		from_id BIGSERIAL NOT NULL,
		to_id BIGSERIAL NOT NULL,
		PRIMARY KEY (from_id, to_id),
                CONSTRAINT fk_from_commit
                  FOREIGN KEY(from_id)
	          REFERENCES pfs.commits(int_id)
	          ON DELETE CASCADE,
                CONSTRAINT fk_to_commit
                  FOREIGN KEY(to_id)
	          REFERENCES pfs.commits(int_id)
	          ON DELETE CASCADE
	);

	CREATE INDEX ON pfs.commit_provenance (
		from_id
	);

	CREATE INDEX ON pfs.commit_provenance (
		to_id
	);
`

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

type commit struct {
	info           *v2_5_0.CommitInfo
	id             int
	totalFilesetMd *fileset.Metadata
}

func removeAliasCommits(ctx context.Context, tx *pachsql.Tx) error {
	cis, err := listCollectionProtos(ctx, tx, "commits", &v2_5_0.CommitInfo{})
	if err != nil {
		return err
	}
	// Insert the commits into pfs.commits.
	if err := func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "insertCommits")
		defer end(log.Errorp(&retErr))
		batcher := newPostgresBatcher(ctx, tx)
		for _, ci := range cis {
			// Skip alias commit.
			if ci.Origin.Kind == 4 {
				continue
			}
			stmt := fmt.Sprintf(`INSERT INTO pfs.commits (commit_id, commit_set_id) VALUES ('%v', '%v')`, oldCommitKey(ci.Commit), ci.Commit.Id)
			if err := batcher.Add(stmt); err != nil {
				return err
			}
		}
		return batcher.Close()
	}(); err != nil {
		return err
	}
	commits := make(map[string]*commit)
	for _, ci := range cis {
		commits[oldCommitKey(ci.Commit)] = &commit{info: ci}
	}
	if err := getCommitIds(ctx, tx, commits); err != nil {
		return err
	}
	if err := getTotalFilesetMds(ctx, tx, commits); err != nil {
		return err
	}
	getCommit := func(c *pfs.Commit) *commit {
		return commits[oldCommitKey(c)]
	}
	realAncestorCommits := make(map[string]*pfs.Commit)
	// Update the ancestry and provenance of the commits.
	if err := func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "updateCommits")
		defer end(log.Errorp(&retErr))
		batcher := newPostgresBatcher(ctx, tx)
		for _, ci := range cis {
			// Skip alias commit.
			if ci.Origin.Kind == 4 {
				continue
			}
			ci := proto.Clone(ci).(*v2_5_0.CommitInfo)
			// Update the parent and child commits.
			if ci.ParentCommit != nil {
				ci.ParentCommit = getRealAncestorCommit(getCommit, realAncestorCommits, ci.ParentCommit)
			}
			var childCommits []*pfs.Commit
			for _, child := range ci.ChildCommits {
				childCommits = append(childCommits, getRealDescendantCommits(getCommit, child)...)
			}
			ci.ChildCommits = childCommits
			// Update the direct provenance.
			var directProvenance []*pfs.Commit
			for _, b := range ci.DirectProvenance {
				toC := b.NewCommit(ci.Commit.Id)
				// If the provenant commit's repo was deleted, its
				// reference may still exist in the old provenance, so
				// we skip mapping it over to the new model.
				if _, ok := commits[oldCommitKey(toC)]; !ok {
					continue
				}
				toC = getRealAncestorCommit(getCommit, realAncestorCommits, toC)
				directProvenance = append(directProvenance, toC)
				from := strconv.Itoa(getCommit(ci.Commit).id)
				to := strconv.Itoa(getCommit(toC).id)
				stmt := fmt.Sprintf(`INSERT INTO pfs.commit_provenance(from_id, to_id) VALUES (%v, %v)`, from, to)
				if err := batcher.Add(stmt); err != nil {
					return err
				}
			}
			newCI := convertCommitInfoToV2_6_0(ci)
			newCI.DirectProvenance = directProvenance
			data, err := proto.Marshal(newCI)
			if err != nil {
				return errors.EnsureStack(err)
			}
			stmt := fmt.Sprintf("UPDATE collections.commits SET proto=decode('%v', 'hex') WHERE key='%v'", hex.EncodeToString(data), oldCommitKey(newCI.Commit))
			if err := batcher.Add(stmt); err != nil {
				return err
			}
		}
		return batcher.Close()
	}(); err != nil {
		return err
	}
	// Delete the alias commits.
	if err := func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "deleteAliasCommits")
		defer end(log.Errorp(&retErr))
		batcher := newPostgresBatcher(ctx, tx)
		for _, ci := range cis {
			// Skip non-alias commit.
			if ci.Origin.Kind != 4 {
				continue
			}
			if err := checkAliasCommit(getCommit, ci.Commit, ci.ParentCommit); err != nil {
				return err
			}
			key := oldCommitKey(ci.Commit)
			stmt := fmt.Sprintf(`DELETE FROM collections.commits WHERE key='%v'`, key)
			if err := batcher.Add(stmt); err != nil {
				return err
			}
			stmt = fmt.Sprintf(`DELETE FROM pfs.commit_diffs WHERE commit_id='%v'`, key)
			if err := batcher.Add(stmt); err != nil {
				return err
			}
			stmt = fmt.Sprintf(`DELETE FROM pfs.commit_totals WHERE commit_id='%v'`, key)
			if err := batcher.Add(stmt); err != nil {
				return err
			}
			// TODO: Why does LIKE perform so poorly?
			stmt = fmt.Sprintf(`DELETE FROM storage.tracker_objects WHERE str_id > 'commit/%v' AND str_id < 'commit/%v/z'`, key, key)
			if err := batcher.Add(stmt); err != nil {
				return err
			}
		}
		return batcher.Close()
	}(); err != nil {
		return err
	}
	bis, err := listCollectionProtos(ctx, tx, "branches", &pfs.BranchInfo{})
	if err != nil {
		return err
	}
	for _, bi := range bis {
		if err := updateBranch(ctx, tx, bi.Branch, func(bi *pfs.BranchInfo) {
			bi.Head = getRealAncestorCommit(getCommit, realAncestorCommits, bi.Head)
		}); err != nil {
			return errors.Wrap(err, "update headless branches")
		}
	}
	return nil
}

func getCommitIds(ctx context.Context, tx *pachsql.Tx, commits map[string]*commit) (retErr error) {
	rs, err := tx.QueryContext(ctx, "SELECT int_id, commit_id FROM pfs.commits")
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rs, "close rows")
	for rs.Next() {
		var id int
		var strId string
		if err := rs.Scan(&id, &strId); err != nil {
			return errors.EnsureStack(err)
		}
		commits[strId].id = id
	}
	return errors.EnsureStack(rs.Err())
}

func getTotalFilesetMds(ctx context.Context, tx *pachsql.Tx, commits map[string]*commit) (retErr error) {
	rs, err := tx.QueryContext(ctx, "SELECT commit_id, metadata_pb FROM pfs.commit_totals JOIN storage.filesets ON fileset_id = id")
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer errors.Close(&retErr, rs, "close rows")
	for rs.Next() {
		var id string
		var data []byte
		if err := rs.Scan(&id, &data); err != nil {
			return errors.EnsureStack(err)
		}
		md := &fileset.Metadata{}
		if err := proto.Unmarshal(data, md); err != nil {
			return errors.EnsureStack(err)
		}
		commits[id].totalFilesetMd = md
	}
	return errors.EnsureStack(rs.Err())
}

func getRealAncestorCommit(getCommit func(*pfs.Commit) *commit, realAncestorCommits map[string]*pfs.Commit, c *pfs.Commit) *pfs.Commit {
	realAncestorCommit, ok := realAncestorCommits[oldCommitKey(c)]
	if ok {
		return realAncestorCommit
	}
	ci := getCommit(c).info
	if ci.Origin.Kind != 4 {
		return c
	}
	realAncestorCommit = getRealAncestorCommit(getCommit, realAncestorCommits, ci.ParentCommit)
	realAncestorCommits[oldCommitKey(c)] = realAncestorCommit
	return realAncestorCommit
}

func getRealDescendantCommits(getCommit func(*pfs.Commit) *commit, c *pfs.Commit) []*pfs.Commit {
	ci := getCommit(c).info
	if ci.Origin.Kind != 4 {
		return []*pfs.Commit{c}
	}
	var childCommits []*pfs.Commit
	for _, child := range ci.ChildCommits {
		childCommits = append(childCommits, getRealDescendantCommits(getCommit, child)...)
	}
	return childCommits
}

func checkAliasCommit(getCommit func(*pfs.Commit) *commit, commit *pfs.Commit, parentCommit *pfs.Commit) error {
	md1 := getCommit(commit).totalFilesetMd
	md2 := getCommit(parentCommit).totalFilesetMd
	same, err := func() (bool, error) {
		if md1 == nil || md2 == nil {
			// the semantics here are a little odd - if either commit doesn't have a fileset yet,
			// we assume that they aren't different and are therefore the same.
			return true, nil
		}
		if md1.GetPrimitive() != nil && md2.GetPrimitive() != nil {
			ser1, err := proto.Marshal(md1.GetPrimitive())
			if err != nil {
				return false, errors.EnsureStack(err)
			}
			ser2, err := proto.Marshal(md2.GetPrimitive())
			if err != nil {
				return false, errors.EnsureStack(err)
			}
			return bytes.Equal(ser1, ser2), nil
		}
		if md1.GetComposite() != nil && md2.GetComposite() != nil {
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
		}
		return false, nil
	}()
	if err != nil {
		return err
	}
	if !same {
		return errors.Errorf("commit %q is listed as ALIAS but has a different ID than its parent commit", oldCommitKey(commit))
	}
	return nil
}

func updateBranch(ctx context.Context, tx *pachsql.Tx, b *pfs.Branch, f func(bi *pfs.BranchInfo)) error {
	k := branchKey(b)
	bi := &pfs.BranchInfo{}
	if err := getCollectionProto(ctx, tx, "branches", k, bi); err != nil {
		return errors.Wrapf(err, "get branch info for branch: %q", b)
	}
	f(bi)
	return updateCollectionProto(ctx, tx, "branches", k, k, bi)
}

func deleteDanglingCommitRefs(ctx context.Context, tx *pachsql.Tx) (retErr error) {
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
	listSourceCommits := func(tx *pachsql.Tx) (map[string]struct{}, error) {
		var keys []string
		if err := tx.Select(&keys, `SELECT key FROM collections.commits`); err != nil {
			return nil, errors.Wrap(err, "select keys from collections.commits")
		}
		cis := make(map[string]struct{})
		for _, k := range keys {
			cis[k] = struct{}{}
		}
		return cis, nil
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
		if err := tx.Select(&ids, `SELECT commit_id FROM pfs.commit_totals`); err != nil {
			return nil, errors.Wrap(err, "select commit ids from pfs.commit_totals")
		}
		for _, id := range ids {
			cs[id], err = parseCommit_2_5(id)
			if err != nil {
				return nil, err
			}
		}
		ids = make([]string, 0)
		if err := tx.Select(&ids, `SELECT commit_id FROM pfs.commit_diffs`); err != nil {
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
	cis, err := listSourceCommits(tx)
	if err != nil {
		return errors.Wrap(err, "list repos")
	}
	var dangCommitKeys []string
	for _, c := range cs {
		if _, ok := cis[oldCommitKey(c)]; !ok {
			dangCommitKeys = append(dangCommitKeys, oldCommitKey(c))
		}
	}
	if len(dangCommitKeys) > 0 {
		log.Info(ctx, "detected dangling commit references", zap.Any("references", dangCommitKeys))
	}
	ctx, end := log.SpanContext(ctx, "deleteDanglingCommits")
	defer end(log.Errorp(&retErr))
	batcher := newPostgresBatcher(ctx, tx)
	for _, id := range dangCommitKeys {
		stmt := fmt.Sprintf(`DELETE FROM pfs.commit_totals WHERE commit_id = '%v'`, id)
		if err := batcher.Add(stmt); err != nil {
			return err
		}
		stmt = fmt.Sprintf(`DELETE FROM pfs.commit_diffs WHERE commit_id = '%v'`, id)
		if err := batcher.Add(stmt); err != nil {
			return err
		}
	}
	return batcher.Close()
}

func branchlessCommitsPFS(ctx context.Context, tx *pachsql.Tx) error {
	// Update commits table.
	cis, err := listCollectionProtos(ctx, tx, "commits", &pfs.CommitInfo{})
	if err != nil {
		return err
	}
	if err := func() (retErr error) {
		ctx, end := log.SpanContext(ctx, "branchlessUpdateCommits")
		defer end(log.Errorp(&retErr))
		batcher := newPostgresBatcher(ctx, tx)
		for _, ci := range cis {
			stmt := fmt.Sprintf(`UPDATE pfs.commits SET commit_id='%v' WHERE commit_id='%v'`, commitBranchlessKey(ci.Commit), oldCommitKey(ci.Commit))
			if err := batcher.Add(stmt); err != nil {
				return err
			}
			ci.Commit.Repo = ci.Commit.Branch.Repo
			if ci.ParentCommit != nil {
				ci.ParentCommit.Repo = ci.ParentCommit.Branch.Repo
			}
			for _, child := range ci.ChildCommits {
				child.Repo = child.Branch.Repo
			}
			for _, prov := range ci.DirectProvenance {
				prov.Repo = prov.Branch.Repo
			}
			data, err := proto.Marshal(ci)
			if err != nil {
				return errors.EnsureStack(err)
			}
			stmt = fmt.Sprintf(`UPDATE collections.commits SET key='%v', proto=decode('%v', 'hex') WHERE key='%v'`, commitBranchlessKey(ci.Commit), hex.EncodeToString(data), oldCommitKey(ci.Commit))
			if err := batcher.Add(stmt); err != nil {
				return err
			}
		}
		return batcher.Close()
	}(); err != nil {
		return err
	}
	// Update branches table.
	bis, err := listCollectionProtos(ctx, tx, "branches", &pfs.BranchInfo{})
	if err != nil {
		return err
	}
	for _, bi := range bis {
		if err := updateBranch(ctx, tx, bi.Branch, func(bi *pfs.BranchInfo) {
			bi.Head.Repo = bi.Head.Branch.Repo
		}); err != nil {
			return errors.Wrap(err, "update headless branches")
		}
	}
	// Update everything else.
	return branchlessCommitKeysPFS(ctx, tx)
}

// map <project>/<repo>@<branch>=<id> -> <project>/<repo>@<id>
func branchlessCommitKeysPFS(ctx context.Context, tx *pachsql.Tx) (retErr error) {
	ctx, end := log.SpanContext(ctx, "branchlessCommitKeysPFS")
	defer end(log.Errorp(&retErr))
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
