package branch //import "go.pachyderm.com/pachyderm/src/pps/branch"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/store"
	"go.pedge.io/pkg/time"
)

type Brancher interface {
	GetOutputCommitID(
		inputRepositoryName string,
		inputCommitID string,
		outputRepositoryName string,
	) (string, error)
	CommitOutstanding() error
	// TODO(pedge)
	//DeleteOutstanding() error
}

func NewBrancher(
	pfsAPIClient pfs.ApiClient,
	storeClient store.Client,
	timer pkgtime.Timer,
) Brancher {
	return newBrancher(
		pfsAPIClient,
		storeClient,
		timer,
	)
}
