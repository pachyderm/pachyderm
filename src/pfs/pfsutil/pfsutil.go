/*
Package pfsutil provides utility functions that wrap a pfs.ApiClient
to make the calling code slightly cleaner.
*/
package pfsutil

import (
	"io"
	"io/ioutil"

	"go.pedge.io/proto/stream"

	"github.com/pachyderm/pachyderm/src/pfs"
	"golang.org/x/net/context"
)

const (
	GetAll int64 = 1<<63 - 1
)

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

func Branch(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.Commit, error) {
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
	if err := protostream.WriteFromStreamingBytesClient(apiGetFileClient, writer); err != nil {
		return err
	}
	return nil
}

func GetFileInfo(apiClient pfs.ApiClient, repositoryName string, commitID string, path string) (*pfs.FileInfo, error) {
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

func ListFiles(apiClient pfs.ApiClient, repositoryName string, commitID string, path string, shard uint64, modulus uint64) (*pfs.FileInfos, error) {
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

func Write(apiClient pfs.ApiClient, repositoryName string, commitID string) error {
	_, err := apiClient.Write(
		context.Background(),
		&pfs.WriteRequest{
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

func GetCommitInfo(apiClient pfs.ApiClient, repositoryName string, commitID string) (*pfs.CommitInfo, error) {
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

func ListCommits(apiClient pfs.ApiClient, repositoryName string) (*pfs.CommitInfos, error) {
	return apiClient.ListCommits(
		context.Background(),
		&pfs.ListCommitsRequest{
			Repository: &pfs.Repository{
				Name: repositoryName,
			},
		},
	)
}

func PullDiff(internalAPIClient pfs.InternalApiClient, repositoryName string, commitID string, shard uint64, writer io.Writer) error {
	apiPullDiffClient, err := internalAPIClient.PullDiff(
		context.Background(),
		&pfs.PullDiffRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
			Shard: shard,
		},
	)
	if err != nil {
		return err
	}
	if err := protostream.WriteFromStreamingBytesClient(apiPullDiffClient, writer); err != nil {
		return err
	}
	return nil
}

func PushDiff(internalAPIClient pfs.InternalApiClient, repositoryName string, commitID string, shard uint64, reader io.Reader) error {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = internalAPIClient.PushDiff(
		context.Background(),
		&pfs.PushDiffRequest{
			Commit: &pfs.Commit{
				Repository: &pfs.Repository{
					Name: repositoryName,
				},
				Id: commitID,
			},
			Shard: shard,
			Value: value,
		},
	)
	return err
}
