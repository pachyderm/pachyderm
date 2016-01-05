package pfs

import (
	"sort"

	"go.pedge.io/proto/time"
)

func ReduceFileInfos(fileInfos []*FileInfo) []*FileInfo {
	reducedFileInfos := make(map[string]*FileInfo)
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
	var result []*FileInfo
	for _, reducedFileInfo := range reducedFileInfos {
		result = append(result, reducedFileInfo)
	}
	return result
}

func ReduceCommitInfos(commitInfos []*CommitInfo) []*CommitInfo {
	reducedCommitInfos := make(map[string]*CommitInfo)
	for _, commitInfo := range commitInfos {
		reducedCommitInfo, ok := reducedCommitInfos[commitInfo.Commit.Id]
		if !ok {
			reducedCommitInfos[commitInfo.Commit.Id] = commitInfo
			continue
		}
		if commitInfo.CommitType == CommitType_COMMIT_TYPE_WRITE {
			reducedCommitInfo.CommitType = CommitType_COMMIT_TYPE_WRITE
		}
		reducedCommitInfo.SizeBytes += commitInfo.SizeBytes
	}
	var result []*CommitInfo
	for _, commitInfo := range reducedCommitInfos {
		result = append(result, commitInfo)
	}
	sort.Sort(sortCommitInfos(result))
	return result
}

type sortCommitInfos []*CommitInfo

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
