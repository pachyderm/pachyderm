/*
Package pfsutil provides utility functions that wrap a pfs.APIClient
to make the calling code slightly cleaner.
*/
package pfsutil

import (
	"io"
	"math"

	client "github.com/pachyderm/pachyderm/src/client/pfs"
	server "github.com/pachyderm/pachyderm/src/pfs"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

const chunkSize = 1024 * 1024

func NewRepo(repoName string) *server.Repo {
	return &server.Repo{Name: repoName}
}

func NewCommit(repoName string, commitID string) *server.Commit {
	return &server.Commit{
		Repo: NewRepo(repoName),
		ID:   commitID,
	}
}

func NewFile(repoName string, commitID string, path string) *server.File {
	return &server.File{
		Commit: NewCommit(repoName, commitID),
		Path:   path,
	}
}

func NewBlock(hash string) *server.Block {
	return &server.Block{
		Hash: hash,
	}
}

func NewDiff(repoName string, commitID string, shard uint64) *server.Diff {
	return &server.Diff{
		Commit: NewCommit(repoName, commitID),
		Shard:  shard,
	}
}

func CreateRepo(apiClient client.APIClient, repoName string) error {
	_, err := apiClient.CreateRepo(
		context.Background(),
		&client.CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func InspectRepo(apiClient client.APIClient, repoName string) (*server.RepoInfo, error) {
	repoInfo, err := apiClient.InspectRepo(
		context.Background(),
		&client.InspectRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func ListRepo(apiClient client.APIClient) ([]*server.RepoInfo, error) {
	repoInfos, err := apiClient.ListRepo(
		context.Background(),
		&client.ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func DeleteRepo(apiClient client.APIClient, repoName string) error {
	_, err := apiClient.DeleteRepo(
		context.Background(),
		&client.DeleteRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func StartCommit(apiClient client.APIClient, repoName string, parentCommit string, branch string) (*server.Commit, error) {
	commit, err := apiClient.StartCommit(
		context.Background(),
		&client.StartCommitRequest{
			Repo:     NewRepo(repoName),
			ParentID: parentCommit,
			Branch:   branch,
		},
	)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func FinishCommit(apiClient client.APIClient, repoName string, commitID string) error {
	_, err := apiClient.FinishCommit(
		context.Background(),
		&client.FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func InspectCommit(apiClient client.APIClient, repoName string, commitID string) (*server.CommitInfo, error) {
	commitInfo, err := apiClient.InspectCommit(
		context.Background(),
		&client.InspectCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

func ListCommit(apiClient client.APIClient, repoNames []string) ([]*server.CommitInfo, error) {
	var repos []*server.Repo
	for _, repoName := range repoNames {
		repos = append(repos, &server.Repo{Name: repoName})
	}
	commitInfos, err := apiClient.ListCommit(
		context.Background(),
		&client.ListCommitRequest{
			Repo: repos,
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func ListBranch(apiClient client.APIClient, repoName string) ([]*server.CommitInfo, error) {
	commitInfos, err := apiClient.ListBranch(
		context.Background(),
		&client.ListBranchRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func DeleteCommit(apiClient client.APIClient, repoName string, commitID string) error {
	_, err := apiClient.DeleteCommit(
		context.Background(),
		&client.DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func PutBlock(apiClient client.BlockAPIClient, reader io.Reader) (*server.BlockRefs, error) {
	putBlockClient, err := apiClient.PutBlock(context.Background())
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(protostream.NewStreamingBytesWriter(putBlockClient), reader); err != nil {
		return nil, err
	}
	return putBlockClient.CloseAndRecv()
}

func GetBlock(apiClient client.BlockAPIClient, hash string, offsetBytes uint64, sizeBytes uint64) (io.Reader, error) {
	apiGetBlockClient, err := apiClient.GetBlock(
		context.Background(),
		&client.GetBlockRequest{
			Block:       NewBlock(hash),
			OffsetBytes: offsetBytes,
			SizeBytes:   sizeBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	return protostream.NewStreamingBytesReader(apiGetBlockClient), nil
}

func InspectBlock(apiClient client.BlockAPIClient, hash string) (*server.BlockInfo, error) {
	blockInfo, err := apiClient.InspectBlock(
		context.Background(),
		&client.InspectBlockRequest{
			Block: NewBlock(hash),
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func ListBlock(apiClient client.BlockAPIClient) ([]*server.BlockInfo, error) {
	blockInfos, err := apiClient.ListBlock(
		context.Background(),
		&client.ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

func PutFile(apiClient client.APIClient, repoName string, commitID string, path string, offset int64, reader io.Reader) (_ int, retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return 0, err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request := client.PutFileRequest{
		File:        NewFile(repoName, commitID, path),
		FileType:    server.FileType_FILE_TYPE_REGULAR,
		OffsetBytes: offset,
	}
	var size int
	eof := false
	for !eof {
		value := make([]byte, chunkSize)
		iSize, err := reader.Read(value)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if err == io.EOF {
			eof = true
		}
		request.Value = value[0:iSize]
		size += iSize
		if err := putFileClient.Send(&request); err != nil {
			return 0, err
		}
	}
	if err != nil && err != io.EOF {
		return 0, err
	}
	return size, err
}

func GetFile(apiClient client.APIClient, repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *server.Shard, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&client.GetFileRequest{
			File:        NewFile(repoName, commitID, path),
			Shard:       shard,
			OffsetBytes: offset,
			SizeBytes:   size,
			FromCommit:  newFromCommit(repoName, fromCommitID),
		},
	)
	if err != nil {
		return err
	}
	if err := protostream.WriteFromStreamingBytesClient(apiGetFileClient, writer); err != nil {
		return err
	}
	return nil
}

func InspectFile(apiClient client.APIClient, repoName string, commitID string, path string, fromCommitID string, shard *server.Shard) (*server.FileInfo, error) {
	fileInfo, err := apiClient.InspectFile(
		context.Background(),
		&client.InspectFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func ListFile(apiClient client.APIClient, repoName string, commitID string, path string, fromCommitID string, shard *server.Shard) ([]*server.FileInfo, error) {
	fileInfos, err := apiClient.ListFile(
		context.Background(),
		&client.ListFileRequest{
			File:       NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: newFromCommit(repoName, fromCommitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfos.FileInfo, nil
}

func DeleteFile(apiClient client.APIClient, repoName string, commitID string, path string) error {
	_, err := apiClient.DeleteFile(
		context.Background(),
		&client.DeleteFileRequest{
			File:     NewFile(repoName, commitID, path),
		},
	)
	return err
}

func newFromCommit(repoName string, fromCommitID string) *server.Commit {
	if fromCommitID != "" {
		return NewCommit(repoName, fromCommitID)
	}
	return nil
}
