package pjs

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"k8s.io/apimachinery/pkg/util/rand"
	"math"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	storagesrv "github.com/pachyderm/pachyderm/v2/src/internal/storage"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreateJob(t *testing.T) {
	t.Run("valid/parent/nil", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		_, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: testFileset,
			Input:   []string{testFileset},
		})
		require.NoError(t, err)
	})
}

func TestInspectJob(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = &testPermitter{mode: permitterAllow}
		})
		ctx := pctx.TestContext(t)
		programFileset := createFileSet(t, fc, map[string][]byte{
			"file/path": []byte(`!#/bin/bash; ls /input/;`),
		})
		inputFileset := createFileSet(t, fc, map[string][]byte{
			"a.txt": []byte(`dummy input`),
		})
		createJobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: programFileset,
			Input:   []string{inputFileset},
		})
		require.NoError(t, err)
		inspectJobResp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: createJobResp.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, inspectJobResp.Details)
		require.NotNil(t, inspectJobResp.Details.JobInfo)
		jobInfo := inspectJobResp.Details.JobInfo
		require.Equal(t, jobInfo.Job.Id, createJobResp.Id.Id)
		// the job is queued
		require.Equal(t, int32(jobInfo.State), int32(1))
		require.Equal(t, jobInfo.Program, programFileset)
		require.Equal(t, len(jobInfo.Input), 1)
		require.Equal(t, jobInfo.Input[0], inputFileset)
	})
	t.Run("invalid/inspect a non-existent job", func(t *testing.T) {
		c, _ := setupTest(t)
		ctx := pctx.TestContext(t)
		_, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: &pjs.Job{
				Id: 1,
			},
		})
		require.YesError(t, err)
		s := status.Convert(err)
		require.Equal(t, codes.NotFound, s.Code())
	})
}

// TestRunJob tests processQueue, AwaitJob, InspectJob as a whole
func TestRunJob(t *testing.T) {
	c, fc := setupTest(t)
	ctx := pctx.TestContext(t)
	programFileset := createFileSet(t, fc, map[string][]byte{
		"file": []byte(`!#/bin/bash; ls /input/;`),
	})
	inputFileset := createFileSet(t, fc, map[string][]byte{
		"a.txt": []byte("dummy input"),
	})
	in := &pjs.CreateJobRequest{
		Program: programFileset,
		Input:   []string{inputFileset},
	}

	out, err := runJob(t, ctx, c, fc, in, func(resp *pjs.ProcessQueueResponse) error {
		// for now we do nothing here, simply return the input filesets
		resp.Input = in.Input
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, in.Input, out.Output)
}

func TestCancelJob(t *testing.T) {
	t.Run("cancel job before processing", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		programFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		resp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: programFileset,
			Input:   []string{programFileset},
		})
		require.NoError(t, err)
		jid := resp.Id.Id
		_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{
			Job: &pjs.Job{
				Id: jid,
			},
		})
		require.NoError(t, err)
		inspectJobResp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: &pjs.Job{
				Id: jid,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, inspectJobResp.Details)
		jobInfo := inspectJobResp.Details.JobInfo
		require.Equal(t, pjs.JobState_DONE, jobInfo.State)
		// todo: the queue should be empty. This can be tested when inspectQueue is ready
	})
	t.Run("cancel a processing job", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)

		program := createFileSet(t, fc, map[string][]byte{"program": []byte("foo")})
		createResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: program,
			Input:   []string{program},
		})
		require.NoError(t, err)

		cancel := func() {
			_, err := c.CancelJob(ctx, &pjs.CancelJobRequest{
				Job: &pjs.Job{
					Id: createResp.Id.Id,
				},
			})
			require.NoError(t, err)
		}

		_, err = runJobFrom(t, ctx, c, fc, createResp.Id.Id, program, func(resp *pjs.ProcessQueueResponse) error {
			cancel()
			// simulate processing time and ensure cancelJob is finished
			time.Sleep(5 * time.Second)
			return nil
		})
		require.NoError(t, err)

		inspectCancelledJob(t, ctx, c, createResp.Id.Id)
	})
	t.Run("cancel processing job tree", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		eg, egCtx := errgroup.WithContext(ctx)

		jobAProcessing := make(chan struct{})
		jobBProcessing := make(chan struct{})
		jobCProcessing := make(chan struct{})

		var jobAContext string
		programA := createFileSet(t, fc, map[string][]byte{"program": []byte("A")})
		createRespA, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: programA,
			Input:   []string{programA},
		})
		require.NoError(t, err)
		eg.Go(func() error {
			_, err := runJobFrom(t, egCtx, c, fc, createRespA.Id.Id, programA, func(resp *pjs.ProcessQueueResponse) error {
				jobAContext = resp.Context
				close(jobAProcessing)
				time.Sleep(5 * time.Second)
				return nil
			})
			return err
		})
		<-jobAProcessing

		programB := createFileSet(t, fc, map[string][]byte{"program": []byte("B")})
		createRespB, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: programB,
			Input:   []string{programB},
			Context: jobAContext,
		})
		require.NoError(t, err)
		var jobBContext string
		eg.Go(func() error {
			_, err := runJobFrom(t, egCtx, c, fc, createRespB.Id.Id, programB, func(resp *pjs.ProcessQueueResponse) error {
				jobBContext = resp.Context
				close(jobBProcessing)
				time.Sleep(5 * time.Second)
				return nil
			})
			return err
		})
		<-jobBProcessing

		programC := createFileSet(t, fc, map[string][]byte{"program": []byte("C")})
		createRespC, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: programC,
			Input:   []string{programC},
			Context: jobBContext,
		})
		require.NoError(t, err)
		eg.Go(func() error {
			_, err := runJobFrom(t, egCtx, c, fc, createRespC.Id.Id, programC, func(resp *pjs.ProcessQueueResponse) error {
				close(jobCProcessing)
				time.Sleep(5 * time.Second)
				return nil
			})
			return err
		})
		<-jobCProcessing

		// Cancel job A while A, B, and C are all processing
		_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{
			Job: &pjs.Job{
				Id: createRespA.Id.Id,
			},
		})
		require.NoError(t, err)

		require.NoError(t, eg.Wait())

		inspectCancelledJob(t, ctx, c, createRespA.Id.Id)
		inspectCancelledJob(t, ctx, c, createRespB.Id.Id)
		inspectCancelledJob(t, ctx, c, createRespC.Id.Id)
	})
}

func TestWalkJob(t *testing.T) {
	c, fc := setupTest(t)
	ctx := pctx.TestContext(t)
	depth := 3
	fullBinaryJobTree(t, ctx, depth, c, fc)

	expectedSize := int64(math.Pow(2, float64(depth)) - 1)
	expected := make(map[int64][]int64)
	for i := int64(1); i <= expectedSize; i++ {
		left := 2 * i
		if left <= expectedSize {
			expected[i] = append(expected[i], left)
		}
		right := 2*i + 1
		if right <= expectedSize {
			expected[i] = append(expected[i], right)

		}
	}
	walkC, err := c.WalkJob(ctx, &pjs.WalkJobRequest{
		Job: &pjs.Job{
			Id: 1,
		},
		Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER,
	})
	require.NoError(t, err)
	actual := make(map[int64][]int64)
	for {
		resp, err := walkC.Recv()
		if err != nil {
			if errors.As(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
		parentJob := resp.Details.JobInfo.ParentJob
		if parentJob != nil && parentJob.Id != 0 {
			actual[parentJob.Id] = append(actual[parentJob.Id], resp.Id.Id)
		}
		require.NoError(t, err)
	}
	require.NoDiff(t, expected, actual, nil)
}

// runJob does work through PJS.
func runJob(t *testing.T, ctx context.Context, c pjs.APIClient, fc storage.FilesetClient, in *pjs.CreateJobRequest,
	fn func(resp *pjs.ProcessQueueResponse) error) (*pjs.JobInfo_Success, error) {
	jres, err := c.CreateJob(ctx, in)
	require.NoError(t, err)
	return runJobFrom(t, ctx, c, fc, jres.Id.Id, in.Program, fn)
}

func runJobFrom(t *testing.T, ctx context.Context, c pjs.APIClient, fc storage.FilesetClient, from int64, programStr string,
	fn func(resp *pjs.ProcessQueueResponse) error) (*pjs.JobInfo_Success, error) {
	programHash, err := HashFileset(ctx, fc, programStr)
	require.NoError(t, err)
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ctx, cf := context.WithCancel(ctx)
		defer cf()
		pqc, err := c.ProcessQueue(ctx)
		require.NoError(t, err)
		err = processQueue(pqc, programHash, fn)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if status.Code(err) == codes.Canceled {
			err = nil
		}
		return err
	})
	var ret *pjs.JobInfo_Success
	eg.Go(func() error {
		jobInfo, err := await(ctx, c, from)
		if err != nil {
			return err
		}
		ret = jobInfo.GetSuccess()
		cf() // success, cancel the other gorountine
		return nil
	})
	err = eg.Wait()
	require.NoError(t, err)
	return ret, nil
}

func processQueue(pqc pjs.API_ProcessQueueClient, programHash []byte, fn func(resp *pjs.ProcessQueueResponse) error) error {
	if err := pqc.Send(&pjs.ProcessQueueRequest{
		Queue: &pjs.Queue{Id: programHash},
	}); err != nil {
		return err
	}
	for {
		msg, err := pqc.Recv()
		if err != nil {
			return err
		}
		err = fn(msg)
		if err != nil {
			if sendErr := pqc.Send(&pjs.ProcessQueueRequest{
				Result: &pjs.ProcessQueueRequest_Failed{
					Failed: true,
				},
			}); sendErr != nil {
				return sendErr
			}
		} else {
			// todo(muayng): if send returns an error, the ctx on the server side will be cancelled. Job cannot transit to
			// Done state. Await will spin forever.
			if err := pqc.Send(&pjs.ProcessQueueRequest{
				Result: &pjs.ProcessQueueRequest_Success_{
					Success: &pjs.ProcessQueueRequest_Success{
						Output: msg.Input,
					},
				},
			}); err != nil {
				return err
			}
		}
	}
}

// await blocks until a Job enters the DONE state
func await(ctx context.Context, s pjs.APIClient, jid int64) (*pjs.JobInfo, error) {
	_, err := s.Await(ctx, &pjs.AwaitRequest{
		Job:          jid,
		DesiredState: pjs.JobState_DONE,
	})
	if err != nil {
		return nil, err
	}
	resp, err := s.InspectJob(ctx, &pjs.InspectJobRequest{
		Job: &pjs.Job{Id: jid},
	})
	if err != nil {
		return nil, err
	}
	return resp.Details.JobInfo, nil
}

func TestAwaitJob(t *testing.T) {
	t.Run("invalid/job does not exist", func(t *testing.T) {
		c, _ := setupTest(t)
		ctx := pctx.TestContext(t)
		_, err := c.Await(ctx, &pjs.AwaitRequest{
			Job:          10,
			DesiredState: pjs.JobState_DONE,
		})
		require.YesError(t, err)
		s := status.Convert(err)
		require.Equal(t, codes.NotFound, s.Code())
	})
	t.Run("invalid/time out", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		_, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: testFileset,
			Input:   []string{testFileset},
		})
		require.NoError(t, err)
		_, err = c.Await(ctx, &pjs.AwaitRequest{
			Job:          1,
			DesiredState: pjs.JobState_DONE,
		})
		require.YesError(t, err)
		s := status.Convert(err)
		require.Equal(t, codes.DeadlineExceeded, s.Code())
	})
	// valid case is tested in TestRunJob with ProcessQueue
}

func setupTest(t testing.TB, opts ...ClientOptions) (pjs.APIClient, storage.FilesetClient) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	client := storagesrv.NewTestFilesetClient(t, db)
	return NewTestClient(t, db, client, opts...), client
}

func createFileSet(t *testing.T, fc storage.FilesetClient, files map[string][]byte) string {
	ctx := pctx.TestContext(t)
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	createFileClient, err := fc.CreateFileset(ctx)
	require.NoError(t, err)
	for path, data := range files {
		require.NoError(t, createFileClient.Send(&storage.CreateFilesetRequest{
			Modification: &storage.CreateFilesetRequest_AppendFile{
				AppendFile: &storage.AppendFile{
					Path: path,
					Data: &wrapperspb.BytesValue{
						Value: data,
					},
				},
			},
		}))
	}
	resp, err := createFileClient.CloseAndRecv()
	require.NoError(t, err)
	return resp.FilesetId
}

func fullBinaryJobTree(t *testing.T, ctx context.Context, maxDepth int, c pjs.APIClient, fc storage.FilesetClient) {
	if maxDepth == 0 {
		return
	}

	// create node at depth == 1
	program := createFileSet(t, fc, map[string][]byte{"program": []byte("foo")})
	req := &pjs.CreateJobRequest{
		Program: program,
		Input:   []string{program},
	}
	createResp, err := c.CreateJob(ctx, req)
	require.NoError(t, err)
	var processResp *pjs.ProcessQueueResponse
	_, err = runJobFrom(t, ctx, c, fc, createResp.Id.Id, program, func(resp *pjs.ProcessQueueResponse) error {
		processResp = resp
		return nil
	})
	require.NoError(t, err)

	parents := make([]*pjs.ProcessQueueResponse, 0)
	parents = append(parents, processResp)

	// create nodes at depth > 1
	for depth := 2; depth <= maxDepth; depth++ {
		newParents := make([]*pjs.ProcessQueueResponse, 0)
		for _, p := range parents {
			createChild := func() *pjs.ProcessQueueResponse {
				prog := createFileSet(t, fc, map[string][]byte{"program": []byte(rand.String(32))})
				req := &pjs.CreateJobRequest{
					Context: p.Context,
					Program: prog,
					Input:   []string{prog},
				}
				cResp, err := c.CreateJob(ctx, req)
				require.NoError(t, err)
				var pResp *pjs.ProcessQueueResponse
				_, err = runJobFrom(t, ctx, c, fc, cResp.Id.Id, prog, func(resp *pjs.ProcessQueueResponse) error {
					pResp = resp
					return nil
				})
				require.NoError(t, err)
				return pResp
			}
			newParents = append(newParents, createChild(), createChild())
		}
		parents = newParents
	}
}

func inspectCancelledJob(t *testing.T, ctx context.Context, c pjs.APIClient, id int64) {
	inspectJobResp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
		Job: &pjs.Job{
			Id: id,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, inspectJobResp.Details)
	jobInfo := inspectJobResp.Details.JobInfo
	// the processing job is still processed successfully, but the output
	// does not get written back to db
	require.Nil(t, jobInfo.GetSuccess())
}

func TestAuth(t *testing.T) {
	//todo: add other functions when implemented: delete, list, etc.
	t.Run("admin", func(t *testing.T) {
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = &testPermitter{mode: permitterAllow}
		})
		ctx := pctx.TestContext(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset, Input: []string{testFileset}})
		require.NoError(t, err)
		walkClient, err := c.WalkJob(ctx, &pjs.WalkJobRequest{Job: jobResp.Id, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
		require.NoError(t, err)
		for _, err := walkClient.Recv(); err != io.EOF; _, err = walkClient.Recv() {
			require.NoError(t, err)
		}
		_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{Job: jobResp.Id})
		require.NoError(t, err)
		_, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Job: jobResp.Id})
		require.NoError(t, err)
	})
	t.Run("permission_denied", func(t *testing.T) {
		p := &testPermitter{mode: permitterAllow}
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = p
		})
		ctx := pctx.TestContext(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job as admin.
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset, Input: []string{testFileset}})
		require.NoError(t, err)
		// flip the permitter to deny
		p.mode = permitterDeny
		walkClient, err := c.WalkJob(ctx, &pjs.WalkJobRequest{Job: jobResp.Id, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
		require.NoError(t, err)
		_, err = walkClient.Recv()
		require.YesError(t, err)
		s := status.Convert(err)
		require.Equal(t, codes.PermissionDenied, s.Code())
		_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{Job: jobResp.Id})
		require.YesError(t, err)
		s = status.Convert(err)
		require.Equal(t, codes.PermissionDenied, s.Code())
		_, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Job: jobResp.Id})
		require.YesError(t, err)
		s = status.Convert(err)
		require.Equal(t, codes.PermissionDenied, s.Code())
		require.YesError(t, err)
		s = status.Convert(err)
		require.Equal(t, codes.PermissionDenied, s.Code())
	})
	t.Run("valid_job_context", func(t *testing.T) {
		p := &testPermitter{mode: permitterAllow}
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = p
		})
		ctx := pctx.TestContext(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job as admin.
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset, Input: []string{testFileset}})
		require.NoError(t, err)

		var jobContext string
		// run process queue to get a job context token.
		_, err = runJobFrom(t, ctx, c, fc, jobResp.Id.Id, testFileset, func(resp *pjs.ProcessQueueResponse) error {
			jobContext = resp.Context
			return nil
		})
		require.NoError(t, err)
		// wait for the job to finish ourselves. runJobFrom calls await as well, but in another goroutine.
		// await currently polls using the inspect RPC, which will require auth.
		_, err = await(ctx, c, jobResp.Id.Id)
		// flip the permitter.
		p.mode = permitterDeny
		require.NoError(t, err)
		jobResp, err = c.CreateJob(ctx, &pjs.CreateJobRequest{Context: jobContext, Program: testFileset, Input: []string{testFileset}})
		require.NoError(t, err)
		walkClient, err := c.WalkJob(ctx, &pjs.WalkJobRequest{Context: jobContext, Job: jobResp.Id, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
		require.NoError(t, err)
		for _, err := walkClient.Recv(); err != io.EOF; _, err = walkClient.Recv() {
			require.NoError(t, err)
		}
		_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{Context: jobContext, Job: jobResp.Id})
		require.NoError(t, err)
		_, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Context: jobContext, Job: jobResp.Id})
		require.NoError(t, err)
	})
	t.Run("invalid_job_context", func(t *testing.T) {
		p := &testPermitter{mode: permitterAllow}
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = p
		})
		ctx := pctx.TestContext(t)
		testFileset := createFileSet(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job without ctx as admin.
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset, Input: []string{testFileset}})
		require.NoError(t, err)
		modes := []permitterEnum{permitterAllow, permitterDeny}
		jobCtx := "00001111222233334444555566666777788889999aaaabbbbccccddddeeeefff"
		for _, mode := range modes {
			p.mode = mode
			_, err = c.CreateJob(ctx, &pjs.CreateJobRequest{Context: jobCtx, Program: testFileset, Input: []string{testFileset}})
			require.YesError(t, err)
			t.Log(err)
			s := status.Convert(err)
			require.Equal(t, codes.NotFound, s.Code())
			walkClient, err := c.WalkJob(ctx, &pjs.WalkJobRequest{Context: jobCtx, Job: jobResp.Id, Algorithm: pjs.WalkAlgorithm_LEVEL_ORDER})
			require.NoError(t, err)
			_, err = walkClient.Recv()
			require.YesError(t, err)
			s = status.Convert(err)
			require.Equal(t, codes.NotFound, s.Code())
			_, err = c.CancelJob(ctx, &pjs.CancelJobRequest{Context: jobCtx, Job: jobResp.Id})
			require.YesError(t, err)
			s = status.Convert(err)
			require.Equal(t, codes.NotFound, s.Code())
			_, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Context: jobCtx, Job: jobResp.Id})
			require.YesError(t, err)
			s = status.Convert(err)
			require.Equal(t, codes.NotFound, s.Code())
		}
	})
}
