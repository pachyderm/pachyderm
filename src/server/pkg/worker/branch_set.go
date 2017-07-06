package worker

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	"golang.org/x/net/context"
)

type branchSet struct {
	Branches  []*pfs.BranchInfo
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

func (a *APIServer) newBranchSetFactory(_ctx context.Context) (branchSetFactory, error) {
	ctx, cancel := context.WithCancel(_ctx)
	pfsClient := a.pachClient.PfsAPIClient

	rootInputs, err := a.rootInputs(ctx)
	if err != nil {
		return nil, err
	}
	directInputs := a.directInputs(ctx)

	commitSets := make([][]*pfs.CommitInfo, len(directInputs))
	var commitSetsMutex sync.Mutex

	ch := make(chan *branchSet)
	for i, input := range directInputs {
		i, input := i, input

		request := &pfs.SubscribeCommitRequest{
			Repo:   client.NewRepo(input.Repo),
			Branch: input.Branch,
		}
		if input.FromCommit != "" {
			request.From = client.NewCommit(input.Repo, input.FromCommit)
		}
		stream, err := pfsClient.SubscribeCommit(ctx, request)
		if err != nil {
			cancel()
			return nil, err
		}

		go func() {
			for {
				commitInfo, err := stream.Recv()
				if err != nil {
					select {
					case <-ctx.Done():
					case ch <- &branchSet{
						Err: err,
					}:
					}
					return
				}

				commitSetsMutex.Lock()
				commitSets[i] = append(commitSets[i], commitInfo)

				// Now we look for a set of commits that have provenance
				// commits from the root input repos.  The set is only valid
				// if there's precisely one provenance commit from each root
				// input repo.
				if set := findCommitSet(commitSets, i, func(set []*pfs.CommitInfo) bool {
					rootCommits := make(map[string]map[string]bool)
					setRootCommit := func(commit *pfs.Commit) {
						if rootCommits[commit.Repo.Name] == nil {
							rootCommits[commit.Repo.Name] = make(map[string]bool)
						}
						rootCommits[commit.Repo.Name][commit.ID] = true
					}
					for _, commitInfo := range set {
						setRootCommit(commitInfo.Commit)
						for _, provCommit := range commitInfo.Provenance {
							setRootCommit(provCommit)
						}
					}
					// ok tells us if it's ok to spawn a job with this
					// commit set.
					for _, rootInput := range rootInputs {
						if len(rootCommits[rootInput.Repo]) != 1 {
							return false
						}
					}
					return true
				}); set != nil {
					bs := &branchSet{
						NewBranch: i,
					}
					for i, commitInfo := range set {
						bs.Branches = append(bs.Branches, &pfs.BranchInfo{
							Name: directInputs[i].Branch,
							Head: commitInfo.Commit,
						})
					}
					select {
					case <-ctx.Done():
						commitSetsMutex.Unlock()
						return
					case ch <- bs:
					}
				}
				commitSetsMutex.Unlock()
			}
		}()
	}

	f := &branchSetFactoryImpl{
		cancel: cancel,
		ch:     ch,
	}

	return f, nil
}

// findCommitSet runs a function on every commit set, starting with the
// most recent one.  If the function returns true, then findCommitSet
// removes the commit sets that are older than the current one and returns
// the current one.
// i is the index of the input from which we just got a new commit.  Since
// it's the "triggering" input, we know that we only have to consider commit
// sets that include the triggering commit since other commit sets must have
// already been considered in previous runs.
func findCommitSet(commitSets [][]*pfs.CommitInfo, i int, f func(commitSet []*pfs.CommitInfo) bool) []*pfs.CommitInfo {
	numCommitSets := 1
	for j, commits := range commitSets {
		if i != j {
			numCommitSets *= len(commits)
		}
	}
	for j := numCommitSets - 1; j >= 0; j-- {
		numCommitSets := j
		var commitSet []*pfs.CommitInfo
		var indexes []int
		for k, commits := range commitSets {
			var index int
			if k == i {
				index = len(commits) - 1
			} else {
				index = numCommitSets % len(commits)
				numCommitSets /= len(commits)
			}
			indexes = append(indexes, index)
			commitSet = append(commitSet, commits[index])
		}
		if f(commitSet) {
			// Remove older commit sets
			for k, index := range indexes {
				commitSets[k] = commitSets[k][index:]
			}
			return commitSet
		}
	}
	return nil
}

// directInputs returns the inputs that should trigger the pipeline.  Inputs
// from the same repo/branch are de-duplicated.
func (a *APIServer) directInputs(ctx context.Context) []*pps.AtomInput {
	repoSet := make(map[string]bool)
	var atomInputs []*pps.AtomInput
	pps.VisitInput(a.pipelineInfo.Input, func(input *pps.Input) {
		if input.Atom != nil {
			if repoSet[input.Atom.Repo] {
				return
			}
			repoSet[input.Atom.Repo] = true
			atomInputs = append(atomInputs, input.Atom)
		}
	})
	return atomInputs
}

// rootInputs returns the root provenance of direct inputs.
func (a *APIServer) rootInputs(ctx context.Context) ([]*pps.AtomInput, error) {
	atomInputs := a.directInputs(ctx)
	return a._rootInputs(ctx, atomInputs)
}

func (a *APIServer) _rootInputs(ctx context.Context, atomInputs []*pps.AtomInput) ([]*pps.AtomInput, error) {
	pfsClient := a.pachClient.PfsAPIClient
	ppsClient := a.pachClient.PpsAPIClient
	// map from repo.Name + branch to *pps.AtomInput so that we don't
	// repeat ourselves.
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
		pipelineInfo, err := ppsClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{client.NewPipeline(atomInput.Repo)})
		if err != nil {
			return nil, err
		}
		// TODO we should be propagating `from_commit` here
		var visitErr error
		pps.VisitInput(pipelineInfo.Input, func(input *pps.Input) {
			if input.Atom != nil {
				subResults, err := a._rootInputs(ctx, []*pps.AtomInput{input.Atom})
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
