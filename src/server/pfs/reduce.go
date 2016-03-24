package pfs

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"go.pedge.io/proto/time"
)

func ReduceRepoInfos(repoInfos []*pfs.RepoInfo) []*pfs.RepoInfo {
	reducedRepoInfos := make(map[string]*pfs.RepoInfo)
	for _, repoInfo := range repoInfos {
		reducedRepoInfo, ok := reducedRepoInfos[repoInfo.Repo.Name]
		if !ok {
			reducedRepoInfos[repoInfo.Repo.Name] = repoInfo
			continue
		}
		reducedRepoInfo.SizeBytes += repoInfo.SizeBytes
	}
	var result []*pfs.RepoInfo
	for _, repoInfo := range reducedRepoInfos {
		result = append(result, repoInfo)
	}
	sort.Sort(sortRepoInfos(result))
	return result
}

func ReduceCommitInfos(commitInfos []*pfs.CommitInfo) []*pfs.CommitInfo {
	reducedCommitInfos := make(map[string]*pfs.CommitInfo)
	for _, commitInfo := range commitInfos {
		reducedCommitInfo, ok := reducedCommitInfos[commitInfo.Commit.ID]
		if !ok {
			reducedCommitInfos[commitInfo.Commit.ID] = commitInfo
			continue
		}
		if commitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_WRITE {
			reducedCommitInfo.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
		}
		reducedCommitInfo.SizeBytes += commitInfo.SizeBytes
	}
	var result []*pfs.CommitInfo
	for _, commitInfo := range reducedCommitInfos {
		result = append(result, commitInfo)
	}
	sort.Sort(sortCommitInfos(result))
	return result
}

func ReduceFileInfos(fileInfos []*pfs.FileInfo) []*pfs.FileInfo {
	reducedFileInfos := make(map[string]*pfs.FileInfo)
	for _, fileInfo := range fileInfos {
		reducedFileInfo, ok := reducedFileInfos[fileInfo.File.Path]
		if !ok {
			reducedFileInfos[fileInfo.File.Path] = fileInfo
			continue
		}
		if prototime.TimestampToTime(fileInfo.Modified).
			After(prototime.TimestampToTime(reducedFileInfo.Modified)) {
			reducedFileInfo.Modified = fileInfo.Modified
			reducedFileInfo.CommitModified = fileInfo.CommitModified
		}
		reducedFileInfo.Children = append(reducedFileInfo.Children, fileInfo.Children...)
	}
	var result []*pfs.FileInfo
	for _, reducedFileInfo := range reducedFileInfos {
		result = append(result, reducedFileInfo)
	}
	return result
}

type sortRepoInfos []*pfs.RepoInfo

func (a sortRepoInfos) Len() int {
	return len(a)
}

func (a sortRepoInfos) Less(i, j int) bool {
	return prototime.TimestampToTime(a[i].Created).After(prototime.TimestampToTime(a[j].Created))
}
func (a sortRepoInfos) Swap(i, j int) {
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}

type sortCommitInfos []*pfs.CommitInfo

func (a sortCommitInfos) Len() int {
	return len(a)
}

func (a sortCommitInfos) Less(i, j int) bool {
	if a[i].Finished != nil && a[j].Finished != nil {
		return prototime.TimestampToTime(a[i].Finished).After(prototime.TimestampToTime(a[j].Finished))
	} else if a[i].Finished != nil {
		return true
	} else if a[j].Finished != nil {
		return false
	}
	return prototime.TimestampToTime(a[i].Started).After(prototime.TimestampToTime(a[j].Started))
}
func (a sortCommitInfos) Swap(i, j int) {
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}
