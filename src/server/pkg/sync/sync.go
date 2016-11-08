// Package sync provides utility functions similar to `git pull/push` for PFS
package sync

import (
	"context"
	"os"
	"path/filepath"

	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"

	protostream "go.pedge.io/proto/stream"
	"golang.org/x/sync/errgroup"
)

// Pull clones an entire repo at a certain commit
//
// root is the local path you want to clone to
// commit is the commit you want to clone
// shard and diffMethod get passed to ListFile and GetFile. See documentations
// for those functions for details on these arguments.
func Pull(ctx context.Context, client pfs.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard) error {
	return pullDir(ctx, client, root, commit, diffMethod, shard, "/")
}

func pullDir(ctx context.Context, client pfs.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard, dir string) error {
	if err := os.MkdirAll(filepath.Join(root, dir), 0777); err != nil {
		return err
	}

	fileInfos, err := client.ListFile(ctx, &pfs.ListFileRequest{
		File: &pfs.File{
			Commit: commit,
			Path:   dir,
		},
		Shard:      shard,
		DiffMethod: diffMethod,
		Mode:       pfs.ListFileMode_ListFile_NORMAL,
	})
	if err != nil {
		return err
	}

	var g errgroup.Group
	for _, fileInfo := range fileInfos.FileInfo {
		fileInfo := fileInfo
		g.Go(func() error {
			switch fileInfo.FileType {
			case pfs.FileType_FILE_TYPE_REGULAR:
				path := filepath.Join(root, fileInfo.File.Path)
				f, err := os.Create(path)
				if err != nil {
					return err
				}
				getFileClient, err := client.GetFile(
					ctx,
					&pfs.GetFileRequest{
						File: &pfs.File{
							Commit: commit,
							Path:   fileInfo.File.Path,
						},
						Shard:      shard,
						DiffMethod: diffMethod,
					},
				)
				if err != nil {
					return err
				}
				if err := protostream.WriteFromStreamingBytesClient(getFileClient, f); err != nil {
					return err
				}
				return f.Close()
			case pfs.FileType_FILE_TYPE_DIR:
				return pullDir(ctx, client, root, commit, diffMethod, shard, fileInfo.File.Path)
			}
			return nil
		})
	}
	return g.Wait()
}

// Push puts files under root into an open commit.
func Push(ctx context.Context, client pfs.APIClient, root string, commit *pfs.Commit, overwrite bool) error {
	var g errgroup.Group
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		g.Go(func() error {
			if path == root || info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if overwrite {
				if _, err := client.DeleteFile(ctx, &pfs.DeleteFileRequest{
					File: &pfs.File{
						Commit: commit,
						Path:   relPath,
					},
				}); err != nil {
					return err
				}
			}

			pclient := pachclient.APIClient{
				PfsAPIClient: client,
			}
			if _, err := pclient.PutFile(commit.Repo.Name, commit.ID, relPath, f); err != nil {
				return err
			}

			return f.Close()
		})
		return nil
	}); err != nil {
		return err
	}

	return g.Wait()
}
