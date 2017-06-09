package server

import (
	"io"

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

func (a *apiServer) newBranchSetFactory(_ctx context.Context, input *pps.Input) (branchSetFactory, error) {
	ctx, cancel := context.WithCancel(_ctx)

	pfsClient, err := a.getPFSClient()
	if err != nil {
		return nil, err
	}
	ppsClient, err := a.getPPSClient()
	if err != nil {
		return nil, err
	}

	var atomInputs []*pps.AtomInput
	repoMap := make(map[string]*pps.AtomInput)
	visit(input, func(input *pps.Input) {
		if input.Atom != nil {
			atomInputs = append(atomInputs, input.Atom)
			repoMap[input.Atom.Repo] = input.Atom
		}
	})

	var repos []*pfs.Repo
	for repo := range repoMap {
		repos = append(repos, client.NewRepo(repo))
	}

	rawInputs, err := a.rawInputs(ctx, atomInputs)
	if err != nil {
		return nil, err
	}

	uniqueBranches := make(map[string]map[string]*pfs.Commit)
	for _, atom := range rawInputs {
		if uniqueBranches[atom.Repo] == nil {
			uniqueBranches[atom.Repo] = make(map[string]*pfs.Commit)
		}
		if input.Atom.FromCommit != "" {
			uniqueBranches[input.Atom.Repo][input.Atom.Branch] =
				client.NewCommit(input.Atom.Repo, input.Atom.FromCommit)
		} else {
			uniqueBranches[input.Atom.Repo][input.Atom.Branch] = nil
		}
	}

	branchCh := make(chan *pfs.Branch)
	errCh := make(chan error)
	for repoName, branches := range uniqueBranches {
		for branchName, fromCommit := range branches {
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
		handleErr := func(err error) {
			select {
			case <-ctx.Done():
				return
			case ch <- &branchSet{Err: err}:
			}
		}
		var currentBranchSet []*pfs.Branch
		for {
			var newBranch *pfs.Branch
			select {
			case <-ctx.Done():
				return
			case newBranch = <-branchCh:
			case err := <-errCh:
				handleErr(err)
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
			if len(currentBranchSet) == len(uniqueBranches) {
				var commits []*pfs.Commit
				for _, branch := range currentBranchSet {
					commits = append(commits, branch.Head)
				}
				flushClient, err := ppsClient.FlushCommit(ctx, &pps.FlushCommitRequest{
					Commits: commits,
					ToRepos: repos,
				})
				if err != nil {
					handleErr(err)
					return
				}
				var newBranchSet []*pfs.Branch
				for {
					jobInfo, err := flushClient.Recv()
					if err != nil {
						if err == io.EOF {
							break
						}
						handleErr(err)
						return
					}
					if atomInput := repoMap[jobInfo.OutputCommit.Repo.Name]; atomInput != nil {
						newBranchSet = append(newBranchSet, &pfs.Branch{
							Name: atomInput.Branch,
							Head: jobInfo.OutputCommit,
						})
					}
				}
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

func (a *apiServer) rawInputs(ctx context.Context, atomInputs []*pps.AtomInput) ([]*pps.AtomInput, error) {
	pfsClient, err := a.getPFSClient()
	if err != nil {
		return nil, err
	}
	// map from repo.Name + branch to *pps.AtomInput so that we don't repeat ourselves
	resultMap := make(map[string]*pps.AtomInput)
	for _, atomInput := range atomInputs {
		repoInfo, err := pfsClient.InspectRepo(ctx, &pfs.InspectRepoRequest{client.NewRepo(atomInput.Repo)})
		if err != nil {
			return nil, err
		}
		if len(repoInfo.Provenance) == 0 {
			resultMap[atomInput.Repo+atomInput.Branch] = atomInput
			continue
		}
		// if the repo has nonzero provenance we know that it's a pipeline
		pipelineInfo, err := a.InspectPipeline(ctx, &pps.InspectPipelineRequest{client.NewPipeline(atomInput.Repo)})
		if err != nil {
			return nil, err
		}
		// TODO we should be propagating `from_commit` here
		var visitErr error
		visit(pipelineInfo.Input, func(input *pps.Input) {
			if input.Atom != nil {
				subResults, err := a.rawInputs(ctx, []*pps.AtomInput{input.Atom})
				if err != nil && visitErr == nil {
					visitErr = err
				}
				for _, atomInput := range subResults {
					resultMap[atomInput.Repo+atomInput.Branch] = atomInput
				}
			}
		})
		if visitErr != nil {
			return nil, visitErr
		}
	}
	var result []*pps.AtomInput
	for _, atomInput := range resultMap {
		result = append(result, atomInput)
	}
	return result, nil
}
