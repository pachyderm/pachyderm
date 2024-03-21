// Package pfsdb_28x contains the database schema that PFS uses.
//
// This is copied from src/internal/pfsdb as of v2.8.6.
package pfsdb_28x

import (
	"strings"

	"google.golang.org/protobuf/proto"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

const (
	branchesCollectionName = "branches"
	commitsCollectionName  = "commits"
)

func ProjectKey(project *pfs.Project) string {
	return project.Name
}

func RepoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

func ParseRepo(key string) *pfs.Repo {
	slashSplit := strings.Split(key, "/")
	dotSplit := strings.Split(slashSplit[1], ".")
	return &pfs.Repo{
		Project: &pfs.Project{Name: slashSplit[0]},
		Name:    dotSplit[0],
		Type:    dotSplit[1],
	}
}

func repoKeyCheck(key string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 || len(parts[1]) == 0 {
		return errors.Errorf("repo must have a specified type")
	}
	return nil
}

var CommitsRepoIndex = &col.Index{
	Name: "repo",
	Extract: func(val proto.Message) string {
		return RepoKey(val.(*pfs.CommitInfo).Commit.Repo)
	},
}

var CommitsBranchlessIndex = &col.Index{
	Name: "branchless",
	Extract: func(val proto.Message) string {
		return CommitKey(val.(*pfs.CommitInfo).Commit)
	},
}

var CommitsCommitSetIndex = &col.Index{
	Name: "commitset",
	Extract: func(val proto.Message) string {
		return val.(*pfs.CommitInfo).Commit.Id
	},
}

var commitsIndexes = []*col.Index{CommitsRepoIndex, CommitsBranchlessIndex, CommitsCommitSetIndex}

func ParseCommit(key string) *pfs.Commit {
	split := strings.Split(key, "@")
	return &pfs.Commit{
		Repo: ParseRepo(split[0]),
		Id:   split[1],
	}
}

func CommitKey(commit *pfs.Commit) string {
	return RepoKey(commit.Repo) + "@" + commit.Id
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
		col.WithPutHook(func(tx *pachsql.Tx, commitInfo interface{}) error {
			ci := commitInfo.(*pfs.CommitInfo)
			if ci.Commit.Repo == nil {
				return errors.New("Commits must have the repo field populated")
			}
			return AddCommit(tx, ci.Commit)
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

func ParseBranch(key string) *pfs.Branch {
	split := strings.Split(key, "@")
	return &pfs.Branch{
		Repo: ParseRepo(split[0]),
		Name: split[1],
	}
}

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
