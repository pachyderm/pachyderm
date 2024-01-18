package testing

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

func finishProjectCommit(pachClient *client.APIClient, project, repo, branch, id string) error {
	if err := pachClient.FinishCommit(project, repo, branch, id); err != nil {
		if !pfsserver.IsCommitFinishedErr(err) {
			return err
		}
	}
	_, err := pachClient.WaitCommit(project, repo, branch, id)
	return err
}

func finishCommit(pachClient *client.APIClient, repo, branch, id string) error {
	return finishProjectCommit(pachClient, pfs.DefaultProjectName, repo, branch, id)
}
