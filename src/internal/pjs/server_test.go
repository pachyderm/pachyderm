package pjs

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
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
		resp := createJob(t, ctx, c, fc)
		t.Log(resp)
	})
}

func TestInspectJob(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		c, fc := setupTest(t)
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
		t.Log(createJobResp)
		inspectJobResp, err := c.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: createJobResp.Id,
		})
		require.NoError(t, err)
		t.Log(inspectJobResp)
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

	out, err := runJob(t, ctx, c, in, func(resp *pjs.ProcessQueueResponse) error {
		// todo(muyang)
		// for now we do nothing here, simply return the input filesets
		resp.Input = in.Input
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, in.Input, out.Output)
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
func runJob(t *testing.T, ctx context.Context, c pjs.APIClient, in *pjs.CreateJobRequest,
	fn func(resp *pjs.ProcessQueueResponse) error) (*pjs.JobInfo_Success, error) {
	jres, err := c.CreateJob(ctx, in)
	require.NoError(t, err)
	return runJobFrom(t, ctx, c, jres.Id.Id, in.Program, fn)
}

func runJobFrom(t *testing.T, ctx context.Context, c pjs.APIClient, from int64, programStr string,
	fn func(resp *pjs.ProcessQueueResponse) error) (*pjs.JobInfo_Success, error) {
	program, err := fileset.ParseID(programStr)
	require.NoError(t, err)
	programHash := []byte(program.HexString())
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ctx, cf := context.WithCancel(ctx)
		defer cf()
		pqc, err := c.ProcessQueue(ctx)
		require.NoError(t, err)
		err = processQueue(pqc, programHash, fn)
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
	for {
		res, err := s.InspectJob(ctx, &pjs.InspectJobRequest{
			Job: &pjs.Job{Id: jid},
		})
		if err != nil {
			return nil, err
		}
		if res.Details.JobInfo.State == pjs.JobState_DONE {
			return res.Details.JobInfo, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func setupTest(t testing.TB) (pjs.APIClient, storage.FilesetClient) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewTestDB(t)
	migrationEnv := migrations.Env{EtcdClient: testetcd.NewEnv(ctx, t).EtcdClient}
	require.NoError(t, migrations.ApplyMigrations(ctx, db, migrationEnv, clusterstate.DesiredClusterState), "should be able to set up tables")
	return NewTestClient(t, db), storagesrv.NewTestFilesetClient(t, db)
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

type createJobOpts func(req *pjs.CreateJobRequest)

func createJob(t *testing.T, ctx context.Context, c pjs.APIClient, fc storage.FilesetClient,
	opts ...createJobOpts) *pjs.CreateJobResponse {
	programFileset := createFileSet(t, fc, map[string][]byte{
		"file": []byte(`!#/bin/bash; ls /input/;`),
	})
	inputFileset := createFileSet(t, fc, map[string][]byte{
		"a.txt": []byte("dummy input"),
	})
	req := &pjs.CreateJobRequest{
		Program: programFileset,
		Input:   []string{inputFileset},
	}
	for _, opt := range opts {
		opt(req)
	}
	resp, err := c.CreateJob(ctx, req)
	require.NoError(t, err)
	return resp
}

func fullBinaryJobTree(t *testing.T, ctx context.Context, maxDepth int, c pjs.APIClient, fc storage.FilesetClient) {
	if maxDepth == 0 {
		return
	}

	// create node at depth == 1
	program := createFileSet(t, fc, map[string][]byte{"program": []byte("foo")})
	createResp := createJob(t, ctx, c, fc, func(req *pjs.CreateJobRequest) {
		req.Program = program
	})
	var processResp *pjs.ProcessQueueResponse
	_, err := runJobFrom(t, ctx, c, createResp.Id.Id, program, func(resp *pjs.ProcessQueueResponse) error {
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
				cResp := createJob(t, ctx, c, fc, func(req *pjs.CreateJobRequest) {
					req.Context = p.Context
					req.Program = prog
				})
				var pResp *pjs.ProcessQueueResponse
				_, err = runJobFrom(t, ctx, c, cResp.Id.Id, prog, func(resp *pjs.ProcessQueueResponse) error {
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
