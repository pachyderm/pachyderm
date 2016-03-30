package pfs

import (
	"io"
	"math"

	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
)

const chunkSize = 1024 * 1024

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

func CreateRepo(apiClient APIClient, repoName string) error {
	_, err := apiClient.CreateRepo(
		context.Background(),
		&CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func InspectRepo(apiClient APIClient, repoName string) (*RepoInfo, error) {
	repoInfo, err := apiClient.InspectRepo(
		context.Background(),
		&InspectRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func ListRepo(apiClient APIClient) ([]*RepoInfo, error) {
	repoInfos, err := apiClient.ListRepo(
		context.Background(),
		&ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func DeleteRepo(apiClient APIClient, repoName string) error {
	_, err := apiClient.DeleteRepo(
		context.Background(),
		&DeleteRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func StartCommit(apiClient APIClient, repoName string, parentCommit string, branch string) (*Commit, error) {
	commit, err := apiClient.StartCommit(
		context.Background(),
		&StartCommitRequest{
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

func FinishCommit(apiClient APIClient, repoName string, commitID string) error {
	_, err := apiClient.FinishCommit(
		context.Background(),
		&FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func InspectCommit(apiClient APIClient, repoName string, commitID string) (*CommitInfo, error) {
	commitInfo, err := apiClient.InspectCommit(
		context.Background(),
		&InspectCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

func ListCommit(apiClient APIClient, repoNames []string, fromCommitIDs []string, block bool) ([]*CommitInfo, error) {
	var repos []*Repo
	for _, repoName := range repoNames {
		repos = append(repos, &Repo{Name: repoName})
	}
	var fromCommits []*Commit
	for i, fromCommitID := range fromCommitIDs {
		fromCommits = append(fromCommits, &Commit{
			Repo: repos[i],
			ID: fromCommitID,
		})
	}
	commitInfos, err := apiClient.ListCommit(
		context.Background(),
		&ListCommitRequest{
			Repo: repos,
			FromCommit: fromCommits,
			Block: block,
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func ListBranch(apiClient APIClient, repoName string) ([]*CommitInfo, error) {
	commitInfos, err := apiClient.ListBranch(
		context.Background(),
		&ListBranchRequest{
			Repo: NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func DeleteCommit(apiClient APIClient, repoName string, commitID string) error {
	_, err := apiClient.DeleteCommit(
		context.Background(),
		&DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func PutBlock(apiClient BlockAPIClient, reader io.Reader) (*BlockRefs, error) {
	putBlockClient, err := apiClient.PutBlock(context.Background())
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(protostream.NewStreamingBytesWriter(putBlockClient), reader); err != nil {
		return nil, err
	}
	return putBlockClient.CloseAndRecv()
}

func GetBlock(apiClient BlockAPIClient, hash string, offsetBytes uint64, sizeBytes uint64) (io.Reader, error) {
	apiGetBlockClient, err := apiClient.GetBlock(
		context.Background(),
		&GetBlockRequest{
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

func DeleteBlock(apiClient BlockAPIClient, block *Block) error {
	_, err := apiClient.DeleteBlock(
		context.Background(),
		&DeleteBlockRequest{
			Block: block,
		},
	)

	return err
}

func InspectBlock(apiClient BlockAPIClient, hash string) (*BlockInfo, error) {
	blockInfo, err := apiClient.InspectBlock(
		context.Background(),
		&InspectBlockRequest{
			Block: NewBlock(hash),
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func ListBlock(apiClient BlockAPIClient) ([]*BlockInfo, error) {
	blockInfos, err := apiClient.ListBlock(
		context.Background(),
		&ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

func PutFile(apiClient APIClient, repoName string, commitID string, path string, offset int64, reader io.Reader) (_ int, retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return 0, err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request := PutFileRequest{
		File:        NewFile(repoName, commitID, path),
		FileType:    FileType_FILE_TYPE_REGULAR,
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

func GetFile(apiClient APIClient, repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *Shard, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&GetFileRequest{
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

func InspectFile(apiClient APIClient, repoName string, commitID string, path string, fromCommitID string, shard *Shard) (*FileInfo, error) {
	fileInfo, err := apiClient.InspectFile(
		context.Background(),
		&InspectFileRequest{
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

func ListFile(apiClient APIClient, repoName string, commitID string, path string, fromCommitID string, shard *Shard) ([]*FileInfo, error) {
	fileInfos, err := apiClient.ListFile(
		context.Background(),
		&ListFileRequest{
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

func DeleteFile(apiClient APIClient, repoName string, commitID string, path string) error {
	_, err := apiClient.DeleteFile(
		context.Background(),
		&DeleteFileRequest{
			File:     NewFile(repoName, commitID, path),
		},
	)
	return err
}

func MakeDirectory(apiClient APIClient, repoName string, commitID string, path string) (retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return putFileClient.Send(
		&PutFileRequest{
			File:     NewFile(repoName, commitID, path),
			FileType: FileType_FILE_TYPE_DIR,
		},
	)
}


func newFromCommit(repoName string, fromCommitID string) *Commit {
	if fromCommitID != "" {
		return NewCommit(repoName, fromCommitID)
	}
	return nil
}
