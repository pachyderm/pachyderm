package v2_5_0

import (
	"context"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
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
		return repoKey(val.(*CommitInfo).Commit.Branch.Repo)
	},
}

var commitsBranchlessIndex = &index{
	Name: "branchless",
	Extract: func(val proto.Message) string {
		return commitBranchlessKey(val.(*CommitInfo).Commit)
	},
}

var commitsCommitSetIndex = &index{
	Name: "commitset",
	Extract: func(val proto.Message) string {
		return val.(*CommitInfo).Commit.Id
	},
}

var commitsIndexes = []*index{commitsRepoIndex, commitsBranchlessIndex, commitsCommitSetIndex}

func commitKey(commit *pfs.Commit) string {
	return branchKey(commit.Branch) + "=" + commit.Id
}

func commitBranchlessKey(commit *pfs.Commit) string {
	return repoKey(commit.Branch.Repo) + "@" + commit.Id
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

func migrateCommitInfo(c *CommitInfo) *CommitInfo {
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
		return errors.Wrap(err, "could not update pfs.commit_diffs")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pfs.commit_totals SET commit_id = regexp_replace(commit_id, '^([-a-zA-Z0-9_]+)', 'default/\1') WHERE commit_id ~ '^[-a-zA-Z0-9_]+\.';`); err != nil {
		return errors.Wrap(err, "could not update pfs.commit_totals")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE storage.tracker_objects SET str_id = regexp_replace(str_id, 'commit/([-a-zA-Z0-9_]+)', 'commit/default/\1') WHERE str_id ~ '^commit/[-a-zA-Z0-9_]+\.';`); err != nil {
		return errors.Wrapf(err, "could not update storage.tracker_objects")
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
	var oldCommit = new(CommitInfo)
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
