package branch

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type brancher struct {
	pfsAPIClient pfs.ApiClient
	storeClient  store.Client
}

func newBrancher(
	pfsAPIClient pfs.ApiClient,
	storeClient store.Client,
) *brancher {
	return &brancher{
		pfsAPIClient,
		storeClient,
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
