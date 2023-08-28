package v2_7_0

import "github.com/pachyderm/pachyderm/v2/src/pfs"

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
