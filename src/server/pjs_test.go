//go:build k8s

package server

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestPJSAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	t.Parallel()
	c, _ := minikubetestenv.AcquireCluster(t)
	ctx := pctx.TestContext(t)
	createFilesetClient, err := c.FilesetClient.CreateFileset(ctx)
	if err != nil {
		t.Fatalf("unexpected error from create fileset client: %v", err)
	}
	err = createFilesetClient.Send(&storage.CreateFilesetRequest{
		Modification: &storage.CreateFilesetRequest_AppendFile{
			AppendFile: &storage.AppendFile{
				Path: "/program",
				Data: &wrapperspb.BytesValue{
					Value: []byte("program"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error from create fileset client: send: %v", err)
	}
	createFilesetResp, err := createFilesetClient.CloseAndRecv()
	if err != nil {
		t.Fatalf("unexpected error from create fileset client: close and receive: %v", err)
	}
	createJobResp, err := c.PjsAPIClient.CreateJob(ctx, &pjs.CreateJobRequest{
		Input:      []string{createFilesetResp.FilesetId},
		Program:    createFilesetResp.FilesetId,
		CacheRead:  true,
		CacheWrite: true,
	})
	if err != nil {
		t.Fatalf("unexpected error from create job: %v", err)
	}
	listJobsClient, err := c.PjsAPIClient.ListJob(ctx, &pjs.ListJobRequest{Job: createJobResp.Job})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	for {
		resp, err := listJobsClient.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			t.Fatalf("unexpected error from list jobs client: receive %v", err)
		}
		if resp.Job.Id != createJobResp.Job.Id {
			t.Fatalf("job id doesn't match and we only created one job: got: %d, want: %d", resp.Job.Id, createJobResp.Job.Id)
		}
	}
	if err := listJobsClient.CloseSend(); err != nil {
		t.Fatalf("unexpected error from list jobs client: close send: %v", err)
	}
	listQueueClient, err := c.PjsAPIClient.ListQueue(ctx, &pjs.ListQueueRequest{})
	if err != nil {
		t.Fatalf("unexpected error from list queue client: %v", err)
	}
	listQueueResp, err := listQueueClient.Recv()
	if err != nil {
		t.Fatalf("unexpected error from list queue client: receive: %v", err)
	}
	queue := listQueueResp.Queue
	err = listQueueClient.CloseSend()
	if err != nil {
		t.Fatalf("unexpected error from list queue client: close send: %v", err)
	}
	processQueueClient, err := c.PjsAPIClient.ProcessQueue(ctx)
	if err != nil {
		t.Fatalf("unexpected error from process queue client: %v", err)
	}
	if err = processQueueClient.Send(&pjs.ProcessQueueRequest{Queue: queue}); err != nil {
		t.Fatalf("unexpected error from process queue client: send: %v", err)
	}
	if _, err = processQueueClient.Recv(); err != nil {
		t.Fatalf("unexpected error from process queue client: receive: %v", err)
	}
	if err = processQueueClient.Send(&pjs.ProcessQueueRequest{
		Queue: queue,
		Result: &pjs.ProcessQueueRequest_Success_{
			Success: &pjs.ProcessQueueRequest_Success{
				Output: []string{createFilesetResp.FilesetId},
			},
		},
	}); err != nil {
		t.Fatalf("unexpected error from process queue client: send: %v", err)
	}
	if err = processQueueClient.CloseSend(); err != nil {
		t.Fatalf("unexpected error from process queue client: close send: %v", err)
	}
	awaitResp, err := c.PjsAPIClient.AwaitJob(ctx, &pjs.AwaitJobRequest{
		Job:          createJobResp.Job,
		DesiredState: pjs.JobState_DONE,
	})
	if err != nil {
		t.Fatalf("unexpected error calling await job: %v", err)
	}
	if awaitResp.ActualState != pjs.JobState_DONE {
		t.Fatalf("state mismatch: job: %d got: %v want: %v", createJobResp.Job.Id, pjs.JobState_name[int32(awaitResp.ActualState)],
			pjs.JobState_name[int32(pjs.JobState_DONE)])
	}
	inspectResp, err := c.PjsAPIClient.InspectJob(ctx, &pjs.InspectJobRequest{
		Job: createJobResp.Job,
	})
	if err != nil {
		t.Fatalf("unexpected error inspecting job: %v", err)
	}
	if inspectResp.Details.JobInfo.Job.Id != createJobResp.Job.Id {
		t.Fatalf("expect job ids to match: got: %d want: %d ", inspectResp.Details.JobInfo.Job.Id, createJobResp.Job.Id)
	}
	switch inspectResp.Details.JobInfo.Result.(type) {
	case *pjs.JobInfo_Success_:
		break
	default:
		jsonBytes, err := json.Marshal(inspectResp.Details.JobInfo.Result)
		if err != nil {
			t.Fatalf("unexpected result from inspect job: want type *pjs.JobInfo_Success,"+
				"couldn't marshal to json: err: %v", err)
		}
		t.Fatalf("unexpected result from inspect job: got: %v, want type *pjs.JobInfo_Success", string(jsonBytes))
	}
	if _, err = c.PjsAPIClient.DeleteJob(ctx, &pjs.DeleteJobRequest{Job: createJobResp.Job}); err != nil {
		t.Fatalf("unexpected error deleting job: %v", err)
	}
}
