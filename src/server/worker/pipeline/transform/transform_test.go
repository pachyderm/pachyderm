package transform

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd/realenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

func setupPachAndWorker(t *testing.T, dbConfig serviceenv.ConfigOption, pipelineInfo *pps.PipelineInfo) *testEnv {
	// We only support simple pfs input pipelines in this test suite at the moment
	require.NotNil(t, pipelineInfo.Details.Input)
	require.NotNil(t, pipelineInfo.Details.Input.Pfs)

	env := realenv.NewRealEnvWithPPSTransactionMock(t, dbConfig)
	eg, ctx := errgroup.WithContext(env.PachClient.Ctx())

	// Set env vars that the object storage layer expects in the env
	// This is global but it should be fine because all tests use the same value.
	require.NoError(t, os.Setenv(obj.StorageBackendEnvVar, obj.Local))
	require.NoError(t, os.MkdirAll(env.ServiceEnv.Config().StorageRoot, 0777))

	// Set up the input repo and branch
	input := pipelineInfo.Details.Input.Pfs
	require.NoError(t, env.PachClient.CreateRepo(input.Repo))
	require.NoError(t, env.PachClient.CreateBranch(input.Repo, input.Branch, "", "", nil))

	// Create the output repo
	pipelineRepo := client.NewRepo(pipelineInfo.Pipeline.Name)
	_, err := env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: pipelineRepo})
	require.NoError(t, err)

	// Create the spec system repo and create the initial spec commit
	specRepo := client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.SpecRepoType)
	_, err = env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: specRepo})
	require.NoError(t, err)
	specCommit, err := env.PachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{Branch: specRepo.NewBranch("master")})
	require.NoError(t, err)
	_, err = env.PachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{Commit: specCommit})
	require.NoError(t, err)

	// Make the output branch provenant on the spec and input branches
	_, err = env.PachClient.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{
		Branch: pipelineRepo.NewBranch(pipelineInfo.Details.OutputBranch),
		Provenance: []*pfs.Branch{
			client.NewBranch(input.Repo, input.Branch),
			specRepo.NewBranch("master"),
		},
	})
	require.NoError(t, err)
	err = closeHeadCommit(ctx, env, pipelineRepo.NewBranch(pipelineInfo.Details.OutputBranch))
	require.NoError(t, err)
	// Create the meta system repo and set up the branch provenance
	metaRepo := client.NewSystemRepo(pipelineInfo.Pipeline.Name, pfs.MetaRepoType)
	_, err = env.PachClient.PfsAPIClient.CreateRepo(ctx, &pfs.CreateRepoRequest{Repo: metaRepo})
	require.NoError(t, err)
	branch := metaRepo.NewBranch("master")
	_, err = env.PachClient.PfsAPIClient.CreateBranch(ctx, &pfs.CreateBranchRequest{
		Branch: branch,
		Provenance: []*pfs.Branch{
			client.NewBranch(input.Repo, input.Branch),
			specRepo.NewBranch("master"),
		},
	})
	require.NoError(t, err)
	// the worker needs all meta commits to have an output commit. we force close this so its skipped by worker code
	err = closeHeadCommit(ctx, env, metaRepo.NewBranch("master"))
	require.NoError(t, err)
	// we create a new spec commit and CLOSE the corresponding meta and output commits so that the worker can process the base commit
	specCommit, err = env.PachClient.PfsAPIClient.StartCommit(ctx, &pfs.StartCommitRequest{Branch: specRepo.NewBranch("master")})
	require.NoError(t, err)
	_, err = env.PachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{Commit: specCommit, Force: true})
	require.NoError(t, err)
	err = closeHeadCommit(ctx, env, metaRepo.NewBranch("master"))
	require.NoError(t, err)
	err = closeHeadCommit(ctx, env, pipelineRepo.NewBranch(pipelineInfo.Details.OutputBranch))
	require.NoError(t, err)

	pipelineInfo.SpecCommit = specCommit
	testEnv := newTestEnv(t, pipelineInfo, env)
	testEnv.driver = testEnv.driver.WithContext(ctx)
	testEnv.PachClient = testEnv.driver.PachClient()
	// Put the pipeline info into the collection (which is read by the master)
	err = testEnv.driver.NewSQLTx(func(sqlTx *pachsql.Tx) error {
		rw := testEnv.driver.Pipelines().ReadWrite(sqlTx)
		err := rw.Put(specCommit, pipelineInfo) // pipeline/info needs to contain spec
		return errors.EnsureStack(err)
	})
	require.NoError(t, err)
	eg.Go(func() error {
		err := backoff.RetryUntilCancel(testEnv.driver.PachClient().Ctx(), func() error {
			return Worker(testEnv.driver.PachClient().Ctx(), testEnv.driver, testEnv.logger, &Status{})
		}, &backoff.ZeroBackOff{}, func(err error, d time.Duration) error {
			testEnv.logger.Logf("worker failed, retrying immediately, err: %v", err)
			return nil
		})
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	})
	return testEnv
}

func closeHeadCommit(ctx context.Context, env *realenv.RealEnv, branch *pfs.Branch) error {
	branchInfo, err := env.PachClient.PfsAPIClient.InspectBranch(ctx, &pfs.InspectBranchRequest{Branch: branch})
	if err != nil {
		return errors.Wrap(err, "could not inspect branch")
	}
	if _, err := env.PachClient.PfsAPIClient.FinishCommit(ctx, &pfs.FinishCommitRequest{Commit: branchInfo.Head, Force: true}); err != nil {
		return errors.Wrap(err, "could not finish commit")
	}
	return nil
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

func mockJobFromCommit(t *testing.T, env *testEnv, pi *pps.PipelineInfo, commit *pfs.Commit) (context.Context, *pps.JobInfo) {
	// Create a context that the caller can wait on
	ctx, cancel := context.WithCancel(env.PachClient.Ctx())
	// Mock out the initial ListJob, and InspectJob calls
	jobInfo := &pps.JobInfo{Job: client.NewJob(pi.Pipeline.Name, commit.ID)}
	jobInfo.OutputCommit = client.NewCommit(pi.Pipeline.Name, pi.Details.OutputBranch, commit.ID)
	jobInfo.Details = &pps.JobInfo_Details{
		Transform:        pi.Details.Transform,
		ParallelismSpec:  pi.Details.ParallelismSpec,
		Egress:           pi.Details.Egress,
		Service:          pi.Details.Service,
		Spout:            pi.Details.Spout,
		ResourceRequests: pi.Details.ResourceRequests,
		ResourceLimits:   pi.Details.ResourceLimits,
		Input:            ppsutil.JobInput(pi, client.NewCommit(pi.Details.Input.Pfs.Repo, pi.Details.OutputBranch, commit.ID)),
		Salt:             pi.Details.Salt,
		DatumSetSpec:     pi.Details.DatumSetSpec,
		DatumTimeout:     pi.Details.DatumTimeout,
		JobTimeout:       pi.Details.JobTimeout,
		DatumTries:       pi.Details.DatumTries,
		SchedulingSpec:   pi.Details.SchedulingSpec,
		PodSpec:          pi.Details.PodSpec,
		PodPatch:         pi.Details.PodPatch,
	}
	env.MockPachd.PPS.InspectJob.Use(func(ctx context.Context, request *pps.InspectJobRequest) (*pps.JobInfo, error) {
		if request.Job.ID == jobInfo.OutputCommit.ID {
			result := proto.Clone(jobInfo).(*pps.JobInfo)
			return result, nil
		}
		mockJI := &pps.JobInfo{Job: client.NewJob(pi.Pipeline.Name, request.Job.ID)}
		mockJI.OutputCommit = client.NewCommit(pi.Pipeline.Name, pi.Details.OutputBranch, request.Job.ID)
		mockJI.Details = &pps.JobInfo_Details{
			Transform:        pi.Details.Transform,
			ParallelismSpec:  pi.Details.ParallelismSpec,
			Egress:           pi.Details.Egress,
			Service:          pi.Details.Service,
			Spout:            pi.Details.Spout,
			ResourceRequests: pi.Details.ResourceRequests,
			ResourceLimits:   pi.Details.ResourceLimits,
			Input:            ppsutil.JobInput(pi, client.NewCommit(pi.Details.Input.Pfs.Repo, pi.Details.OutputBranch, request.Job.ID)),
			Salt:             pi.Details.Salt,
			DatumSetSpec:     pi.Details.DatumSetSpec,
			DatumTimeout:     pi.Details.DatumTimeout,
			JobTimeout:       pi.Details.JobTimeout,
			DatumTries:       pi.Details.DatumTries,
			SchedulingSpec:   pi.Details.SchedulingSpec,
			PodSpec:          pi.Details.PodSpec,
			PodPatch:         pi.Details.PodPatch,
		}
		result := proto.Clone(mockJI).(*pps.JobInfo)
		return result, nil
	})

	updateJobState := func(request *pps.UpdateJobStateRequest) {
		jobInfo.State = request.State
		jobInfo.Reason = request.Reason

		// If setting the job to a terminal state, we are done
		if pps.IsTerminal(request.State) {
			cancel()
		}
	}
	env.MockPPSTransactionServer.UpdateJobStateInTransaction.Use(func(txnCtx *txncontext.TransactionContext, request *pps.UpdateJobStateRequest) error {
		updateJobState(request)
		return nil
	})
	env.MockPachd.PPS.UpdateJobState.Use(func(ctx context.Context, request *pps.UpdateJobStateRequest) (*types.Empty, error) {
		updateJobState(request)
		return &types.Empty{}, nil
	})

	eg, ctx := errgroup.WithContext(ctx)
	reg, err := newRegistry(env.driver, env.logger)
	require.NoError(t, err)
	eg.Go(func() error {
		_ = reg.startJob(proto.Clone(jobInfo).(*pps.JobInfo))
		return nil
	})
	return ctx, jobInfo
}

func writeFiles(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []tarutil.File) *pfs.Commit {
	commit, err := env.PachClient.StartCommit(pi.Details.Input.Pfs.Repo, "master")
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
	require.NoError(t, env.PachClient.PutFileTAR(commit, buf, client.WithAppendPutFile()))
	require.NoError(t, env.PachClient.FinishCommit(pi.Details.Input.Pfs.Repo, commit.Branch.Name, commit.ID))
	return commit
}

func deleteFiles(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []string) *pfs.Commit {
	commit, err := env.PachClient.StartCommit(pi.Details.Input.Pfs.Repo, "master")
	require.NoError(t, err)
	for _, file := range files {
		require.NoError(t, env.PachClient.DeleteFile(commit, file))
	}
	require.NoError(t, env.PachClient.FinishCommit(pi.Details.Input.Pfs.Repo, commit.Branch.Name, commit.ID))
	return commit
}

func testJobSuccess(t *testing.T, env *testEnv, pi *pps.PipelineInfo, files []tarutil.File) {
	commit := writeFiles(t, env, pi, files)
	ctx, jobInfo := mockJobFromCommit(t, env, pi, commit)
	ctx = withTimeout(ctx, 15*time.Second)
	<-ctx.Done()
	// PFS master transitions the job from finishing to finished so dont test that here
	require.Equal(t, pps.JobState_JOB_FINISHING, jobInfo.State)

	// Ensure the output commit is successful
	outputCommitID := jobInfo.OutputCommit.ID
	outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, jobInfo.OutputCommit.Branch.Name, outputCommitID)
	require.NoError(t, err)
	require.NotNil(t, outputCommitInfo.Finished)

	branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.Details.OutputBranch)
	require.NoError(t, err)
	require.NotNil(t, branchInfo)

	if files != nil {
		r, err := env.PachClient.GetFileTAR(jobInfo.OutputCommit, "/*")
		require.NoError(t, err)
		require.NoError(t, tarutil.Iterate(r, func(file tarutil.File) error {
			ok, err := tarutil.Equal(files[0], file)
			require.NoError(t, err)
			require.True(t, ok)
			files = files[1:]
			return nil
		}))
	}
}

func TestTransformPipeline(suite *testing.T) {
	suite.Parallel()

	suite.Run("TestJobSuccess", func(t *testing.T) {
		pi := defaultPipelineInfo()
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)
		testJobSuccess(t, env, pi, []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		})
	})

	suite.Run("TestJobSuccessEgress", func(t *testing.T) {
		objC := dockertestenv.NewTestObjClient(t)
		pi := defaultPipelineInfo()
		egressURL := objC.BucketURL().String()
		pi.Details.Egress = &pps.Egress{URL: egressURL}
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

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

	suite.Run("TestJobSuccessEgressEmpty", func(t *testing.T) {
		objC := dockertestenv.NewTestObjClient(t)
		pi := defaultPipelineInfo()
		pi.Details.Input.Pfs.Glob = "/"
		pi.Details.Egress = &pps.Egress{URL: objC.BucketURL().String()}
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

		testJobSuccess(t, env, pi, nil)
	})

	suite.Run("TestJobFailedDatum", func(t *testing.T) {
		pi := defaultPipelineInfo()
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

		pi.Details.Transform.Cmd = []string{"bash", "-c", "(exit 1)"}
		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/file", []byte("foobar")),
		}
		commit := writeFiles(t, env, pi, tarFiles)
		ctx, jobInfo := mockJobFromCommit(t, env, pi, commit)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FAILURE, jobInfo.State)
		// TODO: check job stats
	})

	suite.Run("TestJobMultiDatum", func(t *testing.T) {
		pi := defaultPipelineInfo()
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		testJobSuccess(t, env, pi, tarFiles)
	})

	suite.Run("TestJobSerial", func(t *testing.T) {
		pi := defaultPipelineInfo()
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
			tarutil.NewMemFile("/b", []byte("barfoo")),
		}
		commit := writeFiles(t, env, pi, tarFiles[:1])
		ctx, jobInfo := mockJobFromCommit(t, env, pi, commit)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FINISHING, jobInfo.State)

		commit = writeFiles(t, env, pi, tarFiles[1:])
		ctx, jobInfo = mockJobFromCommit(t, env, pi, commit)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FINISHING, jobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := jobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, jobInfo.OutputCommit.Branch.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.Details.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		r, err := env.PachClient.GetFileTAR(jobInfo.OutputCommit, "/*")
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
		pi := defaultPipelineInfo()
		env := setupPachAndWorker(t, dockertestenv.NewTestDBConfig(t), pi)

		tarFiles := []tarutil.File{
			tarutil.NewMemFile("/a", []byte("foobar")),
		}
		commit := writeFiles(t, env, pi, tarFiles[:1])
		ctx, jobInfo := mockJobFromCommit(t, env, pi, commit)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FINISHING, jobInfo.State)

		commit = deleteFiles(t, env, pi, []string{"/a"})
		ctx, jobInfo = mockJobFromCommit(t, env, pi, commit)
		ctx = withTimeout(ctx, 10*time.Second)
		<-ctx.Done()
		require.Equal(t, pps.JobState_JOB_FINISHING, jobInfo.State)

		// Ensure the output commit is successful
		outputCommitID := jobInfo.OutputCommit.ID
		outputCommitInfo, err := env.PachClient.InspectCommit(pi.Pipeline.Name, jobInfo.OutputCommit.Branch.Name, outputCommitID)
		require.NoError(t, err)
		require.NotNil(t, outputCommitInfo.Finished)

		branchInfo, err := env.PachClient.InspectBranch(pi.Pipeline.Name, pi.Details.OutputBranch)
		require.NoError(t, err)
		require.NotNil(t, branchInfo)

		// Get the output files.
		files, err := env.PachClient.ListFileAll(jobInfo.OutputCommit, "/*")
		require.NoError(t, err)
		require.Equal(t, len(files), 0)
	})
}
