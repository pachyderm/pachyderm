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
	reposCollectionName    = "repos"
	branchesCollectionName = "branches"
	commitsCollectionName  = "commits"
)

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

var reposIndexes = []*col.Index{ReposNameIndex, ReposTypeIndex}

func RepoKey(repo *pfs.Repo) string {
	return repo.Name + "." + repo.Type
}

func repoKeyCheck(key string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 || len(parts[1]) == 0 {
		return errors.Errorf("repo must have a specified type")
	}
	return nil
}

// Repos returns a collection of repos
func Repos(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		reposCollectionName,
		db,
		listener,
		&pfs.RepoInfo{},
		reposIndexes,
		repoKeyCheck,
	)
}

var CommitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.CommitInfo).Commit.Branch.Repo)
	},
}

var CommitsBranchlessIndex = &col.Index{
	Name: "branchless",
	Extract: func(val proto.Message) string {
		return CommitBranchlessKey(val.(*pfs.CommitInfo).Commit)
	},
}

var CommitsCommitSetIndex = &col.Index{
	Name: "commitset",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.ID
	},
}

var commitsIndexes = []*col.Index{CommitsRepoIndex, CommitsBranchlessIndex, CommitsCommitSetIndex}

func CommitKey(commit *pfs.Commit) string {
	return BranchKey(commit.Branch) + "=" + commit.ID
}

func CommitBranchlessKey(commit *pfs.Commit) string {
	return RepoKey(commit.Branch.Repo) + "@" + commit.ID
}

// Commits returns a collection of commits
func Commits(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		commitsCollectionName,
		db,
		listener,
		&pfs.CommitInfo{},
		commitsIndexes,
		nil,
	)
}

var BranchesRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.BranchInfo).Branch.Repo)
	},
}

var branchesIndexes = []*col.Index{BranchesRepoIndex}

func BranchKey(branch *pfs.Branch) string {
	return RepoKey(branch.Repo) + "@" + branch.Name
}

// Branches returns a collection of branches
func Branches(db *sqlx.DB, listener *col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		branchesCollectionName,
		db,
		listener,
		&pfs.BranchInfo{},
		branchesIndexes,
		func(key string) error {
			keyParts := strings.Split(key, "@")
			if len(keyParts) != 2 {
				return errors.Errorf("branch key %s isn't valid, use BranchKey to generate it", key)
			}
			if uuid.IsUUIDWithoutDashes(keyParts[1]) {
				return errors.Errorf("branch name cannot be a UUID V4")
			}
			return repoKeyCheck(keyParts[0])
		},
	)
}

// AllCollections returns a list of all the PFS collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
func AllCollections() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(reposCollectionName, nil, nil, nil, reposIndexes, nil),
		col.NewPostgresCollection(commitsCollectionName, nil, nil, nil, commitsIndexes, nil),
		col.NewPostgresCollection(branchesCollectionName, nil, nil, nil, branchesIndexes, nil),
	}
}
