package branch

import (
	"sync"

	"go.pedge.io/proto/time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pkg/concurrent"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type repositoryCommit struct {
	repositoryName string
	commitID       string
}

type brancher struct {
	pfsAPIClient pfs.ApiClient
	storeClient  store.Client
	timer        timing.Timer

	destroyable                         concurrent.Destroyable
	outputRepositoryToInputRepositories map[string]map[repositoryCommit]bool
	outputRepositoryToBranchID          map[string]string
	lock                                *sync.RWMutex
}

func newBrancher(
	pfsAPIClient pfs.ApiClient,
	storeClient store.Client,
	timer timing.Timer,
) *brancher {
	return &brancher{
		pfsAPIClient,
		storeClient,
		timer,
		concurrent.NewDestroyable(),
		make(map[string]map[repositoryCommit]bool),
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
	b.lock.Lock()
	if _, ok := b.outputRepositoryToInputRepositories[outputRepositoryName]; !ok {
		b.outputRepositoryToInputRepositories[outputRepositoryName] = make(map[repositoryCommit]bool)
	}
	b.outputRepositoryToInputRepositories[outputRepositoryName][repositoryCommit{
		repositoryName: inputRepositoryName,
		commitID:       inputCommitID,
	}] = true
	b.lock.Unlock()
	b.lock.RLock()
	outputCommitID, ok := b.outputRepositoryToBranchID[outputRepositoryName]
	b.lock.RUnlock()
	if ok {
		return outputCommitID, nil
	}
	parentCommitID, err := b.getParentCommitID(
		inputRepositoryName,
		inputCommitID,
	)
	if err != nil {
		return "", err
	}
	for parentCommitID != "" {
		pfsCommitMapping, err := b.storeClient.GetPfsCommitMappingLatest(
			inputRepositoryName,
			parentCommitID,
		)
		if err != nil {
			return "", err
		}
		if pfsCommitMapping != nil {
			return b.branch(
				outputRepositoryName,
				pfsCommitMapping.OutputCommitId,
			)
		}
	}
	return b.branch(
		outputRepositoryName,
		"scratch",
	)
}

func (b *brancher) CommitOutstanding() error {
	if err := b.destroyable.Destroy(); err != nil {
		return err
	}
	for repositoryName, commitID := range b.outputRepositoryToBranchID {
		if err := pfsutil.Write(
			b.pfsAPIClient,
			repositoryName,
			commitID,
		); err != nil {
			return err
		}
		for inputRepositoryCommit := range b.outputRepositoryToInputRepositories[repositoryName] {
			if err := b.storeClient.CreatePfsCommitMapping(
				&pps.PfsCommitMapping{
					InputRepository:  inputRepositoryCommit.repositoryName,
					InputCommitId:    inputRepositoryCommit.commitID,
					OutputRepository: repositoryName,
					OutputCommitId:   commitID,
					Timestamp:        prototime.TimeToTimestamp(b.timer.Now()),
				},
			); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *brancher) getParentCommitID(
	repositoryName string,
	commitID string,
) (string, error) {
	commitInfo, err := pfsutil.GetCommitInfo(
		b.pfsAPIClient,
		repositoryName,
		commitID,
	)
	if err != nil {
		return "", err
	}
	if commitInfo.ParentCommit == nil {
		return "", nil
	}
	return commitInfo.ParentCommit.Id, nil
}

func (b *brancher) branch(
	repositoryName string,
	commitID string,
) (string, error) {
	commit, err := pfsutil.Branch(
		b.pfsAPIClient,
		repositoryName,
		commitID,
	)
	if err != nil {
		return "", err
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if existingCommitID, ok := b.outputRepositoryToBranchID[repositoryName]; ok {
		// TODO(pedge) delete new branch
		return existingCommitID, nil
	}
	b.outputRepositoryToBranchID[repositoryName] = commit.Id
	return commit.Id, nil
}
