package pfs

import (
	"sort"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"go.pedge.io/proto/time"
)

// ReduceRepoInfos combines repo info for each named repo,
// adding byte-sizes together as appropriate.
func ReduceRepoInfos(repoInfos []*pfs.RepoInfo) []*pfs.RepoInfo {
	// Create a map from repo name to info.
	reducedRepoInfos := make(map[string]*pfs.RepoInfo)
	for _, repoInfo := range repoInfos {
		reducedRepoInfo, ok := reducedRepoInfos[repoInfo.Repo.Name]
		if !ok {
			// Repo name not yet seen, just add the repoInfo directly
			reducedRepoInfos[repoInfo.Repo.Name] = repoInfo
			continue
		}
		// Repo name already seen, add instead of overwriting
		reducedRepoInfo.SizeBytes += repoInfo.SizeBytes
		reducedRepoInfo.Provenance = repoInfo.Provenance
	}
	// Convert the map back to a slice and sort it before returning
	var result []*pfs.RepoInfo
	for _, repoInfo := range reducedRepoInfos {
		result = append(result, repoInfo)
	}
	sort.Sort(sortRepoInfos(result))
	return result
}

// ReduceCommitInfos combines commit info for each commit id,
// resolving writes and adding byte-sizes together as appropriate.
func ReduceCommitInfos(commitInfos []*pfs.CommitInfo) []*pfs.CommitInfo {
	// Create a map from commit id to info.
	reducedCommitInfos := make(map[string]*pfs.CommitInfo)
	for _, commitInfo := range commitInfos {
		reducedCommitInfo, ok := reducedCommitInfos[commitInfo.Commit.ID]
		if !ok {
			// Commit id not yet seen, just add the commitInfo directly
			reducedCommitInfos[commitInfo.Commit.ID] = commitInfo
			continue
		}
		// Commit id already seen, check for write and add instead of overwriting
		if commitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_WRITE {
			// (WRITE && READ) => WRITE
			reducedCommitInfo.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
		}
		reducedCommitInfo.SizeBytes += commitInfo.SizeBytes
		reducedCommitInfo.Provenance = commitInfo.Provenance
	}
	// Convert the map back to a slice and sort it before returning
	var result []*pfs.CommitInfo
	for _, commitInfo := range reducedCommitInfos {
		result = append(result, commitInfo)
	}
	sort.Sort(sortCommitInfos(result))
	return result
}

// ReduceFileInfos combines file info for each file path, taking the
// latest modification time for each path and combining their children.
func ReduceFileInfos(fileInfos []*pfs.FileInfo) []*pfs.FileInfo {
	// Create a map from file path to info
	reducedFileInfos := make(map[string]*pfs.FileInfo)
	for _, fileInfo := range fileInfos {
		reducedFileInfo, ok := reducedFileInfos[fileInfo.File.Path]
		if !ok {
			// File path not yet seen, just add the fileInfo directly
			reducedFileInfos[fileInfo.File.Path] = fileInfo
			continue
		}
		// File path already seen, compare modification dates and update children
		if prototime.TimestampToTime(fileInfo.Modified).
			After(prototime.TimestampToTime(reducedFileInfo.Modified)) {
			reducedFileInfo.Modified = fileInfo.Modified
			reducedFileInfo.CommitModified = fileInfo.CommitModified
		}
		reducedFileInfo.Children = append(reducedFileInfo.Children, fileInfo.Children...)
	}
	// Convert the map back to a slice and return it
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
