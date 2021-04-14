// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// Repos returns a collection of repos
func Repos(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&pfs.RepoInfo{},
		nil,
		nil,
	)
}

var CommitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.Repo.Name
	},
}

func CommitKey(commit *pfs.Commit) string {
	return commit.Repo.Name + "@" + commit.ID
}

// Commits returns a collection of commits
func Commits(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&pfs.CommitInfo{},
		[]*col.Index{CommitsRepoIndex},
		nil,
	)
}

var BranchesRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return val.(*pfs.BranchInfo).Branch.Repo.Name
	},
}

func BranchKey(branch *pfs.Branch) string {
	return branch.Repo.Name + "@" + branch.Name
}

// Branches returns a collection of branches
func Branches(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&pfs.BranchInfo{},
		[]*col.Index{BranchesRepoIndex},
		func(key string) error {
			if uuid.IsUUIDWithoutDashes(key) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return nil
		},
	)
}

// OpenCommits returns a collection of open commits
func OpenCommits(ctx context.Context, db *sqlx.DB, listener *col.PostgresListener) (col.PostgresCollection, error) {
	return col.NewPostgresCollection(
		ctx,
		db,
		listener,
		&pfs.Commit{},
		nil,
		nil,
	)
}
