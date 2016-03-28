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

//type PfsAPIClient pfs.APIClient
type PpsAPIClient pps.APIClient

type PfsAPIClient struct {
	internalClient pfs.APIClient
	internalBlockClient pfs.BlockAPIClient
}

type APIClient struct {
	PfsAPIClient
	PpsAPIClient
}

func NewFromAddress(pachAddr string) (*APIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	newpfsclient := PfsAPIClient{
		internalClient: pfs.NewAPIClient(clientConn),
		internalBlockClient: pfs.NewBlockAPIClient(clientConn),
	}

	return &APIClient{
		newpfsclient,
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


func (c *PfsAPIClient) CreateRepo(repoName string) error {
	_, err := c.internalClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: pfs.NewRepo(repoName),
		},
	)
	return err
}

func (c *PfsAPIClient) InspectRepo(repoName string) (*pfs.RepoInfo, error) {
	repoInfo, err := c.internalClient.InspectRepo(
		context.Background(),
		&pfs.InspectRepoRequest{
			Repo: pfs.NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func (c *PfsAPIClient) ListRepo() ([]*pfs.RepoInfo, error) {
	repoInfos, err := c.internalClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func (c *PfsAPIClient) DeleteRepo(repoName string) error {
	_, err := c.internalClient.DeleteRepo(
		context.Background(),
		&pfs.DeleteRepoRequest{
			Repo: pfs.NewRepo(repoName),
		},
	)
	return err
}

func (c *PfsAPIClient) StartCommit(repoName string, parentCommit string, branch string) (*pfs.Commit, error) {
	commit, err := c.internalClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Repo:     pfs.NewRepo(repoName),
			ParentID: parentCommit,
			Branch:   branch,
		},
	)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (c *PfsAPIClient) FinishCommit(repoName string, commitID string) error {
	_, err := c.internalClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: pfs.NewCommit(repoName, commitID),
		},
	)
	return err
}

func (c *PfsAPIClient) InspectCommit(repoName string, commitID string) (*pfs.CommitInfo, error) {
	commitInfo, err := c.internalClient.InspectCommit(
		context.Background(),
		&pfs.InspectCommitRequest{
			Commit: pfs.NewCommit(repoName, commitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

func (c *PfsAPIClient) ListCommit(repoNames []string) ([]*pfs.CommitInfo, error) {
	var repos []*pfs.Repo
	for _, repoName := range repoNames {
		repos = append(repos, &pfs.Repo{Name: repoName})
	}
	commitInfos, err := c.internalClient.ListCommit(
		context.Background(),
		&pfs.ListCommitRequest{
			Repo: repos,
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func (c *PfsAPIClient) ListBranch(repoName string) ([]*pfs.CommitInfo, error) {
	commitInfos, err := c.internalClient.ListBranch(
		context.Background(),
		&pfs.ListBranchRequest{
			Repo: pfs.NewRepo(repoName),
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func (c *PfsAPIClient) DeleteCommit(repoName string, commitID string) error {
	_, err := c.internalClient.DeleteCommit(
		context.Background(),
		&pfs.DeleteCommitRequest{
			Commit: pfs.NewCommit(repoName, commitID),
		},
	)
	return err
}

func (c *PfsAPIClient) PutBlock(apiClient pfs.BlockAPIClient, reader io.Reader) (*pfs.BlockRefs, error) {
	putBlockClient, err := c.internalBlockClient.PutBlock(context.Background())
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(protostream.NewStreamingBytesWriter(putBlockClient), reader); err != nil {
		return nil, err
	}
	return putBlockClient.CloseAndRecv()
}

func (c *PfsAPIClient) GetBlock(apiClient pfs.BlockAPIClient, hash string, offsetBytes uint64, sizeBytes uint64) (io.Reader, error) {
	apiGetBlockClient, err := c.internalBlockClient.GetBlock(
		context.Background(),
		&pfs.GetBlockRequest{
			Block:       pfs.NewBlock(hash),
			OffsetBytes: offsetBytes,
			SizeBytes:   sizeBytes,
		},
	)
	if err != nil {
		return nil, err
	}
	return protostream.NewStreamingBytesReader(apiGetBlockClient), nil
}

func (c *PfsAPIClient) InspectBlock(apiClient pfs.BlockAPIClient, hash string) (*pfs.BlockInfo, error) {
	blockInfo, err := c.internalBlockClient.InspectBlock(
		context.Background(),
		&pfs.InspectBlockRequest{
			Block: pfs.NewBlock(hash),
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func (c *PfsAPIClient) ListBlock(apiClient pfs.BlockAPIClient) ([]*pfs.BlockInfo, error) {
	blockInfos, err := c.internalBlockClient.ListBlock(
		context.Background(),
		&pfs.ListBlockRequest{},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

func (c *PfsAPIClient) PutFile(repoName string, commitID string, path string, offset int64, reader io.Reader) (_ int, retErr error) {
	putFileClient, err := c.internalClient.PutFile(context.Background())
	if err != nil {
		return 0, err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request := pfs.PutFileRequest{
		File:        pfs.NewFile(repoName, commitID, path),
		FileType:    pfs.FileType_FILE_TYPE_REGULAR,
		OffsetBytes: offset,
	}
	var size int
	eof := false
	for !eof {
		value := make([]byte, pfs.ChunkSize)
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

func (c *PfsAPIClient) GetFile(repoName string, commitID string, path string, offset int64, size int64, fromCommitID string, shard *pfs.Shard, writer io.Writer) error {
	if size == 0 {
		size = math.MaxInt64
	}
	apiGetFileClient, err := c.internalClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			File:        pfs.NewFile(repoName, commitID, path),
			Shard:       shard,
			OffsetBytes: offset,
			SizeBytes:   size,
			FromCommit:  pfs.NewFromCommit(repoName, fromCommitID),
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

func (c *PfsAPIClient) InspectFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard) (*pfs.FileInfo, error) {
	fileInfo, err := c.internalClient.InspectFile(
		context.Background(),
		&pfs.InspectFileRequest{
			File:       pfs.NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: pfs.NewFromCommit(repoName, fromCommitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func (c *PfsAPIClient) ListFile(repoName string, commitID string, path string, fromCommitID string, shard *pfs.Shard) ([]*pfs.FileInfo, error) {
	fileInfos, err := c.internalClient.ListFile(
		context.Background(),
		&pfs.ListFileRequest{
			File:       pfs.NewFile(repoName, commitID, path),
			Shard:      shard,
			FromCommit: pfs.NewFromCommit(repoName, fromCommitID),
		},
	)
	if err != nil {
		return nil, err
	}
	return fileInfos.FileInfo, nil
}

func (c *PfsAPIClient) DeleteFile(repoName string, commitID string, path string) error {
	_, err := c.internalClient.DeleteFile(
		context.Background(),
		&pfs.DeleteFileRequest{
			File:     pfs.NewFile(repoName, commitID, path),
		},
	)
	return err
}

func (c *PfsAPIClient) MakeDirectory(repoName string, commitID string, path string) (retErr error) {
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
		&pfs.PutFileRequest{
			File:     pfs.NewFile(repoName, commitID, path),
			FileType: pfs.FileType_FILE_TYPE_DIR,
		},
	)
}
