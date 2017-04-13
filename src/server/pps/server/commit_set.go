package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type branchSet struct {
	Branches []*pfs.Branch
	Err      error
}

type branchSetFactory interface {
	Chan() chan *branchSet
	Close()
}

type branchSetFactoryImpl struct {
	ch     chan *branchSet
	cancel context.CancelFunc
}

func (f *branchSetFactoryImpl) Close() {
	f.cancel()
}

func (f *branchSetFactoryImpl) Chan() chan *branchSet {
	return f.ch
}

func newBranchSetFactory(_ctx context.Context, pfsClient pfs.APIClient, inputs []*pps.PipelineInput) (branchSetFactory, error) {
	ctx, cancel := context.WithCancel(_ctx)

	uniqueBranches := make(map[string]map[string]*pfs.Commit)
	for _, input := range inputs {
		if uniqueBranches[input.Repo.Name] == nil {
			uniqueBranches[input.Repo.Name] = make(map[string]*pfs.Commit)
		}
		uniqueBranches[input.Repo.Name][input.Branch] = input.From
	}

	var numBranches int
	branchCh := make(chan *pfs.Branch)
	errCh := make(chan error)
	for repoName, branches := range uniqueBranches {
		for branchName, fromCommit := range branches {
			numBranches++
			stream, err := pfsClient.SubscribeCommit(ctx, &pfs.SubscribeCommitRequest{
				Repo:   &pfs.Repo{repoName},
				Branch: branchName,
				From:   fromCommit,
			})
			if err != nil {
				return nil, err
			}
			go func(branchName string) {
				for {
					commitInfo, err := stream.Recv()
					if err != nil {
						select {
						case <-ctx.Done():
						case errCh <- err:
						}
						return
					}
					select {
					case <-ctx.Done():
						return
					case branchCh <- &pfs.Branch{
						Name: branchName,
						Head: commitInfo.Commit,
					}:
					}
				}
			}(branchName)
		}
	}

	ch := make(chan *branchSet)
	go func() {
		var currentBranchSet []*pfs.Branch
		for {
			var newBranch *pfs.Branch
			select {
			case <-ctx.Done():
				return
			case newBranch = <-branchCh:
			case err := <-errCh:
				select {
				case <-ctx.Done():
					return
				case ch <- &branchSet{
					Err: err,
				}:
				}
			}

			var found bool
			for i, branch := range currentBranchSet {
				if branch.Head.Repo.Name == newBranch.Head.Repo.Name && branch.Name == newBranch.Name {
					currentBranchSet[i] = newBranch
					found = true
				}
			}
			if !found {
				currentBranchSet = append(currentBranchSet, newBranch)
			}
			if len(currentBranchSet) == numBranches {
				newBranchSet := make([]*pfs.Branch, numBranches)
				copy(newBranchSet, currentBranchSet)
				select {
				case <-ctx.Done():
					return
				case ch <- &branchSet{
					Branches: newBranchSet,
				}:
				}
			}
		}
		panic("unreachable")
	}()

	f := &branchSetFactoryImpl{
		cancel: cancel,
		ch:     ch,
	}

	return f, nil
}
