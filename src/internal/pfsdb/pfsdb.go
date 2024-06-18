// Package pfsdb contains the database schema that PFS uses.
package pfsdb

import (
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

func ProjectKey(project *pfs.Project) string {
	return project.Name
}

func RepoKey(repo *pfs.Repo) string {
	return repo.Project.Name + "/" + repo.Name + "." + repo.Type
}

func CommitKey(commit *pfs.Commit) string {
	return RepoKey(commit.Repo) + "@" + commit.Id
}

func BranchKey(branch *pfs.Branch) string {
	return RepoKey(branch.Repo) + "@" + branch.Name
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

func ParseCommit(key string) *pfs.Commit {
	split := strings.Split(key, "@")
	return &pfs.Commit{
		Repo: ParseRepo(split[0]),
		Id:   split[1],
	}
}

func ParseBranch(key string) *pfs.Branch {
	split := strings.Split(key, "@")
	return &pfs.Branch{
		Repo: ParseRepo(split[0]),
		Name: split[1],
	}
}
