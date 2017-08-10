package worker

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/robfig/cron"

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
	pachClient := a.pachClient.WithCtx(ctx)
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

		if input.Atom != nil {
			request := &pfs.SubscribeCommitRequest{
				Repo:   client.NewRepo(input.Atom.Repo),
				Branch: input.Atom.Branch,
			}
			if input.Atom.FromCommit != "" {
				request.From = client.NewCommit(input.Atom.Repo, input.Atom.FromCommit)
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
						for _, rootInput := range rootInputs {
							if rootInput.Atom != nil {
								if len(rootCommits[rootInput.Atom.Repo]) != 1 {
									return false
								}
							}
						}
						return true
					}); set != nil {
						bs := &branchSet{
							NewBranch: i,
						}
						for i, commitInfo := range set {
							branch := "master"
							if directInputs[i].Atom != nil {
								branch = directInputs[i].Atom.Branch
							}
							bs.Branches = append(bs.Branches, &pfs.BranchInfo{
								Name: branch,
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
		if input.Cron != nil {
			schedule, err := cron.Parse(input.Cron.Spec)
			// it shouldn't be possible to error here because we validate the spec in CreatePipeline
			if err != nil {
				cancel()
				return nil, err
			}
			repo := fmt.Sprintf("%s_%s", a.pipelineInfo.Pipeline.Name, input.Cron.Name)
			var buffer bytes.Buffer
			if err := pachClient.GetFile(repo, "master", "time", 0, 0, &buffer); err != nil {
				cancel()
				return nil, err
			}
			tstamp := &types.Timestamp{}
			if err := jsonpb.UnmarshalString(buffer.String(), tstamp); err != nil {
				cancel()
				return nil, err
			}
			t, err := types.TimestampFromProto(tstamp)
			if err != nil {
				cancel()
				return nil, err
			}
			go func() {
				for {
					if err := func() error {
						nextT := schedule.Next(t)
						time.Sleep(time.Until(nextT))
						commit, err := pachClient.StartCommit(repo, "master")
						if err != nil {
							return err
						}
						timestamp, err := types.TimestampProto(nextT)
						if err != nil {
							return err
						}
						timeString, err := (&jsonpb.Marshaler{}).MarshalToString(timestamp)
						if err != nil {
							return err
						}
						if _, err := pachClient.PutFile(repo, "master", "time", strings.NewReader(timeString)); err != nil {
							return err
						}
						if err := pachClient.FinishCommit(repo, "master"); err != nil {
							return err
						}
						commitInfo, err := pachClient.InspectCommit(commit.Repo.Name, "master")
						if err != nil {
							return err
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
							for _, rootInput := range rootInputs {
								if rootInput.Atom != nil {
									if len(rootCommits[rootInput.Atom.Repo]) != 1 {
										return false
									}
								}
							}
							return true
						}); set != nil {
							bs := &branchSet{
								NewBranch: i,
							}
							for i, commitInfo := range set {
								branch := "master"
								if directInputs[i].Atom != nil {
									branch = directInputs[i].Atom.Branch
								}
								bs.Branches = append(bs.Branches, &pfs.BranchInfo{
									Name: branch,
									Head: commitInfo.Commit,
								})
							}
							select {
							case <-ctx.Done():
								commitSetsMutex.Unlock()
								return nil
							case ch <- bs:
							}
						}
						commitSetsMutex.Unlock()
						return nil
					}(); err != nil {
						select {
						case <-ctx.Done():
						case ch <- &branchSet{
							Err: err,
						}:
						}
						return
					}
				}
			}()
		}
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
func (a *APIServer) directInputs(ctx context.Context) []*pps.Input {
	repoSet := make(map[string]bool)
	var result []*pps.Input
	pps.VisitInput(a.pipelineInfo.Input, func(input *pps.Input) {
		if input.Atom != nil {
			if repoSet[input.Atom.Repo] {
				return
			}
			repoSet[input.Atom.Repo] = true
			result = append(result, input)
		}
		if input.Cron != nil {
			result = append(result, input)
		}
	})
	return result
}

// rootInputs returns the root provenance of direct inputs.
func (a *APIServer) rootInputs(ctx context.Context) ([]*pps.Input, error) {
	atomInputs := a.directInputs(ctx)
	return a._rootInputs(ctx, atomInputs)
}

func (a *APIServer) _rootInputs(ctx context.Context, inputs []*pps.Input) ([]*pps.Input, error) {
	pfsClient := a.pachClient.PfsAPIClient
	ppsClient := a.pachClient.PpsAPIClient
	// map from repo.Name + branch to *pps.AtomInput so that we don't
	// repeat ourselves.
	resultMap := make(map[string]*pps.Input)
	for _, input := range inputs {
		if input.Atom != nil {
			repoInfo, err := pfsClient.InspectRepo(ctx, &pfs.InspectRepoRequest{client.NewRepo(input.Atom.Repo)})
			if err != nil {
				return nil, err
			}
			if len(repoInfo.Provenance) == 0 {
				resultMap[input.Atom.Repo+input.Atom.Branch] = input
				continue
			}
			// if the repo has nonzero provenance we know that it's a pipeline
			pipelineInfo, err := ppsClient.InspectPipeline(ctx, &pps.InspectPipelineRequest{client.NewPipeline(input.Atom.Repo)})
			if err != nil {
				return nil, err
			}
			// TODO we should be propagating `from_commit` here
			var visitErr error
			pps.VisitInput(pipelineInfo.Input, func(visitInput *pps.Input) {
				if input.Atom != nil {
					subResults, err := a._rootInputs(ctx, []*pps.Input{visitInput})
					if err != nil && visitErr == nil {
						visitErr = err
					}
					for _, input := range subResults {
						resultMap[input.Atom.Repo+input.Atom.Branch] = input
					}
				}
			})
			if visitErr != nil {
				return nil, visitErr
			}
		}
		if input.Cron != nil {
			resultMap[input.Cron.Name] = input
		}
	}
	var result []*pps.Input
	for _, input := range resultMap {
		result = append(result, input)
	}
	return result, nil
}
