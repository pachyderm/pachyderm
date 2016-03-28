package client

import(
	"fmt"
	"os"
	"io"
	"math"

	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/grpc"	
	"go.pedge.io/proto/version"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 0
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 10
	// MicroVersion is the current micro version for pachyderm.
	MicroVersion = 0
	// AdditionalVersion will be "dev" is this is a development branch, "" otherwise.
	AdditionalVersion = "RC1"
)

var (
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
	}
)

type PfsAPIClient pfs.APIClient
type PpsAPIClient pps.APIClient

type pfsAPIClient struct {
	internalClient pfs.APIClient
}

type APIClient struct {
	pfsAPIClient
	PpsAPIClient
}

func NewFromAddress(pachAddr string) (*APIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIClient{
		&pfsAPIClient{ internalClient: pfs.NewAPIClient(clientConn)},
		pps.NewAPIClient(clientConn),
	}, nil
}

func New() (*APIClient, error) {
	pachAddr := os.Getenv("PACHD_PORT_650_TCP_ADDR")

	if pachAddr == "" {
		return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
	}

	return NewFromAddress(fmt.Sprintf("%v:650",pachAddr))
}

//// Syntactic Sugar for client


func (c *pfsAPIClient) CreateRepo(repoName string) error {
	_, err := c.internalClient.CreateRepo(
		context.Background(),
		&CreateRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func (c *pfsAPIClient) InspectRepo(repoName string) (*RepoInfo, error) {
	repoInfo, err := c.internalClient.InspectRepo(
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

func (c *pfsAPIClient) ListRepo() ([]*RepoInfo, error) {
	repoInfos, err := c.internalClient.ListRepo(
		context.Background(),
		&ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func (c *pfsAPIClient) DeleteRepo(repoName string) error {
	_, err := c.internalClient.DeleteRepo(
		context.Background(),
		&DeleteRepoRequest{
			Repo: NewRepo(repoName),
		},
	)
	return err
}

func (c *pfsAPIClient) StartCommit(repoName string, parentCommit string, branch string) (*Commit, error) {
	commit, err := c.internalClient.StartCommit(
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

func (c *pfsAPIClient) FinishCommit(repoName string, commitID string) error {
	_, err := c.internalClient.FinishCommit(
		context.Background(),
		&FinishCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func (c *pfsAPIClient) InspectCommit(repoName string, commitID string) (*CommitInfo, error) {
	commitInfo, err := c.internalClient.InspectCommit(
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

func (c *pfsAPIClient) ListCommit(repoNames []string) ([]*CommitInfo, error) {
	var repos []*Repo
	for _, repoName := range repoNames {
		repos = append(repos, &Repo{Name: repoName})
	}
	commitInfos, err := c.internalClient.ListCommit(
		context.Background(),
		&ListCommitRequest{
			Repo: repos,
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func (c *pfsAPIClient) ListBranch(repoName string) ([]*CommitInfo, error) {
	commitInfos, err := c.internalClient.ListBranch(
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

func (c *pfsAPIClient) DeleteCommit(repoName string, commitID string) error {
	_, err := c.internalClient.DeleteCommit(
		context.Background(),
		&DeleteCommitRequest{
			Commit: NewCommit(repoName, commitID),
		},
	)
	return err
}

func (c *pfsAPIClient) PutBlock(apiClient BlockAPIClient, reader io.Reader) (*BlockRefs, error) {
	putBlockClient, err := c.internalClient.PutBlock(context.Background())
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(protostream.NewStreamingBytesWriter(putBlockClient), reader); err != nil {
		return nil, err
	}
	return putBlockClient.CloseAndRecv()
}

func (c *pfsAPIClient) GetBlock(apiClient BlockAPIClient, hash string, offsetBytes uint64, sizeBytes uint64) (io.Reader, error) {
	apiGetBlockClient, err := c.internalClient.GetBlock(
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

func (c *pfsAPIClient) InspectBlock(apiClient BlockAPIClient, hash string) (*BlockInfo, error) {
	blockInfo, err := c.internalClient.InspectBlock(
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

func (c *pfsAPIClient) ListBlock(apiClient BlockAPIClient) ([]*BlockInfo, error) {
	blockInfos, err := c.internalClient.ListBlock(
		context.Background(),
		&ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

func (c *pfsAPIClient) PutFile(repoName string, commitID string, path string, offset int64, reader io.Reader) (_ int, retErr error) {
	putFileClient, err := c.internalClient.PutFile(context.Background())
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

func (c *pfsAPIClient) GetFile(repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *Shard, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := c.internalClient.GetFile(
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

func (c *pfsAPIClient) InspectFile(repoName string, commitID string, path string, fromCommitID string, shard *Shard) (*FileInfo, error) {
	fileInfo, err := c.internalClient.InspectFile(
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

func (c *pfsAPIClient) ListFile(repoName string, commitID string, path string, fromCommitID string, shard *Shard) ([]*FileInfo, error) {
	fileInfos, err := c.internalClient.ListFile(
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

func (c *pfsAPIClient) DeleteFile(repoName string, commitID string, path string) error {
	_, err := c.internalClient.DeleteFile(
		context.Background(),
		&DeleteFileRequest{
			File:     NewFile(repoName, commitID, path),
		},
	)
	return err
}

func (c *pfsAPIClient) MakeDirectory(repoName string, commitID string, path string) (retErr error) {
	putFileClient, err := c.internalClient.PutFile(context.Background())
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
