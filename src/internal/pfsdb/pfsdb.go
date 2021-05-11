// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

const (
	reposCollectionName       = "repos"
	branchesCollectionName    = "branches"
	commitsCollectionName     = "commits"
	openCommitsCollectionName = "open_commits"
)

// AllCollections returns a list of all the PFS collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(reposCollectionName, nil, nil, nil, []*col.Index{ReposNameIndex, ReposTypeIndex}, nil),
		col.NewPostgresCollection(commitsCollectionName, nil, nil, nil, []*col.Index{CommitsRepoIndex}, nil),
		col.NewPostgresCollection(branchesCollectionName, nil, nil, nil, []*col.Index{BranchesRepoIndex}, nil),
		col.NewPostgresCollection(openCommitsCollectionName, nil, nil, nil, nil, nil),
	}
}

var ReposTypeIndex = &col.Index{
	Name: "type",
	Extract: func(val proto.Message) string {
		return val.(*pfs.RepoInfo).Repo.Type
	},
}

var ReposNameIndex = &col.Index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return val.(*pfs.RepoInfo).Repo.Name
	},
}

func RepoKey(repo *pfs.Repo) string {
	return repo.Name + "." + repo.Type
}

// Repos returns a collection of repos
func Repos(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"repos",
		db,
		listener,
		&pfs.RepoInfo{},
		[]*col.Index{ReposNameIndex, ReposTypeIndex},
		func(key string) error {
			parts := strings.Split(key, ".")
			if len(parts) < 2 || len(parts[1]) == 0 {
				return errors.Errorf("repo must have a specified type")
			}
			return nil
		},
	)
}

var CommitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.CommitInfo).Commit.Repo)
	},
}

func CommitKey(commit *pfs.Commit) string {
	return RepoKey(commit.Repo) + "@" + commit.ID
}

// Commits returns a collection of commits
func Commits(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"commits",
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
		return RepoKey(val.(*pfs.BranchInfo).Branch.Repo)
	},
}

func BranchKey(branch *pfs.Branch) string {
	return RepoKey(branch.Repo) + "@" + branch.Name
}

// Branches returns a collection of branches
func Branches(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"branches",
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
func OpenCommits(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"open_commits",
		db,
		listener,
		&pfs.Commit{},
		nil,
		nil,
	)
}
