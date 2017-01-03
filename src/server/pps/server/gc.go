package server

import (
	"context"
	"time"

	"k8s.io/kubernetes/pkg/api"

	"github.com/gogo/protobuf/types"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"

	"go.pedge.io/lion/proto"
)

var (
	// DefaultGCPolicy is the default GC policy used by a pipeline if one is not
	// specified.
	DefaultGCPolicy = &ppsclient.GCPolicy{
		// a day
		Success: &types.Duration{
			Seconds: 24 * 60 * 60,
		},
		// a week
		Failure: &types.Duration{
			Seconds: 7 * 24 * 60 * 60,
		},
	}
)

func (a *apiServer) runGC(ctx context.Context, pipelineInfo *ppsclient.PipelineInfo) {
	dSuccess, _ := types.DurationFromProto(pipelineInfo.GcPolicy.Success)
	dFailure, _ := types.DurationFromProto(pipelineInfo.GcPolicy.Failure)
	successTick := time.Tick(dSuccess)
	failureTick := time.Tick(dFailure)
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
			zero := int64(0)
			go func(jobID string) {
				if err := a.kubeClient.Extensions().Jobs(a.namespace).Delete(jobID, &api.DeleteOptions{
					GracePeriodSeconds: &zero,
				}); err != nil {
					// TODO: if the error indicates that the job has already been
					// deleted, just proceed to the next step
					protolion.Errorf("error deleting kubernetes job %s: %s", jobID, err)
					return
				}
				if _, err := client.GCJob(ctx, &ppsclient.Job{jobID}); err != nil {
					protolion.Errorf("error marking job %s as GC-ed: %s", jobID, err)
				}
				return
			}(jobID)
		}
		wait()
	}
	panic("unreachable")
}
