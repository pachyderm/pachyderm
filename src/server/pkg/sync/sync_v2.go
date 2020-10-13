package sync

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

// PullV2 pulls a file from PFS and stores it in the local filesystem.
func PullV2(pachClient *client.APIClient, file *pfs.File, storageRoot string, cb ...func(*tar.Header) error) error {
	r, err := pachClient.GetTarV2(file.Commit.Repo.Name, file.Commit.ID, file.Path)
	if err != nil {
		return err
	}
	return tarutil.Import(storageRoot, r, cb...)
}
