package pfs

import (
)

const ChunkSize = 1024 * 1024

func NewRepo(repoName string) *Repo {
	return &Repo{Name: repoName}
}

func NewCommit(repoName string, commitID string) *Commit {
	return &Commit{
		Repo: NewRepo(repoName),
		ID:   commitID,
	}
}

func NewFile(repoName string, commitID string, path string) *File {
	return &File{
		Commit: NewCommit(repoName, commitID),
		Path:   path,
	}
}

func NewBlock(hash string) *Block {
	return &Block{
		Hash: hash,
	}
}

func NewDiff(repoName string, commitID string, shard uint64) *Diff {
	return &Diff{
		Commit: NewCommit(repoName, commitID),
		Shard:  shard,
	}
}

func NewFromCommit(repoName string, fromCommitID string) *Commit {
	if fromCommitID != "" {
		return NewCommit(repoName, fromCommitID)
	}
	return nil
}
