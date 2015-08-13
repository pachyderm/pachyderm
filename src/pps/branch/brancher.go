package branch

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/concurrent"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type brancher struct {
	pfsAPIClient pfs.ApiClient
	storeClient  store.Client

	destroyable          concurrent.Destroyable
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
		concurrent.NewDestroyable(),
		make(map[string]string),
		&sync.RWMutex{},
	}
}

func (b *brancher) GetOutputCommitID(
	inputRepositoryName string,
	inputCommitID string,
	outputRepositoryName string,
) (string, error) {
	value, err := b.destroyable.Do(func() (interface{}, error) {
		return b.getOutputCommitID(inputRepositoryName, inputCommitID, outputRepositoryName)
	})
	if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (b *brancher) getOutputCommitID(
	inputRepositoryName string,
	inputCommitID string,
	outputRepositoryName string,
) (interface{}, error) {
	return nil, nil
}

func (b *brancher) CommitOutstanding() error {
	if err := b.destroyable.Destroy(); err != nil {
		return err
	}
	return nil
}
