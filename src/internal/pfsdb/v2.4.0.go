package pfsdb

import (
	"context"
	"strings"

	"github.com/gogo/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// returns collections released in v2.4.0 - specifically the projects collection
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV2_4_0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection("projects", nil, nil, nil, nil),
	}
}

func MigrateProjectV2_4_0(p *pfs.Project) *pfs.Project {
	if p == nil || p.Name == "" {
		return &pfs.Project{Name: "default"}
	}
	return p
}

func MigrateRepoV2_4_0(r *pfs.Repo) *pfs.Repo {
	r.Project = MigrateProjectV2_4_0(r.Project)
	return r
}

func migrateRepoInvoV2_4_0(r *pfs.RepoInfo) *pfs.RepoInfo {
	r.Repo = MigrateRepoV2_4_0(r.Repo)
	for i, b := range r.Branches {
		r.Branches[i] = MigrateBranchV2_4_0(b)
	}
	return r
}

func MigrateBranchV2_4_0(b *pfs.Branch) *pfs.Branch {
	b.Repo = MigrateRepoV2_4_0(b.Repo)
	return b
}

func migrateBranchInfoV2_4_0(b *pfs.BranchInfo) *pfs.BranchInfo {
	b.Branch = MigrateBranchV2_4_0(b.Branch)
	if b.Head != nil {
		b.Head = MigrateCommitV2_4_0(b.Head)
	}
	for i, bb := range b.Provenance {
		b.Provenance[i] = MigrateBranchV2_4_0(bb)
	}
	for i, bb := range b.DirectProvenance {
		b.DirectProvenance[i] = MigrateBranchV2_4_0(bb)
	}
	for i, bb := range b.Subvenance {
		b.Subvenance[i] = MigrateBranchV2_4_0(bb)
	}
	return b
}

func MigrateCommitV2_4_0(c *pfs.Commit) *pfs.Commit {
	c.Branch = MigrateBranchV2_4_0(c.Branch)
	return c
}

func migrateCommitInfoV2_4_0(c *pfs.CommitInfo) *pfs.CommitInfo {
	c.Commit = MigrateCommitV2_4_0(c.Commit)
	if c.ParentCommit != nil {
		c.ParentCommit = MigrateCommitV2_4_0(c.ParentCommit)
	}
	for i, cc := range c.ChildCommits {
		c.ChildCommits[i] = MigrateCommitV2_4_0(cc)
	}
	for i, bb := range c.DirectProvenance {
		c.DirectProvenance[i] = MigrateBranchV2_4_0(bb)
	}
	return c
}

// MigrateV2_4_0 migrates PFS to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.
func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
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
	if err := col.MigratePostgreSQLCollection_v2_4_0(ctx, tx, "repos", reposIndexes, oldRepo, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldRepo = migrateRepoInvoV2_4_0(oldRepo)
		return RepoKey(oldRepo.Repo), oldRepo, nil

	},
		col.WithKeyCheck(repoKeyCheck),
		col.WithKeyGen(func(key interface{}) (string, error) {
			if repo, ok := key.(*pfs.Repo); !ok {
				return "", errors.New("key must be a repo")
			} else {
				return RepoKey(repo), nil
			}
		}),
	); err != nil {
		return errors.Wrap(err, "could not migrate repos")
	}
	var oldBranch = new(pfs.BranchInfo)
	if err := col.MigratePostgreSQLCollection_v2_4_0(ctx, tx, "branches", branchesIndexes, oldBranch, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldBranch = migrateBranchInfoV2_4_0(oldBranch)
		return BranchKey(oldBranch.Branch), oldBranch, nil

	},
		col.WithKeyGen(func(key interface{}) (string, error) {
			if branch, ok := key.(*pfs.Branch); !ok {
				return "", errors.New("key must be a branch")
			} else {
				return BranchKey(branch), nil
			}
		}),
		col.WithKeyCheck(func(key string) error {
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
	if err := col.MigratePostgreSQLCollection_v2_4_0(ctx, tx, "commits", commitsIndexes, oldCommit, func(oldKey string) (newKey string, newVal proto.Message, err error) {
		oldCommit = migrateCommitInfoV2_4_0(oldCommit)
		return CommitKey(oldCommit.Commit), oldCommit, nil

	}, col.WithKeyGen(func(key interface{}) (string, error) {
		if commit, ok := key.(*pfs.Commit); !ok {
			return "", errors.New("key must be a commit")
		} else {
			return CommitKey(commit), nil
		}
	})); err != nil {
		return errors.Wrap(err, "could not migrate commits")
	}

	return nil
}
