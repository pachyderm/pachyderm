package branch

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type brancher struct {
	pfsAPIClient pfs.ApiClient
	storeClient  store.Client

	repositoryToBranchID map[string]string
	lock                 *sync.RWMutex
}

func newBrancher(
	pfsAPIClient pfs.ApiClient,
	storeClient store.Client,
) *brancher {
	return &brancher{
		pfsAPIClient,
		storeClient,
		make(map[string]string),
		&sync.RWMutex{},
	}
}

func (b *brancher) GetOutputCommitID(
	inputRepositoryName string,
	inputCommitID string,
	outputRepositoryName string,
) (string, error) {
	return "", nil
}

func (b *brancher) CommitOutstanding() error {
	return nil
}
