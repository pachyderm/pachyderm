package server

import (
	"context"
	"time"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"go.pedge.io/lion/proto"
	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/time"
)

var (
	// DefaultGCPolicy is the default GC policy used by a pipeline if one is not
	// specified.
	DefaultGCPolicy = &ppsclient.GCPolicy{
		// a day
		Success: &google_protobuf.Duration{
			Seconds: 24 * 60 * 60,
		},
		// a week
		Failure: &google_protobuf.Duration{
			Seconds: 7 * 24 * 60 * 60,
		},
	}
)

func (a *apiServer) runGC(ctx context.Context, pipelineInfo *ppsclient.PipelineInfo) {
	successTick := time.Tick(prototime.DurationFromProto(pipelineInfo.GcPolicy.Success))
	failureTick := time.Tick(prototime.DurationFromProto(pipelineInfo.GcPolicy.Failure))
	// wait blocks until it's time to run GC again
	wait := func() {
		select {
		case <-successTick:
		case <-failureTick:
		}
	}
	for {
		client, err := a.getPersistClient()
		if err != nil {
			protolion.Errorf("error getting persist client: %s", err)
			wait()
			continue
		}
		jobIDs, err := client.ListGCJobs(ctx, &persist.ListGCJobsRequest{
			PipelineName: pipelineInfo.Pipeline.Name,
			GcPolicy:     pipelineInfo.GcPolicy,
		})
		if err != nil {
			protolion.Errorf("error listing jobs to GC: %s", err)
			wait()
			continue
		}
		for _, jobID := range jobIDs.Jobs {
			jobID := jobID
			go func() {
				jobInfo, err := client.InspectJob(ctx, &ppsclient.InspectJobRequest{
					Job: &ppsclient.Job{ID: jobID},
				})
				if err != nil {
					protolion.Errorf("error deleting job: %s", err)
					return
				}
				if err := a.deleteJob(ctx, jobInfo); err != nil {
					protolion.Errorf("error deleting job: %s", err)
					return
				}
				if _, err := client.GCJob(ctx, &ppsclient.Job{jobID}); err != nil {
					protolion.Errorf("error marking job %s as GC-ed: %s", jobID, err)
				}
				return
			}()
		}
		wait()
	}
	panic("unreachable")
}
