package pjs

import (
	"context"
	"io"
	"math"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/rand"

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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		req := &pjs.CreateJobRequest{
			Program: testFileset.HexString(),
			Input:   []string{testFileset.HexString()},
		}
		createJob(ctx, t, c, req)
	})
}

func createJob(ctx context.Context, t *testing.T, c pjs.APIClient, req *pjs.CreateJobRequest) (*pjs.CreateJobResponse, *fileset.Handle) {
	createResp, err := c.CreateJob(ctx, req)
	require.NoError(t, err)
	inspectResp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
		Job: createResp.Id,
	})
	require.NoError(t, err)
	program, err := fileset.ParseHandleKeepID(inspectResp.Details.JobInfo.Program)
	require.NoError(t, err)
	return createResp, program
}

func TestInspectJob(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c, fc := setupTest(t, func(env *Env) {
			env.GetPermissionser = &testPermitter{mode: permitterAllow}
		})
		ctx := pctx.TestContext(t)
		programFileset := createFileset(t, fc, map[string][]byte{
			"file/path": []byte(`!#/bin/bash; ls /input/;`),
		})
		inputFileset := createFileset(t, fc, map[string][]byte{
			"a.txt": []byte(`dummy input`),
		})
		var createJobResp *pjs.CreateJobResponse
		req := &pjs.CreateJobRequest{
			Program: programFileset.HexString(),
			Input:   []string{inputFileset.HexString()},
		}
		createJobResp, programFileset = createJob(ctx, t, c, req)
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
		checkEqualHandleIDs(t, programFileset.HexString(), jobInfo.Program)
		require.Equal(t, len(jobInfo.Input), 1)
		checkEqualHandleIDs(t, inputFileset.HexString(), jobInfo.Input[0])
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

func checkEqualHandleIDs(t *testing.T, handleStr1, handleStr2 string) {
	handle1, err := fileset.ParseHandleKeepID(handleStr1)
	require.NoError(t, err)
	handle2, err := fileset.ParseHandleKeepID(handleStr2)
	require.NoError(t, err)
	require.Equal(t, handle1.ID(), handle2.ID())
}

// TestRunJob tests processQueue, AwaitJob, InspectJob as a whole
func TestRunJob(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		c, fc := setupTest(t)
		ctx := pctx.TestContext(t)
		programFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		inputFileset := createFileset(t, fc, map[string][]byte{
			"a.txt": []byte("dummy input"),
		})
		in := &pjs.CreateJobRequest{
			Program: programFileset.HexString(),
			Input:   []string{inputFileset.HexString()},
		}

		out, err := runJob(t, ctx, c, fc, in, func(resp *pjs.ProcessQueueResponse) error {
			// for now we do nothing here, simply return the input filesets
			resp.Input = in.Input
			return nil
		})
		require.NoError(t, err)
		for i, input := range in.Input {
			checkEqualHandleIDs(t, input, out.success.Output[i])
		}
	})
	t.Run("run job callback fails", func(t *testing.T) {
		c, fc := setupTest(t)
		ctx := pctx.TestContext(t)
		programFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		inputFileset := createFileset(t, fc, map[string][]byte{
			"a.txt": []byte("dummy input"),
		})
		in := &pjs.CreateJobRequest{
			Program: programFileset.HexString(),
			Input:   []string{inputFileset.HexString()},
		}
		out, err := runJob(t, ctx, c, fc, in, func(resp *pjs.ProcessQueueResponse) error {
			// for now we do nothing here, simply return the input filesets
			resp.Input = in.Input
			return errors.New("intentional error to fail the callback")
		})
		require.NoError(t, err)
		require.Equal(t, pjs.JobErrorCode_FAILED, out.errorCode)
	})
}

func TestDeleteJob(t *testing.T) {
	type setup struct {
		ctx     context.Context
		c       pjs.APIClient
		fc      storage.FilesetClient
		id      *pjs.Job
		program *fileset.Handle
	}
	setupJob := func() (s setup) {
		s.ctx = pctx.TestContext(t)
		s.c, s.fc = setupTest(t)
		s.program = createFileset(t, s.fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		req := &pjs.CreateJobRequest{
			Program: s.program.HexString(),
			Input:   []string{s.program.HexString()},
		}
		var resp *pjs.CreateJobResponse
		resp, s.program = createJob(s.ctx, t, s.c, req)
		s.id = resp.Id
		return s
	}
	attemptDelete := func(s setup) {
		_, err := s.c.DeleteJob(s.ctx, &pjs.DeleteJobRequest{Job: s.id})
		require.NoError(t, err)
		_, err = s.c.InspectJob(s.ctx, &pjs.InspectJobRequest{Job: s.id})
		require.YesError(t, err)
		require.Equal(t, codes.NotFound, status.Convert(err).Code())
	}
	t.Run("valid/single/queued", func(t *testing.T) {
		s := setupJob()
		attemptDelete(s)
	})
	t.Run("valid/single/done", func(t *testing.T) {
		s := setupJob()
		_, err := runJobFrom(t, s.ctx, s.c, s.fc, s.id.Id, s.program, func(resp *pjs.ProcessQueueResponse) error {
			return nil
		})
		require.NoError(t, err)
		attemptDelete(s)
	})
	t.Run("valid/single/cancelled", func(t *testing.T) {
		s := setupJob()
		_, err := runJobFrom(t, s.ctx, s.c, s.fc, s.id.Id, s.program, func(resp *pjs.ProcessQueueResponse) error {
			_, err := s.c.CancelJob(s.ctx, &pjs.CancelJobRequest{
				Job: &pjs.Job{
					Id: s.id.Id,
				},
			})
			require.NoError(t, err)
			return nil
		})
		require.NoError(t, err)
		attemptDelete(s)
	})
	t.Run("valid/tree/done", func(t *testing.T) {
		s := setupJob()
		eg, egCtx := errgroup.WithContext(s.ctx)
		parentCtx := make(chan string)
		childDone := make(chan struct{})
		eg.Go(func() error {
			// runJobFrom blocks, so it needs to be run in a separate goroutine.
			_, err := runJobFrom(t, egCtx, s.c, s.fc, s.id.Id, s.program, func(resp *pjs.ProcessQueueResponse) error {
				parentCtx <- resp.Context
				<-childDone // wait for the child job to finish before parent can finish.
				return nil
			})
			require.NoError(t, err)
			return nil
		})
		jCtx := <-parentCtx // block on context being ready to read.
		childJob, err := s.c.CreateJob(s.ctx, &pjs.CreateJobRequest{Context: jCtx, Program: s.program.HexString(), Input: []string{s.program.HexString()}})
		require.NoError(t, err)
		_, err = runJobFrom(t, s.ctx, s.c, s.fc, childJob.Id.Id, s.program, func(resp *pjs.ProcessQueueResponse) error {
			close(childDone) // mark child job as done.
			return nil
		})
		require.NoError(t, err)
		require.NoError(t, eg.Wait())
		attemptDelete(s)
		// also confirm the child job was deleted.
		_, err = s.c.InspectJob(s.ctx, &pjs.InspectJobRequest{Job: childJob.Id})
		require.YesError(t, err)
		require.Equal(t, codes.NotFound, status.Convert(err).Code())
	})
	t.Run("mixed/single/processing", func(t *testing.T) {
		s := setupJob()
		eg, egCtx := errgroup.WithContext(s.ctx)
		processing, done := make(chan struct{}), make(chan struct{})
		eg.Go(func() error {
			// runJobFrom blocks so it must be run in another goroutine.
			_, err := runJobFrom(t, egCtx, s.c, s.fc, s.id.Id, s.program, func(resp *pjs.ProcessQueueResponse) error {
				close(processing) //at this point, the job is in the 'processing' job state.
				<-done            // wait for main thread to tell job to finish.
				return nil
			})
			return err
		})
		<-processing // wait for processing job state before attempting deletion.
		_, err := s.c.DeleteJob(s.ctx, &pjs.DeleteJobRequest{Job: s.id})
		require.YesError(t, err)
		require.Equal(t, codes.FailedPrecondition, status.Convert(err).Code())
		close(done) // tell the job it can finish.
		require.NoError(t, eg.Wait())
		_, err = s.c.InspectJob(s.ctx, &pjs.InspectJobRequest{Job: s.id})
		require.NoError(t, err)
		// test delete works after everything is done.
		attemptDelete(s)
	})
}

func TestCancelJob(t *testing.T) {
	t.Run("cancel job before processing", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		programFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		req := &pjs.CreateJobRequest{
			Program: programFileset.HexString(),
			Input:   []string{programFileset.HexString()},
		}
		resp, programFileset := createJob(ctx, t, c, req)
		jid := resp.Id.Id
		_, err := c.CancelJob(ctx, &pjs.CancelJobRequest{
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
		id := programFileset.ID()
		_, err = c.InspectQueue(ctx, &pjs.InspectQueueRequest{
			Queue: &pjs.Queue{
				Id: id[:],
			},
		})
		require.YesError(t, err)
	})
	t.Run("cancel a processing job", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)

		program := createFileset(t, fc, map[string][]byte{"program": []byte("foo")})
		req := &pjs.CreateJobRequest{
			Program: program.HexString(),
			Input:   []string{program.HexString()},
		}
		var createResp *pjs.CreateJobResponse
		createResp, program = createJob(ctx, t, c, req)

		cancel := func() {
			_, err := c.CancelJob(ctx, &pjs.CancelJobRequest{
				Job: &pjs.Job{
					Id: createResp.Id.Id,
				},
			})
			require.NoError(t, err)
		}

		_, err := runJobFrom(t, ctx, c, fc, createResp.Id.Id, program, func(resp *pjs.ProcessQueueResponse) error {
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
		programA := createFileset(t, fc, map[string][]byte{"program": []byte("A")})
		req := &pjs.CreateJobRequest{
			Program: programA.HexString(),
			Input:   []string{programA.HexString()},
		}
		var createRespA *pjs.CreateJobResponse
		createRespA, programA = createJob(ctx, t, c, req)
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

		programB := createFileset(t, fc, map[string][]byte{"program": []byte("B")})
		req = &pjs.CreateJobRequest{
			Program: programB.HexString(),
			Input:   []string{programB.HexString()},
			Context: jobAContext,
		}
		var createRespB *pjs.CreateJobResponse
		createRespB, programB = createJob(ctx, t, c, req)
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

		programC := createFileset(t, fc, map[string][]byte{"program": []byte("C")})
		req = &pjs.CreateJobRequest{
			Program: programC.HexString(),
			Input:   []string{programC.HexString()},
			Context: jobBContext,
		}
		var createRespC *pjs.CreateJobResponse
		createRespC, programC = createJob(ctx, t, c, req)
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
		_, err := c.CancelJob(ctx, &pjs.CancelJobRequest{
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

type jobInfoResult struct {
	success   *pjs.JobInfo_Success
	errorCode pjs.JobErrorCode
}

// runJob does work through PJS.
func runJob(t *testing.T, ctx context.Context, c pjs.APIClient, fc storage.FilesetClient, req *pjs.CreateJobRequest,
	fn func(resp *pjs.ProcessQueueResponse) error) (*jobInfoResult, error) {
	jres, handle := createJob(ctx, t, c, req)
	return runJobFrom(t, ctx, c, fc, jres.Id.Id, handle, fn)
}

func runJobFrom(t *testing.T, ctx context.Context, c pjs.APIClient, fc storage.FilesetClient, from int64, programHandle *fileset.Handle,
	fn func(resp *pjs.ProcessQueueResponse) error) (*jobInfoResult, error) {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ctx, cf := context.WithCancel(ctx)
		defer cf()
		pqc, err := c.ProcessQueue(ctx)
		require.NoError(t, err)
		id := programHandle.ID()
		err = processQueue(pqc, id[:], fn)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if status.Code(err) == codes.Canceled {
			err = nil
		}
		return err
	})
	var ret = &jobInfoResult{}
	eg.Go(func() error {
		jobInfo, err := await(ctx, c, from)
		if err != nil {
			return err
		}
		ret.success = jobInfo.GetSuccess()
		ret.errorCode = jobInfo.GetError()
		cf() // success, cancel the other gorountine
		return nil
	})
	require.NoError(t, eg.Wait())
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
			// if send returns an error, the ctx on the server side will be cancelled. Job cannot transit to
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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		_, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
			Program: testFileset.HexString(),
			Input:   []string{testFileset.HexString()},
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

func TestInspectQueue(t *testing.T) {
	t.Run("empty queue", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, _ := setupTest(t)
		_, err := c.InspectQueue(ctx, &pjs.InspectQueueRequest{
			Queue: &pjs.Queue{
				Id: []byte(`dummyHash`),
			},
		})
		require.YesError(t, err)
	})
	t.Run("multiple queues", func(t *testing.T) {
		ctx := pctx.TestContext(t)
		c, fc := setupTest(t)
		programFileset1 := createFileset(t, fc, map[string][]byte{
			"file": []byte(`program1`),
		})
		programFileset2 := createFileset(t, fc, map[string][]byte{
			"file": []byte(`program2`),
		})
		programFileset3 := createFileset(t, fc, map[string][]byte{
			"file": []byte(`program3`),
		})
		inspect := func(hash []byte, i int) {
			inspectQueueResp1, err := c.InspectQueue(ctx, &pjs.InspectQueueRequest{
				Queue: &pjs.Queue{
					Id: hash,
				},
			})
			require.NoError(t, err)
			require.Equal(t, hash, inspectQueueResp1.Details.QueueInfo.Queue.Id)
			handle, err := fileset.ParseHandleKeepID(inspectQueueResp1.Details.QueueInfo.Program)
			require.NoError(t, err)
			id := handle.ID()
			require.Equal(t, hash, id[:])
			require.Equal(t, int64(i), inspectQueueResp1.Details.Size)
		}
		for i := 1; i <= 10; i++ {
			req := &pjs.CreateJobRequest{
				Program: programFileset1.HexString(),
				Input:   []string{programFileset1.HexString()},
			}
			_, program := createJob(ctx, t, c, req)
			id := program.ID()
			inspect(id[:], i)

			req = &pjs.CreateJobRequest{
				Program: programFileset2.HexString(),
				Input:   []string{programFileset2.HexString()},
			}
			_, program = createJob(ctx, t, c, req)
			id = program.ID()
			inspect(id[:], i)

			req = &pjs.CreateJobRequest{
				Program: programFileset3.HexString(),
				Input:   []string{programFileset3.HexString()},
			}
			_, program = createJob(ctx, t, c, req)
			id = program.ID()
			inspect(id[:], i)
		}
	})
}

func TestListJob(t *testing.T) {
	t.Run("list jobs given parents", func(t *testing.T) {
		c, fc := setupTest(t)
		ctx := pctx.TestContext(t)
		depth := 3
		fullBinaryJobTree(t, ctx, depth, c, fc)

		for jid := 1; jid <= 3; jid++ {
			expected := []int64{int64(jid * 2), int64(jid*2 + 1)}
			listC, err := c.ListJob(ctx, &pjs.ListJobRequest{
				Job: &pjs.Job{
					Id: int64(jid),
				},
			})
			require.NoError(t, err)
			var actual []int64
			for {
				resp, err := listC.Recv()
				if err != nil {
					if errors.As(err, io.EOF) {
						break
					}
					require.NoError(t, err)
				}
				actual = append(actual, resp.Id.Id)
			}
			require.NoDiff(t, expected, actual, nil)
		}
	})
	t.Run("list no parent jobs", func(t *testing.T) {
		c, fc := setupTest(t)
		ctx := pctx.TestContext(t)
		depth := 3
		fullBinaryJobTree(t, ctx, depth, c, fc)
		// list job request without giving parent job
		listC, err := c.ListJob(ctx, &pjs.ListJobRequest{})
		expected := []int64{1}
		require.NoError(t, err)
		var actual []int64
		for {
			resp, err := listC.Recv()
			if err != nil {
				if errors.As(err, io.EOF) {
					break
				}
				require.NoError(t, err)
			}
			actual = append(actual, resp.Id.Id)
		}
		require.NoDiff(t, expected, actual, nil)
	})
}

func TestListQueue(t *testing.T) {
	ctx := pctx.TestContext(t)
	c, fc := setupTest(t)
	inputFileset := createFileset(t, fc, map[string][]byte{
		"a.txt": []byte(`dummy input`),
	})
	programFileset1 := createFileset(t, fc, map[string][]byte{
		"file": []byte(`program1`),
	})
	req1 := &pjs.CreateJobRequest{
		Program: programFileset1.HexString(),
		Input:   []string{inputFileset.HexString()},
	}
	programFileset2 := createFileset(t, fc, map[string][]byte{
		"file": []byte(`program2`),
	})
	req2 := &pjs.CreateJobRequest{
		Program: programFileset2.HexString(),
		Input:   []string{inputFileset.HexString()},
	}
	programFileset3 := createFileset(t, fc, map[string][]byte{
		"file": []byte(`program3`),
	})
	// ListQueue should return QueueIDs and lengths based on all jobs in all states.
	// Length should come from counting only jobs in QUEUED state.
	for i := 1; i <= 10; i++ {
		// queue1 and queue2 has no queued jobs.
		_, err := runJob(t, ctx, c, fc, req1, func(resp *pjs.ProcessQueueResponse) error {
			return nil
		})
		require.NoError(t, err)
		_, err = runJob(t, ctx, c, fc, req2, func(resp *pjs.ProcessQueueResponse) error {
			return nil
		})
		require.NoError(t, err)

		req := &pjs.CreateJobRequest{
			Program: programFileset3.HexString(),
			Input:   []string{inputFileset.HexString()},
		}
		_, programFileset3 = createJob(ctx, t, c, req)
	}
	id3 := programFileset3.ID()
	listC, err := c.ListQueue(ctx, &pjs.ListQueueRequest{})
	require.NoError(t, err)
	for {
		resp, err := listC.Recv()
		if err != nil {
			if errors.As(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
		require.NoError(t, err)
		if !idHashEqual(resp.Id.Id, id3[:]) {
			require.Equal(t, int64(0), resp.Details.Size)
		} else {
			require.Equal(t, int64(10), resp.Details.Size)
		}
	}
}

func idHashEqual(a []byte, b []byte) bool {
	return string(a) == string(b)
}

func setupTest(t testing.TB, opts ...ClientOptions) (pjs.APIClient, storage.FilesetClient) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	server := storagesrv.NewTestServer(t, db)
	client := storagesrv.NewTestFilesetClient(t, server)
	return NewTestClient(t, db, server, opts...), client
}

func createFileset(t testing.TB, fc storage.FilesetClient, files map[string][]byte) *fileset.Handle {
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
	handle, err := fileset.ParseHandleKeepID(resp.FilesetId)
	require.NoError(t, err)
	return handle
}

func fullBinaryJobTree(t *testing.T, ctx context.Context, maxDepth int, c pjs.APIClient, fc storage.FilesetClient) {
	if maxDepth == 0 {
		return
	}

	// create node at depth == 1
	program := createFileset(t, fc, map[string][]byte{"program": []byte("foo")})
	req := &pjs.CreateJobRequest{
		Program: program.HexString(),
		Input:   []string{program.HexString()},
	}
	var createResp *pjs.CreateJobResponse
	createResp, program = createJob(ctx, t, c, req)
	var processResp *pjs.ProcessQueueResponse
	_, err := runJobFrom(t, ctx, c, fc, createResp.Id.Id, program, func(resp *pjs.ProcessQueueResponse) error {
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
				prog := createFileset(t, fc, map[string][]byte{"program": []byte(rand.String(32))})
				req := &pjs.CreateJobRequest{
					Context: p.Context,
					Program: prog.HexString(),
					Input:   []string{prog.HexString()},
				}
				var cResp *pjs.CreateJobResponse
				cResp, prog = createJob(ctx, t, c, req)
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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset.HexString(), Input: []string{testFileset.HexString()}})
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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job as admin.
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset.HexString(), Input: []string{testFileset.HexString()}})
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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job as admin.
		req := &pjs.CreateJobRequest{
			Program: testFileset.HexString(),
			Input:   []string{testFileset.HexString()},
		}
		var jobResp *pjs.CreateJobResponse
		jobResp, testFileset = createJob(ctx, t, c, req)
		var jobContext string
		// run process queue to get a job context token.
		_, err := runJobFrom(t, ctx, c, fc, jobResp.Id.Id, testFileset, func(resp *pjs.ProcessQueueResponse) error {
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
		jobResp, err = c.CreateJob(ctx, &pjs.CreateJobRequest{Context: jobContext, Program: testFileset.HexString(), Input: []string{testFileset.HexString()}})
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
		testFileset := createFileset(t, fc, map[string][]byte{
			"file": []byte(`!#/bin/bash; ls /input/;`),
		})
		// create one job without ctx as admin.
		jobResp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{Program: testFileset.HexString(), Input: []string{testFileset.HexString()}})
		require.NoError(t, err)
		modes := []permitterEnum{permitterAllow, permitterDeny}
		jobCtx := "00001111222233334444555566666777788889999aaaabbbbccccddddeeeefff"
		for _, mode := range modes {
			p.mode = mode
			_, err = c.CreateJob(ctx, &pjs.CreateJobRequest{Context: jobCtx, Program: testFileset.HexString(), Input: []string{testFileset.HexString()}})
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

func TestJobCaching(t *testing.T) {
	ctx := pctx.TestContext(t)
	c, fc := setupTest(t)
	inputFileset := createFileset(t, fc, map[string][]byte{
		"a.txt": []byte(`dummy input`),
	})
	programFileset1 := createFileset(t, fc, map[string][]byte{
		"file": []byte(`program1`),
	})
	req1 := &pjs.CreateJobRequest{
		Program:    programFileset1.HexString(),
		Input:      []string{inputFileset.HexString()},
		CacheWrite: true,
		CacheRead:  true,
	}
	var jobCtx string
	_, err := runJob(t, ctx, c, fc, req1, func(resp *pjs.ProcessQueueResponse) error {
		jobCtx = resp.Context
		return nil
	})
	require.NoError(t, err)
	// job2 should be created from the cache.
	job2, err := c.CreateJob(ctx, req1)
	require.NoError(t, err)
	resp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{Job: job2.Id})
	require.NoError(t, err)
	require.Equal(t, pjs.JobState_DONE, resp.Details.JobInfo.State)
	// now, delete the first job. The second job should be used for the cache.
	_, err = c.DeleteJob(ctx, &pjs.DeleteJobRequest{Context: jobCtx})
	require.NoError(t, err)
	_, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Context: jobCtx})
	require.YesError(t, err)
	require.Equal(t, codes.NotFound, status.Convert(err).Code())
	// job3 should be created using job2.
	job3, err := c.CreateJob(ctx, req1)
	require.NoError(t, err)
	resp, err = c.InspectJob(ctx, &pjs.InspectJobRequest{Job: job3.Id})
	require.NoError(t, err)
	require.Equal(t, pjs.JobState_DONE, resp.Details.JobInfo.State)

}
