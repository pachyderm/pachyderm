package sync

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

func PullV2(pachClient *client.APIClient, file *pfs.File, storageRoot string) error {
	r, err := pachClient.GetTarV2(file.Commit.Repo.Name, file.Commit.ID, file.Path)
	if err != nil {
		return err
	}
	return tarutil.TarToLocal(storageRoot, r)
}
