package server

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type branchSetFactory interface {
	Next() ([]*pfs.Branch, error)
}

type branchSetFactoryImpl struct {
	branchCh    chan *pfs.Branch
	errCh       chan error
	branchSet   []*pfs.Branch
	numBranches int // number of unique branches
}

func (c *branchSetFactoryImpl) Next() ([]*pfs.Branch, error) {
	for {
		var newBranch *pfs.Branch
		select {
		case newBranch = <-c.branchCh:
		case err := <-c.errCh:
			return nil, err
		}

		var found bool
		for i, branch := range c.branchSet {
			if branch.Head.Repo.Name == newBranch.Head.Repo.Name && branch.Name == newBranch.Name {
				c.branchSet[i] = newBranch
				found = true
			}
		}
		if !found {
			c.branchSet = append(c.branchSet, newBranch)
		}
		if len(c.branchSet) == c.numBranches {
			newBranchSet := make([]*pfs.Branch, len(c.branchSet))
			copy(newBranchSet, c.branchSet)
			return newBranchSet, nil
		}
	}
	panic("unreachable")
}

func newBranchSetFactory(ctx context.Context, pfsClient pfs.APIClient, inputs []*pps.PipelineInput) (branchSetFactory, error) {
	branchCh := make(chan *pfs.Branch)
	errCh := make(chan error)

	f := &branchSetFactoryImpl{
		branchCh: branchCh,
		errCh:    errCh,
	}

	uniqueBranches := make(map[string]map[string]bool)
	for _, input := range inputs {
		if uniqueBranches[input.Repo.Name] == nil {
			uniqueBranches[input.Repo.Name] = make(map[string]bool)
			uniqueBranches[input.Repo.Name][input.Branch] = true
		}
	}

	for repoName, branches := range uniqueBranches {
		for branchName := range branches {
			f.numBranches++
			stream, err := pfsClient.SubscribeCommit(ctx, &pfs.SubscribeCommitRequest{
				Repo:   &pfs.Repo{repoName},
				Branch: branchName,
			})
			if err != nil {
				return nil, err
			}
			go func(branchName string) {
				commitInfo, err := stream.Recv()
				if err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
				branchCh <- &pfs.Branch{
					Name: branchName,
					Head: commitInfo.Commit,
				}
			}(branchName)
		}
	}

	return f, nil
}
