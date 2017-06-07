package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type branchSet struct {
	Branches  []*pfs.Branch
	NewBranch int //newBranch indicates which branch is new
	Err       error
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

func newBranchSetFactory(_ctx context.Context, pfsClient pfs.APIClient, input *pps.Input) (branchSetFactory, error) {
	ctx, cancel := context.WithCancel(_ctx)

	uniqueBranches := make(map[string]map[string]*pfs.Commit)
	visit(input, func(input *pps.Input) {
		if input.Atom != nil {
			if uniqueBranches[input.Atom.Repo] == nil {
				uniqueBranches[input.Atom.Repo] = make(map[string]*pfs.Commit)
			}
			if input.Atom.FromCommit != "" {
				uniqueBranches[input.Atom.Repo][input.Atom.Branch] =
					client.NewCommit(input.Atom.Repo, input.Atom.FromCommit)
			} else {
				uniqueBranches[input.Atom.Repo][input.Atom.Branch] = nil
			}
		}
	})

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
				return
			}

			var found bool
			var newBranchIndex int
			for i, branch := range currentBranchSet {
				if branch.Head.Repo.Name == newBranch.Head.Repo.Name && branch.Name == newBranch.Name {
					currentBranchSet[i] = newBranch
					found = true
					newBranchIndex = i
				}
			}
			if !found {
				currentBranchSet = append(currentBranchSet, newBranch)
				newBranchIndex = len(currentBranchSet) - 1
			}
			if len(currentBranchSet) == numBranches {
				newBranchSet := make([]*pfs.Branch, numBranches)
				copy(newBranchSet, currentBranchSet)
				select {
				case <-ctx.Done():
					return
				case ch <- &branchSet{
					Branches:  newBranchSet,
					NewBranch: newBranchIndex,
				}:
				}
			}
		}
	}()

	f := &branchSetFactoryImpl{
		cancel: cancel,
		ch:     ch,
	}

	return f, nil
}
