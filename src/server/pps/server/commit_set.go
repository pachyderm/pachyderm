package server

import (
	"github.com/pachyderm/pachyderm/src/client"
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

func newBranchSetFactory(ctx context.Context, pfsClient pfs.APIClient, input *pps.Input) (branchSetFactory, error) {
	branchCh := make(chan *pfs.Branch)
	errCh := make(chan error)

	f := &branchSetFactoryImpl{
		branchCh: branchCh,
		errCh:    errCh,
	}

	uniqueBranches := make(map[string]map[string]*pfs.Commit)
	visit(input, func(input *pps.Input) {
		if input.Atom != nil {
			if uniqueBranches[input.Atom.Commit.Repo.Name] == nil {
				uniqueBranches[input.Atom.Commit.Repo.Name] = make(map[string]*pfs.Commit)
			}
			uniqueBranches[input.Atom.Commit.Repo.Name][input.Atom.Commit.ID] =
				client.NewCommit(input.Atom.Commit.Repo.Name, input.Atom.FromCommitID)
		}
	})

	for repoName, branches := range uniqueBranches {
		for branchName, fromCommit := range branches {
			f.numBranches++
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
						case errCh <- err:
						default:
						}
						return
					}
					branchCh <- &pfs.Branch{
						Name: branchName,
						Head: commitInfo.Commit,
					}
				}
			}(branchName)
		}
	}

	return f, nil
}
