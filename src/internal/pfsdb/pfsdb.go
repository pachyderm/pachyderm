// Package pfsdb contains the database schema that PFS uses.
package pfsdb

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
	reposCollectionName    = "repos"
	branchesCollectionName = "branches"
	commitsCollectionName  = "commits"
	projectsCollectionName = "projects"
)

var ReposTypeIndex = &col.Index{
	Name: "type",
	Extract: func(val proto.Message) string {
		return val.(*pfs.RepoInfo).Repo.Type
	},
}

func ReposNameKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name
}

var ReposNameIndex = &col.Index{
	Name: "name",
	Extract: func(val proto.Message) string {
		return ReposNameKey(val.(*pfs.RepoInfo).Repo)
	},
}

var reposIndexes = []*col.Index{ReposNameIndex, ReposTypeIndex}

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

func ProjectKey(project *pfs.Project) string {
	return project.Name
}

func Projects(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		projectsCollectionName,
		db,
		listener,
		&pfs.ProjectInfo{},
		nil,
		col.WithKeyGen(func(key interface{}) (string, error) {
			if project, ok := key.(*pfs.Project); !ok {
				return "", errors.New("key must be a project")
			} else {
				return ProjectKey(project), nil
			}
		}),
		col.WithNotFoundMessage(func(key interface{}) string {
			return pfsserver.ErrProjectNotFound{Project: key.(*pfs.Project)}.Error()
		}),
		col.WithExistsMessage(func(key interface{}) string {
			return pfsserver.ErrProjectExists{Project: key.(*pfs.Project)}.Error()
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
