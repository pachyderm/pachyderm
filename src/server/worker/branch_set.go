package worker

import (
	"bytes"
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
	ch              chan *branchSet
	cancel          context.CancelFunc
	rootInputs      []*pps.Input
	directInputs    []*pps.Input
	commitSets      [][]*pfs.CommitInfo
	commitSetsMutex sync.Mutex
}

func (f *branchSetFactoryImpl) Close() {
	f.cancel()
}

func (f *branchSetFactoryImpl) Chan() chan *branchSet {
	return f.ch
}

func (a *APIServer) newBranchSetFactory(_ctx context.Context) (_ branchSetFactory, retErr error) {
	ctx, cancel := context.WithCancel(_ctx)
	defer func() {
		// We call cancel if the there's an error, if there's not then cancel
		// will be returned as part of the branchSetFactory
		if retErr != nil {
			cancel()
		}
	}()
	pachClient := a.pachClient.WithCtx(ctx)

	result := &branchSetFactoryImpl{
		cancel: cancel,
		ch:     make(chan *branchSet),
	}
	var err error
	result.rootInputs, err = a.rootInputs(pachClient)
	if err != nil {
		return nil, err
	}
	result.directInputs = a.directInputs()
	result.commitSets = make([][]*pfs.CommitInfo, len(result.directInputs))

	for i, input := range result.directInputs {
		i, input := i, input

		subscribe := func(repo string, branch string, fromCommit string) error {
			iter, err := pachClient.SubscribeCommit(repo, branch, fromCommit)
			if err != nil {
				return err
			}
			go func() {
				for {
					commitInfo, err := iter.Next()
					if err != nil {
						select {
						case <-ctx.Done():
							return
						case result.ch <- &branchSet{
							Err: err,
						}:
						}
						return
					}
					result.sendBranchSet(ctx, i, commitInfo)
				}
			}()
			return nil
		}
		if input.Atom != nil {
			err = subscribe(input.Atom.Repo, input.Atom.Branch, input.Atom.FromCommit)
			if err != nil {
				return nil, err
			}
		}
		if input.Git != nil {
			err = subscribe(input.Git.Name, input.Git.Branch, "")
			if err != nil {
				return nil, err
			}
		}
		if input.Cron != nil {
			schedule, err := cron.Parse(input.Cron.Spec)
			// it shouldn't be possible to error here because we validate the spec in CreatePipeline
			if err != nil {
				return nil, err
			}
			tstamp := &types.Timestamp{}
			var buffer bytes.Buffer
			if err := pachClient.GetFile(input.Cron.Repo, "master", "time", 0, 0, &buffer); err != nil && !isNotFoundErr(err) {
				return nil, err
			} else if err != nil {
				// File not found, this happens the first time the pipeline is run
				tstamp = input.Cron.Start
			} else {
				if err := jsonpb.UnmarshalString(buffer.String(), tstamp); err != nil {
					return nil, err
				}
			}
			t, err := types.TimestampFromProto(tstamp)
			if err != nil {
				return nil, err
			}
			go func() {
				for {
					if err := func() error {
						nextT := schedule.Next(t)
						t = nextT
						time.Sleep(time.Until(nextT))
						commit, err := pachClient.StartCommit(input.Cron.Repo, "master")
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
						if err := pachClient.DeleteFile(input.Cron.Repo, "master", "time"); err != nil {
							return err
						}
						if _, err := pachClient.PutFile(input.Cron.Repo, "master", "time", strings.NewReader(timeString)); err != nil {
							return err
						}
						if err := pachClient.FinishCommit(input.Cron.Repo, "master"); err != nil {
							return err
						}
						commitInfo, err := pachClient.InspectCommit(commit.Repo.Name, "master")
						if err != nil {
							return err
						}
						result.sendBranchSet(ctx, i, commitInfo)
						return nil
					}(); err != nil {
						select {
						case <-ctx.Done():
							return
						case result.ch <- &branchSet{
							Err: err,
						}:
						}
						return
					}
				}
			}()
		}
	}

	return result, nil
}

func (f *branchSetFactoryImpl) sendBranchSet(ctx context.Context, i int, commitInfo *pfs.CommitInfo) {
	f.commitSetsMutex.Lock()
	defer f.commitSetsMutex.Unlock()
	f.commitSets[i] = append(f.commitSets[i], commitInfo)
	// Now we look for a set of commits that have provenance
	// commits from the root input repos.  The set is only valid
	// if there's precisely one provenance commit from each root
	// input repo.
	if set := findCommitSet(f.commitSets, i, func(set []*pfs.CommitInfo) bool {
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
		for _, rootInput := range f.rootInputs {
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
			if f.directInputs[i].Atom != nil {
				branch = f.directInputs[i].Atom.Branch
			}
			bs.Branches = append(bs.Branches, &pfs.BranchInfo{
				Name: branch,
				Head: commitInfo.Commit,
			})
		}
		select {
		case <-ctx.Done():
		case f.ch <- bs:
		}
	}
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
func (a *APIServer) directInputs() []*pps.Input {
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
		if input.Git != nil {
			result = append(result, input)
		}
	})
	return result
}

// rootInputs returns the root provenance of direct inputs.
func (a *APIServer) rootInputs(c *client.APIClient) ([]*pps.Input, error) {
	atomInputs := a.directInputs()
	return a._rootInputs(c, atomInputs)
}

func (a *APIServer) _rootInputs(c *client.APIClient, inputs []*pps.Input) ([]*pps.Input, error) {
	// map from repo.Name + branch to *pps.AtomInput so that we don't
	// repeat ourselves.
	resultMap := make(map[string]*pps.Input)
	for _, input := range inputs {
		if input.Atom != nil {
			repoInfo, err := c.InspectRepo(input.Atom.Repo)
			if err != nil {
				return nil, err
			}
			if len(repoInfo.Provenance) == 0 {
				resultMap[input.Atom.Repo+input.Atom.Branch] = input
				continue
			}
			// if the repo has nonzero provenance we know that it's a pipeline
			pipelineInfo, err := c.InspectPipeline(input.Atom.Repo)
			if err != nil {
				return nil, err
			}
			// TODO we should be propagating `from_commit` here
			var visitErr error
			pps.VisitInput(pipelineInfo.Input, func(visitInput *pps.Input) {
				if input.Atom != nil {
					subResults, err := a._rootInputs(c, []*pps.Input{visitInput})
					if err != nil && visitErr == nil {
						visitErr = err
					}
					for _, input := range subResults {
						if input.Atom != nil {
							resultMap[input.Atom.Repo+input.Atom.Branch] = input
						} else {
							resultMap[input.Cron.Name] = input
						}
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
		if input.Git != nil {
			// Sanity check that the repo exists
			_, err := c.InspectRepo(input.Git.Name)
			if err != nil {
				return nil, err
			}
			resultMap[input.Git.Name] = input
		}
	}
	var result []*pps.Input
	for _, input := range resultMap {
		result = append(result, input)
	}
	return result, nil
}
