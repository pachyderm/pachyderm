package pjs

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		c, fc := setupTest(t)
		resp := createJob(t, c, fc)
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
	out, err := runJob(t, ctx, c, in, func(inputs []string) ([]string, error) {
		// todo(muyang)
		// for now we do nothing here, simply return the input filesets

		return inputs, nil
	})
	require.NoError(t, err)
	require.Equal(t, in.Input, out.Output)
}

// runJob does work through PJS.
func runJob(t *testing.T, ctx context.Context, s pjs.APIClient, in *pjs.CreateJobRequest, fn func(inputs []string) ([]string, error)) (*pjs.JobInfo_Success, error) {
	jres, err := s.CreateJob(ctx, in)
	require.NoError(t, err)
	program, err := fileset.ParseID(in.Program)
	require.NoError(t, err)
	programHash := []byte(program.HexString())
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		err := processQueue(ctx, s, programHash, fn)
		if status.Code(err) == codes.Canceled {
			err = nil
		}
		return err
	})
	var ret *pjs.JobInfo_Success
	eg.Go(func() error {
		jobInfo, err := await(ctx, s, jres.Id.Id)
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

func processQueue(ctx context.Context, s pjs.APIClient, programHash []byte, fn func([]string) ([]string, error)) error {
	ctx, cf := context.WithCancel(ctx)
	defer cf()
	pqc, err := s.ProcessQueue(ctx)
	if err != nil {
		return err
	}
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
		out, err := fn(msg.Input)
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
						Output: out,
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

func createJob(t *testing.T, c pjs.APIClient, fc storage.FilesetClient) *pjs.CreateJobResponse {
	ctx := pctx.TestContext(t)
	programFileset := createFileSet(t, fc, map[string][]byte{
		"file": []byte(`!#/bin/bash; ls /input/;`),
	})
	inputFileset := createFileSet(t, fc, map[string][]byte{
		"a.txt": []byte("dummy input"),
	})
	resp, err := c.CreateJob(ctx, &pjs.CreateJobRequest{
		Program: programFileset,
		Input:   []string{inputFileset},
	})
	require.NoError(t, err)
	return resp
}
