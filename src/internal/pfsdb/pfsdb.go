// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"strings"

	"github.com/gogo/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
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
func Repos(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		reposCollectionName,
		db,
		listener,
		&pfs.RepoInfo{},
		reposIndexes,
		col.WithKeyCheck(repoKeyCheck),
		col.WithKeyGen(func(key interface{}) (string, error) {
			if repo, ok := key.(*pfs.Repo); !ok {
				return "", errors.New("key must be a repo")
			} else {
				return RepoKey(repo), nil
			}
		}),
		col.WithNotFoundMessage(func(key interface{}) string {
			return pfsserver.ErrRepoNotFound{Repo: key.(*pfs.Repo)}.Error()
		}),
		col.WithExistsMessage(func(key interface{}) string {
			return pfsserver.ErrRepoExists{Repo: key.(*pfs.Repo)}.Error()
		}),
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
func Commits(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		commitsCollectionName,
		db,
		listener,
		&pfs.CommitInfo{},
		commitsIndexes,
		col.WithKeyGen(func(key interface{}) (string, error) {
			if commit, ok := key.(*pfs.Commit); !ok {
				return "", errors.New("key must be a commit")
			} else {
				return CommitKey(commit), nil
			}
		}),
		col.WithNotFoundMessage(func(key interface{}) string {
			return pfsserver.ErrCommitNotFound{Commit: key.(*pfs.Commit)}.Error()
		}),
		col.WithExistsMessage(func(key interface{}) string {
			return pfsserver.ErrCommitExists{Commit: key.(*pfs.Commit)}.Error()
		}),
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
func Branches(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		branchesCollectionName,
		db,
		listener,
		&pfs.BranchInfo{},
		branchesIndexes,
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
		}),
		col.WithNotFoundMessage(func(key interface{}) string {
			return pfsserver.ErrBranchNotFound{Branch: key.(*pfs.Branch)}.Error()
		}),
		col.WithExistsMessage(func(key interface{}) string {
			return pfsserver.ErrBranchExists{Branch: key.(*pfs.Branch)}.Error()
		}),
	)
}

// AllCollections returns a list of all the PFS collections for
// postgres-initialization purposes. These collections are not usable for
// querying.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(reposCollectionName, nil, nil, nil, reposIndexes),
		col.NewPostgresCollection(commitsCollectionName, nil, nil, nil, commitsIndexes),
		col.NewPostgresCollection(branchesCollectionName, nil, nil, nil, branchesIndexes),
	}
}
