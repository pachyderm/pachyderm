package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

func reduce(commitInfos []*pfs.CommitInfo) []*pfs.CommitInfo {
	reducedCommitInfos := make(map[string]*pfs.CommitInfo)
	for _, commitInfo := range commitInfos {
		reducedCommitInfo, ok := reducedCommitInfos[commitInfo.Commit.Id]
		if !ok {
			reducedCommitInfos[commitInfo.Commit.Id] = commitInfo
			continue
		}
		if commitInfo.CommitType == pfs.CommitType_COMMIT_TYPE_WRITE {
			reducedCommitInfo.CommitType = pfs.CommitType_COMMIT_TYPE_WRITE
		}
		reducedCommitInfo.CommitBytes += commitInfo.CommitBytes
		reducedCommitInfo.TotalBytes += commitInfo.TotalBytes
	}
	var result []*pfs.CommitInfo
	for _, commitInfo := range reducedCommitInfos {
		result = append(result, commitInfo)
	}
	return result
}
