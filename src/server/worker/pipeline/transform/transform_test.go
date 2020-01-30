package transform

import (
	"context"
	"fmt"
	"os"
	"strings"
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

func isCanceledError(err error) bool {
	// TODO: comparing against context.Canceled isn't working - something is wrapping it?
	return err != nil && err.Error() == "context canceled"
}

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
			err := Run(env.driver, env.logger)
			if isCanceledError(err) {
				return nil
			}
			return err
		})

		eg.Go(func() error {
			err := env.driver.NewTaskWorker().Run(
				env.driver.PachClient().Ctx(),
				func(ctx context.Context, task *work.Task, subtask *work.Task) error {
					return Worker(env.driver, env.logger, task, subtask)
				},
			)
			if isCanceledError(err) {
				return nil
			}
			return err
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

func triggerJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo) {
	_, err := env.PachClient.PutFile(pi.Input.Pfs.Repo, "master", "file", strings.NewReader("foobar"))
	require.NoError(t, err)

	inputCommitInfo, err := env.PachClient.InspectCommit(pi.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	require.Equal(t, int64(1), inputCommitInfo.SubvenantCommitsTotal)
	require.Equal(t, inputCommitInfo.Subvenance[0].Lower, inputCommitInfo.Subvenance[0].Upper)

	outputCommit := inputCommitInfo.Subvenance[0].Lower
	require.Equal(t, pi.Pipeline.Name, outputCommit.Repo.Name)

	files, err := env.PachClient.ListFile(pi.Input.Pfs.Repo, "master", "")
	require.NoError(t, err)
	require.Equal(t, 1, len(files))

	fmt.Printf("trigger job input files: %v\n", files)
}

func withTimeout(ctx context.Context, duration time.Duration) context.Context {
	// Create a context that the caller can wait on
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(duration):
			fmt.Printf("Canceling test after timeout")
			cancel()
		}
	}()

	return ctx
}

func mockBasicJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo) (context.Context, *pps.EtcdJobInfo) {
	// Create a context that the caller can wait on
	ctx, cancel := context.WithCancel(env.PachClient.Ctx())

	// Mock out the initial ListJob, CreateJob, and InspectJob calls
	etcdJobInfo := &pps.EtcdJobInfo{Job: client.NewJob(uuid.NewWithoutDashes())}

	// TODO: use a 'real' pps if we can make one that doesn't need a real kube client
	env.MockPachd.PPS.ListJobStream.Use(func(*pps.ListJobRequest, pps.API_ListJobStreamServer) error {
		return nil
	})

	env.MockPachd.PPS.CreateJob.Use(func(ctx context.Context, request *pps.CreateJobRequest) (*pps.Job, error) {
		etcdJobInfo.OutputCommit = request.OutputCommit
		etcdJobInfo.Pipeline = request.Pipeline
		etcdJobInfo.Stats = request.Stats
		etcdJobInfo.Restart = request.Restart
		etcdJobInfo.DataProcessed = request.DataProcessed
		etcdJobInfo.DataSkipped = request.DataSkipped
		etcdJobInfo.DataTotal = request.DataTotal
		etcdJobInfo.DataFailed = request.DataFailed
		etcdJobInfo.DataRecovered = request.DataRecovered
		etcdJobInfo.StatsCommit = request.StatsCommit
		etcdJobInfo.Started = request.Started
		etcdJobInfo.Finished = request.Finished
		return etcdJobInfo.Job, nil
	})

	env.MockPachd.PPS.InspectJob.Use(func(ctx context.Context, request *pps.InspectJobRequest) (*pps.JobInfo, error) {
		outputCommitInfo, err := env.PachClient.InspectCommit(etcdJobInfo.OutputCommit.Repo.Name, etcdJobInfo.OutputCommit.ID)
		require.NoError(t, err)

		fmt.Printf("inspecting job for commit %v\n", outputCommitInfo)

		input := ppsutil.JobInput(pi, outputCommitInfo)
		fmt.Printf("job input: %v\n", input)

		files, err := env.PachClient.ListFile(input.Pfs.Repo, input.Pfs.Commit, input.Pfs.Glob)
		require.NoError(t, err)

		fmt.Printf("input files: %v\n", files)

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

		// If setting the job to a terminal state, we are done
		// TODO: this sometimes cancels the transform master too early
		if request.State == pps.JobState_JOB_SUCCESS {
			cancel()
		}

		return &types.Empty{}, nil
	})

	triggerJob(t, env, pi)

	return ctx, etcdJobInfo
}

func TestJob(t *testing.T) {
	pi := defaultPipelineInfo()
	t.Parallel()
	err := withWorkerSpawnerPair(pi, func(env *testEnv) error {
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, etcdJobInfo.State, pps.JobState_JOB_SUCCESS)
		return nil
	})
	require.NoError(t, err)
}
