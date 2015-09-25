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

func CreateRepo(apiClient pfs.ApiClient, repoName string) error {
	_, err := apiClient.CreateRepo(
		context.Background(),
		&pfs.CreateRepoRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	return err
}

func InspectRepo(apiClient pfs.ApiClient, repoName string) (*pfs.RepoInfo, error) {
	repoInfo, err := apiClient.InspectRepo(
		context.Background(),
		&pfs.InspectRepoRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return repoInfo, nil
}

func ListRepo(apiClient pfs.ApiClient) ([]*pfs.RepoInfo, error) {
	repoInfos, err := apiClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func DeleteRepo(apiClient pfs.ApiClient, repoName string) error {
	_, err := apiClient.DeleteRepo(
		context.Background(),
		&pfs.DeleteRepoRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	return err
}

func StartCommit(apiClient pfs.ApiClient, repoName string, parentCommit string) (*pfs.Commit, error) {
	commit, err := apiClient.StartCommit(
		context.Background(),
		&pfs.StartCommitRequest{
			Parent: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: parentCommit,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func FinishCommit(apiClient pfs.ApiClient, repoName string, commitID string) error {
	_, err := apiClient.FinishCommit(
		context.Background(),
		&pfs.FinishCommitRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: commitID,
			},
		},
	)
	return err
}

func InspectCommit(apiClient pfs.ApiClient, repoName string, commitID string) (*pfs.CommitInfo, error) {
	commitInfo, err := apiClient.InspectCommit(
		context.Background(),
		&pfs.InspectCommitRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: commitID,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfo, nil
}

func ListCommit(apiClient pfs.ApiClient, repoName string) ([]*pfs.CommitInfo, error) {
	commitInfos, err := apiClient.ListCommit(
		context.Background(),
		&pfs.ListCommitRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return commitInfos.CommitInfo, nil
}

func DeleteCommit(apiClient pfs.ApiClient, repoName string, commitID string) error {
	_, err := apiClient.DeleteCommit(
		context.Background(),
		&pfs.DeleteCommitRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: commitID,
			},
		},
	)
	return err
}

func PutFile(apiClient pfs.ApiClient, repoName string, commitID string, path string, offset int64, reader io.Reader) (int64, error) {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	_, err = apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
					},
					Id: commitID,
				},
				Path: path,
			},
			FileType:    pfs.FileType_FILE_TYPE_REGULAR,
			OffsetBytes: offset,
			Value:       value,
		},
	)
	return int64(len(value)), err
}

func GetFile(apiClient pfs.ApiClient, repoName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
	apiGetFileClient, err := apiClient.GetFile(
		context.Background(),
		&pfs.GetFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
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

func InspectFile(apiClient pfs.ApiClient, repoName string, commitID string, path string) (*pfs.FileInfo, error) {
	fileInfo, err := apiClient.InspectFile(
		context.Background(),
		&pfs.InspectFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
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
	return fileInfo, nil
}

func ListFile(apiClient pfs.ApiClient, repoName string, commitID string, path string, shard uint64, modulus uint64) ([]*pfs.FileInfo, error) {
	fileInfos, err := apiClient.ListFile(
		context.Background(),
		&pfs.ListFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
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
	if err != nil {
		return nil, err
	}
	return fileInfos.FileInfo, nil
}

func DeleteFile(apiClient pfs.ApiClient, repoName string, commitID string, path string) error {
	_, err := apiClient.DeleteFile(
		context.Background(),
		&pfs.DeleteFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
					},
					Id: commitID,
				},
				Path: path,
			},
		},
	)
	return err
}

func MakeDirectory(apiClient pfs.ApiClient, repoName string, commitID string, path string) error {
	_, err := apiClient.PutFile(
		context.Background(),
		&pfs.PutFileRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
					},
					Id: commitID,
				},
				Path: path,
			},
			FileType: pfs.FileType_FILE_TYPE_DIR,
		},
	)
	return err
}

func InspectServer(apiClient pfs.ApiClient, serverID string) (*pfs.ServerInfo, error) {
	return apiClient.InspectServer(
		context.Background(),
		&pfs.InspectServerRequest{
			Server: &pfs.Server{
				Id: serverID,
			},
		},
	)
}

func ListServer(apiClient pfs.ApiClient) ([]*pfs.ServerInfo, error) {
	serverInfos, err := apiClient.ListServer(
		context.Background(),
		&pfs.ListServerRequest{},
	)
	if err != nil {
		return nil, err
	}
	return serverInfos.ServerInfo, nil
}

func PullDiff(internalAPIClient pfs.InternalApiClient, repoName string, commitID string, shard uint64, writer io.Writer) error {
	apiPullDiffClient, err := internalAPIClient.PullDiff(
		context.Background(),
		&pfs.PullDiffRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
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

func PushDiff(internalAPIClient pfs.InternalApiClient, repoName string, commitID string, shard uint64, reader io.Reader) error {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = internalAPIClient.PushDiff(
		context.Background(),
		&pfs.PushDiffRequest{
			Commit: &pfs.Commit{
				Repo: &pfs.Repo{
					Name: repoName,
				},
				Id: commitID,
			},
			Shard: shard,
			Value: value,
		},
	)
	return err
}
