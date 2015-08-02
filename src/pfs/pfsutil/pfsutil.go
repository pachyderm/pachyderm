package pfsutil

import (
	"io"
	"io/ioutil"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/peter-edge/go-google-protobuf"
	"golang.org/x/net/context"
)

func GetVersion(apiClient pfs.ApiClient) (*pfs.GetVersionResponse, error) {
	return apiClient.GetVersion(
		context.Background(),
		&google_protobuf.Empty{},
	)
}

func InitRepository(apiClient pfs.ApiClient, repositoryName string) error {
	_, err := apiClient.InitRepository(
		context.Background(),
		&pfs.InitRepositoryRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
	return err
}

func Branch(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.BranchResponse, error) {
	return apiClient.Branch(
		context.Background(),
		&pfs.BranchRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

func MakeDirectory(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) error {
	_, err := apiClient.MakeDirectory(
		context.Background(),
		&pfs.MakeDirectoryRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	return err
}

func PutFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, offset int64, reader io.Reader) (int64, error) {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	_, err = apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			OffsetBytes: offset,
			Value:       value,
		},
	)
	return int64(len(value)), err
}

func GetFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			OffsetBytes: offset,
			SizeBytes:   size,
		},
	)
	if err != nil {
		return err
	}
	if err := protoutil.WriteFromStreamingBytesClient(apiGetFileClient, writer); err != nil {
		return err
	}
	return nil
}

func GetFileInfo(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) (*pfs.GetFileInfoResponse, error) {
	return apiClient.GetFileInfo(
		context.Background(),
		&pfs.GetFileInfoRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
}

func ListFiles(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, shard uint64, modulus uint64) (*pfs.ListFilesResponse, error) {
	return apiClient.ListFiles(
		context.Background(),
		&pfs.ListFilesRequest{
			Path: &pfs.Path{
				Commit: &pfs.Commit{
					Repository: &pfs.Repository{
						Name: repositoryName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Shard: &pfs.Shard{
				Number: shard,
				Modulo: modulus,
			},
		},
	)
}

func Commit(apiClient pfs.ApiClient, repositoryName string, commitID string) error {
	_, err := apiClient.Commit(
		context.Background(),
		&pfs.CommitRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
	return err
}

func GetCommitInfo(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.GetCommitInfoResponse, error) {
	return apiClient.GetCommitInfo(
		context.Background(),
		&pfs.GetCommitInfoRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
		},
	)
}

func ListCommits(apiClient pfs.ApiClient, repositoryName string) (*pfs.ListCommitsResponse, error) {
	return apiClient.ListCommits(
		context.Background(),
		&pfs.ListCommitsRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
}
