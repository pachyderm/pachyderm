package transform

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pkg/work"
)

func withWorkerSpawnerPair(pipelineInfo *pps.PipelineInfo, cb func(env *testEnv) error) error {
	// We only support simple pfs input pipelines in this test suite at the moment
	if pipelineInfo.Input == nil || pipelineInfo.Input.Pfs == nil {
		return fmt.Errorf("invalid pipeline, only a single PFS input is supported")
	}

	eg := &errgroup.Group{}

	err := withTestEnv(pipelineInfo, func(env *testEnv) error {
		// TODO: for debugging purposes, remove
		env.logger.Writer = os.Stdout

		// Set up repos and branches for the pipeline
		input := pipelineInfo.Input.Pfs
		if err := env.PachClient.CreateRepo(input.Repo); err != nil {
			return err
		}
		if err := env.PachClient.CreateBranch(input.Repo, input.Branch, "", nil); err != nil {
			return err
		}
		if err := env.PachClient.CreateBranch(pipelineInfo.SpecCommit.Repo.Name, pipelineInfo.Pipeline.Name, "", nil); err != nil {
			return err
		}
		commit, err := env.PachClient.StartCommit(pipelineInfo.SpecCommit.Repo.Name, pipelineInfo.Pipeline.Name)
		if err != nil {
			return err
		}
		pipelineInfo.SpecCommit = commit
		if err := env.PachClient.FinishCommit(pipelineInfo.SpecCommit.Repo.Name, commit.ID); err != nil {
			return err
		}
		if err := env.PachClient.CreateRepo(pipelineInfo.Pipeline.Name); err != nil {
			return err
		}
		if err := env.PachClient.CreateBranch(
			pipelineInfo.Pipeline.Name,
			pipelineInfo.OutputBranch,
			"",
			[]*pfs.Branch{
				client.NewBranch(input.Repo, input.Branch),
				client.NewBranch(pipelineInfo.SpecCommit.Repo.Name, pipelineInfo.Pipeline.Name),
			},
		); err != nil {
			return err
		}

		eg.Go(func() error {
			return Run(env.driver, env.logger)
		})

		eg.Go(func() error {
			return env.driver.NewTaskWorker().Run(
				env.driver.PachClient().Ctx(),
				func(ctx context.Context, task *work.Task, subtask *work.Task) error {
					return Worker(env.driver, env.logger, task, subtask)
				},
			)
		})

		return cb(env)
	})

	workerSpawnerErr := eg.Wait()
	// TODO: comparing against context.Canceled isn't working - something is wrapping it?
	if workerSpawnerErr != nil && workerSpawnerErr.Error() != "context canceled" {
		return workerSpawnerErr
	}
	return err
}

func triggerJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo) *pfs.Commit {
	commit, err := env.PachClient.StartCommit(pi.Input.Pfs.Repo, pi.Input.Pfs.Branch)
	require.NoError(t, err)
	require.NoError(t, env.PachClient.FinishCommit(pi.Input.Pfs.Repo, commit.ID))
	return commit
}

func mockBasicJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo) {
	// Mock out the initial ListJob, CreateJob, and InspectJob calls
	var etcdJobInfo *pps.EtcdJobInfo
	var outputCommit *pfs.Commit

	// TODO: use a 'real' pps if we can make one that doesn't need a real kube client
	env.MockPachd.PPS.ListJobStream.Use(func(*pps.ListJobRequest, pps.API_ListJobStreamServer) error {
		return nil
	})
	env.MockPachd.PPS.CreateJob.Use(func(ctx context.Context, request *pps.CreateJobRequest) (*pps.Job, error) {
		etcdJobInfo = &pps.EtcdJobInfo{
			Job:           client.NewJob(uuid.NewWithoutDashes()),
			OutputCommit:  request.OutputCommit,
			Pipeline:      request.Pipeline,
			Stats:         request.Stats,
			Restart:       request.Restart,
			DataProcessed: request.DataProcessed,
			DataSkipped:   request.DataSkipped,
			DataTotal:     request.DataTotal,
			DataFailed:    request.DataFailed,
			DataRecovered: request.DataRecovered,
			StatsCommit:   request.StatsCommit,
			Started:       request.Started,
			Finished:      request.Finished,
		}
		return etcdJobInfo.Job, nil
	})
	env.MockPachd.PPS.InspectJob.Use(func(ctx context.Context, request *pps.InspectJobRequest) (*pps.JobInfo, error) {
		outputCommitInfo, err := env.PachClient.InspectCommit(outputCommit.Repo.Name, outputCommit.ID)
		if err != nil {
			return nil, err
		}
		return &pps.JobInfo{
			Job:              etcdJobInfo.Job,
			Pipeline:         etcdJobInfo.Pipeline,
			OutputRepo:       &pfs.Repo{Name: etcdJobInfo.Pipeline.Name},
			OutputCommit:     etcdJobInfo.OutputCommit,
			Restart:          etcdJobInfo.Restart,
			DataProcessed:    etcdJobInfo.DataProcessed,
			DataSkipped:      etcdJobInfo.DataSkipped,
			DataTotal:        etcdJobInfo.DataTotal,
			DataFailed:       etcdJobInfo.DataFailed,
			DataRecovered:    etcdJobInfo.DataRecovered,
			Stats:            etcdJobInfo.Stats,
			StatsCommit:      etcdJobInfo.StatsCommit,
			State:            etcdJobInfo.State,
			Reason:           etcdJobInfo.Reason,
			Started:          etcdJobInfo.Started,
			Finished:         etcdJobInfo.Finished,
			Transform:        pi.Transform,
			PipelineVersion:  pi.Version,
			ParallelismSpec:  pi.ParallelismSpec,
			Egress:           pi.Egress,
			Service:          pi.Service,
			Spout:            pi.Spout,
			OutputBranch:     pi.OutputBranch,
			ResourceRequests: pi.ResourceRequests,
			ResourceLimits:   pi.ResourceLimits,
			Input:            ppsutil.JobInput(pi, outputCommitInfo),
			EnableStats:      pi.EnableStats,
			Salt:             pi.Salt,
			ChunkSpec:        pi.ChunkSpec,
			DatumTimeout:     pi.DatumTimeout,
			JobTimeout:       pi.JobTimeout,
			DatumTries:       pi.DatumTries,
			SchedulingSpec:   pi.SchedulingSpec,
			PodSpec:          pi.PodSpec,
			PodPatch:         pi.PodPatch,
		}, nil
	})

	env.MockPachd.PPS.UpdateJobState.Use(func(ctx context.Context, request *pps.UpdateJobStateRequest) (*types.Empty, error) {
		etcdJobInfo.State = request.State
		etcdJobInfo.Reason = request.Reason
		return &types.Empty{}, nil
	})

	// TODO: this may race - the InspectJob call might run before this returns
	outputCommit = triggerJob(t, env, pi)
}

func TestJob(t *testing.T) {
	pi := defaultPipelineInfo()
	t.Parallel()
	err := withWorkerSpawnerPair(pi, func(env *testEnv) error {
		mockBasicJob(t, env, pi)
		time.Sleep(time.Second * 4)
		return nil
	})
	require.NoError(t, err)
}
