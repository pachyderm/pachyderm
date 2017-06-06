package worker

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.pedge.io/lion/proto"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/dlock"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
)

const (
	masterLockPath = "_master_worker_lock"
)

func (a *APIServer) master() {
	backoff.RetryNotify(func() error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		masterLock, err := dlock.NewDLock(ctx, a.etcdClient, path.Join(a.etcdPrefix, masterLockPath, a.pipelineInfo.Pipeline.Name))
		if err != nil {
			return err
		}
		defer masterLock.Unlock()
		ctx = masterLock.Context()

		jobInfos, err := a.pachClient.PpsAPIClient.ListJob(ctx, &pps.ListJobRequest{
			Pipeline: a.pipelineInfo.Pipeline,
		})
		if err != nil {
			return err
		}

		jobCh := make(chan *pps.JobInfo)
		go a.datumFeeder(jobCh)

		for _, jobInfo := range jobInfos.JobInfo {
			switch jobInfo.State {
			case pps.JobState_JOB_STARTING, pps.JobState_JOB_RUNNING:
				jobCh <- jobInfo
			}
		}

		bsf, err := newBranchSetFactory(ctx, a.pachClient.PfsAPIClient, a.pipelineInfo.Input)
		if err != nil {
			return err
		}
		defer bsf.Close()
	nextInput:
		for {
			bs := <-bsf.Chan()
			if bs.Err != nil {
				return err
			}

			// (create JobInput for new processing job)
			jobInput := proto.Clone(a.pipelineInfo.Input).(*pps.Input)
			var visitErr error
			pps.VisitInput(jobInput, func(input *pps.Input) {
				if input.Atom != nil {
					for _, branch := range bs.Branches {
						if input.Atom.Repo == branch.Head.Repo.Name && input.Atom.Branch == branch.Name {
							input.Atom.Commit = branch.Head.ID
						}
					}
					if input.Atom.Commit == "" {
						visitErr = fmt.Errorf("didn't find input commit for %s/%s", input.Atom.Repo, input.Atom.Branch)
					}
					input.Atom.FromCommit = ""
				}
			})
			if visitErr != nil {
				return visitErr
			}

			jobsRO := a.jobs.ReadOnly(ctx)
			// Check if this input set has already been processed
			jobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputIndex, jobInput)
			if err != nil {
				return err
			}

			// This is doing the same thing as the line above but for 1.4.5 style jobs
			oldJobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputsIndex, untranslateJobInputs(jobInput))
			if err != nil {
				return err
			}

			// Check if any of the jobs in jobIter have been run already. If so, skip
			// this input.
			for {
				var jobID string
				var jobInfo pps.JobInfo
				ok, err := jobIter.Next(&jobID, &jobInfo)
				if err != nil {
					return err
				}
				if !ok {
					ok, err := oldJobIter.Next(&jobID, &jobInfo)
					if err != nil {
						return err
					}
					if !ok {
						break
					}
				}
				if jobInfo.PipelineID == a.pipelineInfo.ID && jobInfo.PipelineVersion == a.pipelineInfo.Version {
					// TODO(derek): we should check if the output commit exists.  If the
					// output commit has been deleted, we should re-run the job.
					continue nextInput
				}
			}

			// now we need to find the parentJob for this job. The parent job
			// is defined as the job with the same input commits except for the
			// newest input commit which triggered this job. In place of it we
			// use that commit's parent.
			newBranch := bs.Branches[bs.NewBranch]
			newCommitInfo, err := a.pachClient.PfsAPIClient.InspectCommit(ctx, &pfs.InspectCommitRequest{
				Commit: newBranch.Head,
			})
			if err != nil {
				return err
			}
			var parentJob *pps.Job
			if newCommitInfo.ParentCommit != nil {
				// recreate our parent's inputs so we can look it up in etcd
				parentJobInput := proto.Clone(jobInput).(*pps.Input)
				pps.VisitInput(parentJobInput, func(input *pps.Input) {
					if input.Atom != nil && input.Atom.Repo == newBranch.Head.Repo.Name && input.Atom.Branch == newBranch.Name {
						input.Atom.Commit = newCommitInfo.ParentCommit.ID
					}
				})
				if visitErr != nil {
					return visitErr
				}
				jobIter, err := jobsRO.GetByIndex(ppsdb.JobsInputIndex, parentJobInput)
				if err != nil {
					return err
				}
				for {
					var jobID string
					var jobInfo pps.JobInfo
					ok, err := jobIter.Next(&jobID, &jobInfo)
					if err != nil {
						return err
					}
					if !ok {
						break
					}
					if jobInfo.PipelineID == a.pipelineInfo.ID && jobInfo.PipelineVersion == a.pipelineInfo.Version {
						parentJob = jobInfo.Job
					}
				}
			}

			job, err := a.pachClient.PpsAPIClient.CreateJob(ctx, &pps.CreateJobRequest{
				Pipeline: a.pipelineInfo.Pipeline,
				Input:    jobInput,
				// TODO(derek): Note that once the pipeline restarts, the `job`
				// variable is lost and we don't know who is our parent job.
				ParentJob: parentJob,
				NewBranch: newBranch,
			})
			if err != nil {
				return err
			}

			jobInfo, err := a.pachClient.PpsAPIClient.InspectJob(ctx, &pps.InspectJobRequest{
				Job: job,
			})
			if err != nil {
				return err
			}

			jobCh <- jobInfo
		}

		return nil
	}, backoff.NewInfiniteBackOff(), func(err error, d time.Duration) error {
		protolion.Errorf("master: error running the master process: %v; retrying in %v", err, d)
		return nil
	})
}

func (a *APIServer) datumFeeder(jobCh chan *pps.JobInfo) {

}

func untranslateJobInputs(input *pps.Input) []*pps.JobInput {
	var result []*pps.JobInput
	if input.Cross != nil {
		for _, input := range input.Cross {
			if input.Atom == nil {
				return nil
			}
			result = append(result, &pps.JobInput{
				Name:   input.Atom.Name,
				Commit: client.NewCommit(input.Atom.Repo, input.Atom.Commit),
				Glob:   input.Atom.Glob,
				Lazy:   input.Atom.Lazy,
			})
		}
	}
	return result
}
