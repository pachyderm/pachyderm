// Package sync provides utility functions similar to `git pull/push` for PFS
package sync

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	pachclient "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	"go.pedge.io/lion"
	protostream "go.pedge.io/proto/stream"
	"golang.org/x/sync/errgroup"
)

// Pull clones an entire repo at a certain commit
//
// root is the local path you want to clone to
// commit is the commit you want to clone
// shard and diffMethod get passed to ListFile and GetFile. See documentations
// for those functions for details on these arguments.
// pipes causes the function to create named pipes in place of files, thus
// lazily downloading the data as it's needed
func Pull(ctx context.Context, client pfs.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard, pipes bool) error {
	return pullDir(ctx, client, root, commit, diffMethod, shard, "/", pipes)
}

func pullDir(ctx context.Context, client pfs.APIClient, root string, commit *pfs.Commit, diffMethod *pfs.DiffMethod, shard *pfs.Shard, dir string, pipes bool) error {
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
		g.Go(func() (retErr error) {
			switch fileInfo.FileType {
			case pfs.FileType_FILE_TYPE_REGULAR:
				path := filepath.Join(root, fileInfo.File.Path)
				request := &pfs.GetFileRequest{
					File: &pfs.File{
						Commit: commit,
						Path:   fileInfo.File.Path,
					},
					Shard:      shard,
					DiffMethod: diffMethod,
				}
				if pipes {
					if err := syscall.Mkfifo(path, 0666); err != nil {
						return err
					}
					// This goro will block until the user's code opens the
					// fifo.  That means we need to "abandon" this goro so that
					// the function can return and the caller can execute the
					// user's code. Waiting for this goro to return would
					// produce a deadlock.
					go func() {
						f, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
						if err != nil {
							lion.Printf("error opening %s: %s", path, err)
							return
						}
						defer func() {
							if err := f.Close(); err != nil {
								lion.Printf("error closing %s: %s", path, err)
							}
						}()
						getFileClient, err := client.GetFile(ctx, request)
						if err != nil {
							lion.Printf("error from GetFile: %s", err)
							return
						}
						if err := protostream.WriteFromStreamingBytesClient(getFileClient, f); err != nil {
							lion.Printf("error streaming data: %s", err)
							return
						}
					}()
				} else {
					f, err := os.Create(path)
					if err != nil {
						return err
					}
					defer func() {
						if err := f.Close(); err != nil && retErr == nil {
							retErr = err
						}
					}()
					getFileClient, err := client.GetFile(ctx, request)
					if err != nil {
						return err
					}
					return protostream.WriteFromStreamingBytesClient(getFileClient, f)
				}
			case pfs.FileType_FILE_TYPE_DIR:
				return pullDir(ctx, client, root, commit, diffMethod, shard, fileInfo.File.Path, pipes)
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
		g.Go(func() (retErr error) {
			if path == root || info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()

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
			_, err = pclient.PutFile(commit.Repo.Name, commit.ID, relPath, f)
			return err
		})
		return nil
	}); err != nil {
		return err
	}

	return g.Wait()
}

// PushObj pushes data from commit to an object store.
func PushObj(pachClient pachclient.APIClient, commit *pfs.Commit, objClient obj.Client, root string) error {
	var eg errgroup.Group
	if err := pachClient.Walk(commit.Repo.Name, commit.ID, "", "", false, nil, func(fileInfo *pfs.FileInfo) error {
		if fileInfo.FileType != pfs.FileType_FILE_TYPE_REGULAR {
			return nil
		}
		eg.Go(func() (retErr error) {
			w, err := objClient.Writer(filepath.Join(root, fileInfo.File.Path))
			if err != nil {
				return err
			}
			defer func() {
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			pachClient.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, "", false, nil, w)
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}

func PushSQL(pachClient pachclient.APIClient, commit *pfs.Commit, db *sql.DB) error {
	var eg errgroup.Group
	if err := pachClient.Walk(commit.Repo.Name, commit.ID, "", "", false, nil, func(fileInfo *pfs.FileInfo) error {
		if fileInfo.FileType != pfs.FileType_FILE_TYPE_REGULAR {
			return nil
		}
		eg.Go(func() (retErr error) {
			r, w := io.Pipe()
			defer func() {
				if err := r.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			go func() {
				if err := pachClient.GetFile(commit.Repo.Name, commit.ID, fileInfo.File.Path, 0, 0, "", false, nil, w); err != nil && retErr == nil {
					retErr = err
				}
				if err := w.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			csvR := csv.NewReader(r)
			keys, err := csvR.Read()
			if err == io.EOF {
				// tolerate empty files
				return nil
			}
			if err != nil {
				return err
			}
			for {
				vals, err := csvR.Read()
				if err == io.EOF {
					break
				}
				if err != nil {
					return err
				}
				query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", fileInfo.File.Path, strings.Join(keys, ","), strings.Join(vals, ","))
				if _, err := db.Exec(query); err != nil {
					return err
				}
			}
			return nil
		})
		return nil
	}); err != nil {
		return err
	}
	return eg.Wait()
}
