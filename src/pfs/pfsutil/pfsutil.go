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

func RepoCreate(apiClient pfs.ApiClient, repoName string) error {
	_, err := apiClient.RepoCreate(
		context.Background(),
		&pfs.RepoCreateRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	return err
}

func RepoInspect(apiClient pfs.ApiClient, repoName string) (*pfs.RepoInfo, error) {
	response, err := apiClient.RepoInspect(
		context.Background(),
		&pfs.RepoInspectRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return response.RepoInfo, nil
}

func RepoList(apiClient pfs.ApiClient) ([]*pfs.RepoInfo, error) {
	response, err := apiClient.RepoList(
		context.Background(),
		&pfs.RepoListRequest{},
	)
	if err != nil {
		return nil, err
	}
	return response.RepoInfo, nil
}

func RepoDelete(apiClient pfs.ApiClient, repoName string) error {
	_, err := apiClient.RepoDelete(
		context.Background(),
		&pfs.RepoDeleteRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	return err
}

func CommitStart(apiClient pfs.ApiClient, repoName string, parentCommit string) (*pfs.Commit, error) {
	response, err := apiClient.CommitStart(
		context.Background(),
		&pfs.CommitStartRequest{
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
	return response.Commit, nil
}

func CommitFinish(apiClient pfs.ApiClient, repoName string, commitID string) error {
	_, err := apiClient.CommitFinish(
		context.Background(),
		&pfs.CommitFinishRequest{
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

func CommitInspect(apiClient pfs.ApiClient, repoName string, commitID string) (*pfs.CommitInfo, error) {
	response, err := apiClient.CommitInspect(
		context.Background(),
		&pfs.CommitInspectRequest{
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
	return response.CommitInfo, nil
}

func CommitList(apiClient pfs.ApiClient, repoName string) ([]*pfs.CommitInfo, error) {
	response, err := apiClient.CommitList(
		context.Background(),
		&pfs.CommitListRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return response.CommitInfo, nil
}

func CommitDelete(apiClient pfs.ApiClient, repoName string, commitID string) error {
	_, err := apiClient.CommitList(
		context.Background(),
		&pfs.CommitListRequest{
			Repo: &pfs.Repo{
				Name: repoName,
			},
		},
	)
	return err
}

func FilePut(apiClient pfs.ApiClient, repoName string, commitID string, path string, offset int64, reader io.Reader) (int64, error) {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	_, err = apiClient.FilePut(
		context.Background(),
		&pfs.FilePutRequest{
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

func FileGet(apiClient pfs.ApiClient, repoName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
	apiGetFileClient, err := apiClient.FileGet(
		context.Background(),
		&pfs.FileGetRequest{
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

func FileInspect(apiClient pfs.ApiClient, repoName string, commitID string, path string) (*pfs.FileInfo, error) {
	response, err := apiClient.FileInspect(
		context.Background(),
		&pfs.FileInspectRequest{
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
	return response.FileInfo, nil
}

func FileList(apiClient pfs.ApiClient, repoName string, commitID string, path string, shard uint64, modulus uint64) ([]*pfs.FileInfo, error) {
	response, err := apiClient.FileList(
		context.Background(),
		&pfs.FileListRequest{
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
	return response.FileInfo, nil
}

func MakeDirectory(apiClient pfs.ApiClient, repoName string, commitID string, path string) error {
	_, err := apiClient.FilePut(
		context.Background(),
		&pfs.FilePutRequest{
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
