package transform

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/work"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func newWorkerSpawnerPair(t *testing.T, dbConfig serviceenv.ConfigOption, pipelineInfo *pps.PipelineInfo) *testEnv {
	// We only support simple pfs input pipelines in this test suite at the moment
	require.NotNil(t, pipelineInfo.Input)
	require.NotNil(t, pipelineInfo.Input.Pfs)

	env := newTestEnv(t, dbConfig, pipelineInfo)

	eg, ctx := errgroup.WithContext(env.driver.PachClient().Ctx())
	t.Cleanup(func() {
		require.NoError(t, eg.Wait())
	})

	env.driver = env.driver.WithContext(ctx)
	env.PachClient = env.driver.PachClient()

	// Set env vars that the object storage layer expects in the env
	// This is global but it should be fine because all tests use the same value.
	require.NoError(t, os.Setenv(obj.StorageBackendEnvVar, obj.Local))
	require.NoError(t, os.MkdirAll(env.ServiceEnv.Config().StorageRoot, 0777))

	// Set up repos and branches for the pipeline
	input := pipelineInfo.Input.Pfs
	require.NoError(t, env.PachClient.CreateRepo(input.Repo))
	require.NoError(t, env.PachClient.CreateBranch(input.Repo, input.Branch, "", "", nil))

	_, err := env.PachClient.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{Branch: pipelineInfo.SpecCommit.Branch})
	require.NoError(t, err)

	commit, err := env.PachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{Branch: pipelineInfo.SpecCommit.Branch})
	require.NoError(t, err)
	pipelineInfo.SpecCommit = commit

	_, err = env.PachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{Commit: commit})
	require.NoError(t, err)

	err = env.PachClient.CreateRepo(pipelineInfo.Pipeline.Name)
	require.NoError(t, err)

	err = env.PachClient.CreateBranch(
		pipelineInfo.Pipeline.Name,
		pipelineInfo.OutputBranch,
		"",
		"",
		[]*pfs.Branch{
			client.NewBranch(input.Repo, input.Branch),
			pipelineInfo.SpecCommit.Branch.Repo.NewBranch(pipelineInfo.Pipeline.Name),
		},
	)
	require.NoError(t, err)

	metaRepo := client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.MetaRepoType)
	_, err = env.PachClient.PfsAPIClient.CreateBranch(
		env.PachClient.Ctx(),
		&pfs.CreateBranchRequest{
			Head:       metaRepo.NewCommit("master", ""),
			Branch:     metaRepo.NewBranch("master"),
			Provenance: []*pfs.Branch{client.NewBranch(pipelineInfo.Pipeline.Name, pipelineInfo.OutputBranch)},
		})
	require.NoError(t, err)

	// Put the pipeline info into etcd (which is read by the master)
	err = env.driver.NewSQLTx(func(sqlTx *sqlx.Tx) error {
		etcdPipelineInfo := &pps.StoredPipelineInfo{
			State:       pps.PipelineState_PIPELINE_STARTING,
			SpecCommit:  pipelineInfo.SpecCommit,
			Parallelism: 1,
		}
		return env.driver.Pipelines().ReadWrite(sqlTx).Put(pipelineInfo.Pipeline.Name, etcdPipelineInfo)
	})
	require.NoError(t, err)

	eg.Go(func() error {
		if err := Run(env.driver, env.logger); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	eg.Go(func() error {
		err := backoff.RetryUntilCancel(env.driver.PachClient().Ctx(), func() error {
			return env.driver.NewTaskWorker().Run(
				env.driver.PachClient().Ctx(),
				func(ctx context.Context, subtask *work.Task) (*types.Any, error) {
					status := &Status{}
					return nil, Worker(env.driver, env.logger, subtask, status)
				},
			)
		}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
			env.logger.Logf("worker failed, retrying immediately, err: %v", err)
			return nil
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})

	return env
}

func withTimeout(ctx context.Context, duration time.Duration) context.Context {
	// Create a context that the caller can wait on
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-time.After(duration):
			fmt.Printf("Canceling test after timeout\n")
			cancel()
		}
	}()

	return ctx
}

func mockBasicJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo) (context.Context, *pps.StoredPipelineJobInfo) {
	// Create a context that the caller can wait on
	ctx, cancel := context.WithCancel(env.PachClient.Ctx())

	// Mock out the initial ListPipelineJob, CreatePipelineJob, and InspectPipelineJob calls
	etcdJobInfo := &pps.StoredPipelineJobInfo{PipelineJob: client.NewPipelineJob(uuid.NewWithoutDashes())}

	// TODO: use a 'real' pps if we can make one that doesn't need a real kube client
	env.MockPachd.PPS.ListPipelineJob.Use(func(*pps.ListPipelineJobRequest, pps.API_ListPipelineJobServer) error {
		return nil
	})

	env.MockPachd.PPS.CreatePipelineJob.Use(func(ctx context.Context, request *pps.CreatePipelineJobRequest) (*pps.PipelineJob, error) {
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
		return etcdJobInfo.PipelineJob, nil
	})

	env.MockPachd.PPS.InspectPipelineJob.Use(func(ctx context.Context, request *pps.InspectPipelineJobRequest) (*pps.PipelineJobInfo, error) {
		if etcdJobInfo.OutputCommit == nil {
			return nil, errors.Errorf("job with output commit %s not found", request.OutputCommit.ID)
		}
		outputCommitInfo, err := env.PachClient.InspectCommit(etcdJobInfo.OutputCommit.Branch.Repo.Name, etcdJobInfo.OutputCommit.Branch.Name, etcdJobInfo.OutputCommit.ID)
		require.NoError(t, err)

		return &pps.PipelineJobInfo{
			PipelineJob:      etcdJobInfo.PipelineJob,
			Pipeline:         etcdJobInfo.Pipeline,
			OutputRepo:       client.NewRepo(etcdJobInfo.Pipeline.Name),
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
			Input:            ppsutil.PipelineJobInput(pi, outputCommitInfo),
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

	updateJobState := func(request *pps.UpdatePipelineJobStateRequest) {
		if ppsutil.IsTerminal(etcdJobInfo.State) {
			return
		}

		etcdJobInfo.State = request.State
		etcdJobInfo.Reason = request.Reason

		// If setting the job to a terminal state, we are done
		if ppsutil.IsTerminal(request.State) {
			cancel()
		}
	}

	env.MockPPSTransactionServer.UpdatePipelineJobStateInTransaction.Use(func(txnctx *txnenv.TransactionContext, request *pps.UpdatePipelineJobStateRequest) error {
		updateJobState(request)
		return nil
	})

	env.MockPachd.PPS.UpdatePipelineJobState.Use(func(ctx context.Context, request *pps.UpdatePipelineJobStateRequest) (*types.Empty, error) {
		updateJobState(request)
		return &types.Empty{}, nil
	})

	return ctx, etcdJobInfo
}

func triggerJob(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []tarutil.File) {
	commit, err := env.PachClient.StartCommit(pi.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	buf := &bytes.Buffer{}
	require.NoError(t, tarutil.WithWriter(buf, func(tw *tar.Writer) error {
		for _, f := range files {
			if err := tarutil.WriteFile(tw, f); err != nil {
				return err
			}
		}
		return nil
	}))
	require.NoError(t, env.PachClient.PutFileTar(commit, buf, client.WithAppendPutFile()))
	require.NoError(t, env.PachClient.FinishCommit(pi.Input.Pfs.Repo, commit.Branch.Name, commit.ID))
}

func testJobSuccess(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []tarutil.File) {
	ctx, etcdJobInfo := mockBasicJob(t, env, pi)
	triggerJob(t, env, pi, files)
	ctx = withTimeout(ctx, 10*time.Second)
	<-ctx.Done()
	require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

	// Ensure the output commit is successful
	outputCommitID := etcdJobInfo.OutputCommit.ID
	outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, etcdJobInfo.OutputCommit.Branch.Name, outputCommitID)
	require.NoError(t, err)
	require.NotNil(t, outputCommitInfo.Finished)

	branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
	require.NoError(t, err)
	require.NotNil(t, branchInfo)

	r, err := env.PachClient.GetFileTar(etcdJobInfo.OutputCommit, "/*")
	require.NoError(t, err)
	require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
		ok, err := tarutil.Equal(files[0], file)
		require.NoError(t, err)
		require.True(t, ok)
		files = files[1:]
		return nil
	}))
}

func TestTransformPipeline(suite *testing.T) {
	suite.Parallel()

	suite.Run("TestJobSuccess", func(t *testing.T) {
		t.Parallel()
		pi := defaultPipelineInfo()
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)
		testJobSuccess(t, env, pi, []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		})
	})

	suite.Run("TestJobSuccessEgress", func(t *testing.T) {
		t.Parallel()
		objC, bucket := testutil.NewObjectClient(t)
		pi := defaultPipelineInfo()
		pi.Egress = &pps.Egress{URL: fmt.Sprintf("local://%s/", bucket)}
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)

		files := []tarutil.File{
			tarutil.NewMemFile("/file1", []byte("foo")),
			tarutil.NewMemFile("/file2", []byte("bar")),
		}
		testJobSuccess(t, env, pi, files)
		for _, file := range files {
			hdr, err := file.Header()
			require.NoError(t, err)

			buf1 := &bytes.Buffer{}
			require.NoError(t, file.Content(buf1))

			buf2 := &bytes.Buffer{}
			err = objC.Get(context.Background(), hdr.Name, buf2)
			require.NoError(t, err)

			require.True(t, bytes.Equal(buf1.Bytes(), buf2.Bytes()))
		}
	})

	suite.Run("TestJobFailedDatum", func(t *testing.T) {
		t.Parallel()
		pi := defaultPipelineInfo()
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)

		pi.Transform.Cmd = []string{"bash", "-c", "(exit 1)"}
		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		}
		triggerJob(t, env, pi, tarFiles)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_FAILURE, etcdJobInfo.State)
		// TODO: check job stats
	})

	suite.Run("TestJobMultiDatum", func(t *testing.T) {
		t.Parallel()
		pi := defaultPipelineInfo()
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)

		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJob(t, env, pi, tarFiles)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful.
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, etcdJobInfo.OutputCommit.Branch.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetFileTar(etcdJobInfo.OutputCommit, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			tarFiles = tarFiles[1:]
			return nil
		}))
	})

	suite.Run("TestJobSerial", func(t *testing.T) {
		t.Parallel()
		pi := defaultPipelineInfo()
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)

		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJob(t, env, pi, tarFiles[:1])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		triggerJob(t, env, pi, tarFiles[1:])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, etcdJobInfo.OutputCommit.Branch.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetFileTar(etcdJobInfo.OutputCommit, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			tarFiles = tarFiles[1:]
			return nil
		}))
	})

	suite.Run("TestJobSerialDelete", func(t *testing.T) {
		t.Parallel()
		pi := defaultPipelineInfo()
		env := newWorkerSpawnerPair(t, testutil.NewTestDBConfig(t), pi)

		ctx, etcdJobInfo := mockBasicJob(t, env, pi)
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		triggerJob(t, env, pi, tarFiles[:1])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		triggerJob(t, env, pi, tarFiles[1:])
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		ctx, etcdJobInfo = mockBasicJob(t, env, pi)
		deleteFiles(t, env, pi, []string{"/a"})
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.PipelineJobState_JOB_SUCCESS, etcdJobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := etcdJobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, etcdJobInfo.OutputCommit.Branch.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetFileTar(etcdJobInfo.OutputCommit, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(tarFiles[1], file)
			require.NoError(t, err)
			require.True(t, ok)
			return nil
		}))
	})
}

func deleteFiles(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []string) {
	commit, err := env.PachClient.StartCommit(pi.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	for _, file := range files {
		require.NoError(t, env.PachClient.DeleteFile(commit, file))
	}
	require.NoError(t, env.PachClient.FinishCommit(pi.Input.Pfs.Repo, commit.Branch.Name, commit.ID))
}
