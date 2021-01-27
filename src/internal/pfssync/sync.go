package pfssync

import (
	"archive/tar"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/src/pfs"
)

// Pull pulls a file from PFS and stores it in the local filesystem.
func Pull(pachClient *client.APIClient, file *pfs.File, storageRoot string, cb ...func(*tar.Header) error) error {
	r, err := pachClient.GetTarFile(file.Commit.Repo.Name, file.Commit.ID, file.Path)
	if err != nil {
		return err
	}
	return tarutil.Import(storageRoot, r, cb...)
}
