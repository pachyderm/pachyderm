package pfs

func Reduce(commitInfos []*CommitInfo) []*CommitInfo {
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
		reducedCommitInfo.CommitBytes += commitInfo.CommitBytes
		reducedCommitInfo.TotalBytes += commitInfo.TotalBytes
	}
	var result []*CommitInfo
	for _, commitInfo := range reducedCommitInfos {
		result = append(result, commitInfo)
	}
	return result
}
