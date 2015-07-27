package pfsutil

import (
	"bytes"
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

func PutFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, reader io.Reader) (int, error) {
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
			Value: value,
		},
	)
	return len(value), err
}

func GetFile(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) (io.Reader, error) {
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
		},
	)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	if err := protoutil.WriteFromStreamingBytesClient(apiGetFileClient, buffer); err != nil {
		return nil, err
	}
	return buffer, nil
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

func ListFiles(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, shardNum int, shardModulo int) (*pfs.ListFilesResponse, error) {
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
				Number: uint64(shardNum),
				Modulo: uint64(shardModulo),
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
