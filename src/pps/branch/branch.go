package branch //import "go.pachyderm.com/pachyderm/src/pps/branch"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/timing"
	"go.pachyderm.com/pachyderm/src/pps/store"
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
	timer timing.Timer,
) Brancher {
	return newBrancher(
		pfsAPIClient,
		storeClient,
		timer,
	)
}
