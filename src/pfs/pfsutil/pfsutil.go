/*
Package pfsutil provides utility functions that wrap a pfs.APIClient
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

const chunkSize = 4096

func CreateRepo(apiClient pfs.APIClient, repoName string) error {
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

func InspectRepo(apiClient pfs.APIClient, repoName string) (*pfs.RepoInfo, error) {
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

func ListRepo(apiClient pfs.APIClient) ([]*pfs.RepoInfo, error) {
	repoInfos, err := apiClient.ListRepo(
		context.Background(),
		&pfs.ListRepoRequest{},
	)
	if err != nil {
		return nil, err
	}
	return repoInfos.RepoInfo, nil
}

func DeleteRepo(apiClient pfs.APIClient, repoName string) error {
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

func StartCommit(apiClient pfs.APIClient, repoName string, parentCommit string) (*pfs.Commit, error) {
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

func FinishCommit(apiClient pfs.APIClient, repoName string, commitID string) error {
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

func InspectCommit(apiClient pfs.APIClient, repoName string, commitID string) (*pfs.CommitInfo, error) {
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

func ListCommit(apiClient pfs.APIClient, repoName string) ([]*pfs.CommitInfo, error) {
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

func DeleteCommit(apiClient pfs.APIClient, repoName string, commitID string) error {
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

func PutBlock(apiClient pfs.APIClient, repoName string, commitID string, path string, reader io.Reader) (*pfs.Block, error) {
	value, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return apiClient.PutBlock(
		context.Background(),
		&pfs.PutBlockRequest{
			File: &pfs.File{
				Commit: &pfs.Commit{
					Repo: &pfs.Repo{
						Name: repoName,
					},
					Id: commitID,
				},
				Path: path,
			},
			Value: value,
		},
	)
}

func GetBlock(apiClient pfs.APIClient, hash string, writer io.Writer) error {
	apiGetBlockClient, err := apiClient.GetBlock(
		context.Background(),
		&pfs.GetBlockRequest{
			Block: &pfs.Block{
				Hash: hash,
			},
		},
	)
	if err != nil {
		return err
	}
	if err := protostream.WriteFromStreamingBytesClient(apiGetBlockClient, writer); err != nil {
		return err
	}
	return nil
}

func InspectBlock(apiClient pfs.APIClient, hash string) (*pfs.BlockInfo, error) {
	blockInfo, err := apiClient.InspectBlock(
		context.Background(),
		&pfs.InspectBlockRequest{
			Block: &pfs.Block{
				Hash: hash,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func ListBlock(apiClient pfs.APIClient, shard uint64, modulus uint64) ([]*pfs.BlockInfo, error) {
	blockInfos, err := apiClient.ListBlock(
		context.Background(),
		&pfs.ListBlockRequest{
			Shard: &pfs.Shard{
				Number: shard,
				Modulo: modulus,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return blockInfos.BlockInfo, nil
}

func PutFile(apiClient pfs.APIClient, repoName string, commitID string, path string, offset int64, reader io.Reader) (_ int, retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return 0, err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	request := pfs.PutFileRequest{
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
	}
	var size int
	for {
		value := make([]byte, chunkSize)
		iSize, err := reader.Read(value)
		request.Value = value[0:iSize]
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
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

func GetFile(apiClient pfs.APIClient, repoName string, commitID string, path string, offset int64, size int64, writer io.Writer) error {
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

func InspectFile(apiClient pfs.APIClient, repoName string, commitID string, path string) (*pfs.FileInfo, error) {
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

func ListFile(apiClient pfs.APIClient, repoName string, commitID string, path string, shard uint64, modulus uint64) ([]*pfs.FileInfo, error) {
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

func ListChange(apiClient pfs.APIClient, repoName string, commitID string, path string, shard uint64, modulus uint64) ([]*pfs.Change, error) {
	changes, err := apiClient.ListChange(
		context.Background(),
		&pfs.ListChangeRequest{
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
	return changes.Change, nil
}

func DeleteFile(apiClient pfs.APIClient, repoName string, commitID string, path string) error {
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

func MakeDirectory(apiClient pfs.APIClient, repoName string, commitID string, path string) (retErr error) {
	putFileClient, err := apiClient.PutFile(context.Background())
	if err != nil {
		return err
	}
	defer func() {
		if _, err := putFileClient.CloseAndRecv(); err != nil && retErr != nil {
			retErr = err
		}
	}()
	return putFileClient.Send(
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
}

func InspectServer(apiClient pfs.APIClient, serverID string) (*pfs.ServerInfo, error) {
	return apiClient.InspectServer(
		context.Background(),
		&pfs.InspectServerRequest{
			Server: &pfs.Server{
				Id: serverID,
			},
		},
	)
}

func ListServer(apiClient pfs.APIClient) ([]*pfs.ServerInfo, error) {
	serverInfos, err := apiClient.ListServer(
		context.Background(),
		&pfs.ListServerRequest{},
	)
	if err != nil {
		return nil, err
	}
	return serverInfos.ServerInfo, nil
}

func PullDiff(internalAPIClient pfs.InternalAPIClient, repoName string, commitID string, shard uint64, writer io.Writer) error {
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

func PushDiff(internalAPIClient pfs.InternalAPIClient, repoName string, commitID string, shard uint64, reader io.Reader) error {
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
