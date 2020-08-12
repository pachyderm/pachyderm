package server

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
	"github.com/pachyderm/pachyderm/src/server/worker/datum"
)

var _ pps.APIServer = &apiServerV2{}

type apiServerV2 struct {
	*apiServer
}

func newAPIServerV2(embeddedServer *apiServer) *apiServerV2 {
	return &apiServerV2{apiServer: embeddedServer}
}

func (a *apiServerV2) InspectDatumV2(ctx context.Context, request *pps.InspectDatumRequest) (response *pps.DatumInfoV2, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.collectDatums(ctx, request.Datum.Job, func(meta *datum.Meta) error {
		if common.DatumIDV2(meta.Inputs) == request.Datum.ID {
			response = convertDatumMetaToInfo(meta)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return response, nil
}

func (a *apiServerV2) ListDatumV2(request *pps.ListDatumRequest, server pps.API_ListDatumV2Server) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.collectDatums(server.Context(), request.Job, func(meta *datum.Meta) error {
		return server.Send(convertDatumMetaToInfo(meta))
	})
}

func convertDatumMetaToInfo(meta *datum.Meta) *pps.DatumInfoV2 {
	di := &pps.DatumInfoV2{
		Datum: &pps.Datum{
			Job: &pps.Job{
				ID: meta.JobID,
			},
			ID: common.DatumIDV2(meta.Inputs),
		},
		State: convertDatumState(meta.State),
		Stats: meta.Stats,
	}
	for _, input := range meta.Inputs {
		di.Data = append(di.Data, input.FileInfo)
	}
	return di
}

// (bryce) this is a bit wonky, but it is necessary based on the dependency graph.
func convertDatumState(state datum.State) pps.DatumState {
	switch state {
	case datum.State_FAILED:
		return pps.DatumState_FAILED
	case datum.State_RECOVERED:
		return pps.DatumState_RECOVERED
	default:
		return pps.DatumState_SUCCESS
	}
}

func (a *apiServerV2) collectDatums(ctx context.Context, job *pps.Job, cb func(*datum.Meta) error) error {
	jobInfo, err := a.InspectJob(ctx, &pps.InspectJobRequest{
		Job: &pps.Job{
			ID: job.ID,
		},
	})
	if err != nil {
		return err
	}
	if jobInfo.StatsCommit == nil {
		return errors.Errorf("job not finished")
	}
	pachClient := a.env.GetPachClient(ctx)
	return datum.NewFileSetIterator(pachClient, jobInfo.StatsCommit.Repo.Name, jobInfo.StatsCommit.ID).Iterate(cb)
}

var errV1NotImplemented = errors.Errorf("v1 method not implemented")

func (a *apiServerV2) ListDatum(_ context.Context, _ *pps.ListDatumRequest) (*pps.ListDatumResponse, error) {
	return nil, errV1NotImplemented
}

func (a *apiServerV2) ListDatumStream(_ *pps.ListDatumRequest, _ pps.API_ListDatumStreamServer) error {
	return errV1NotImplemented
}

func (a *apiServerV2) InspectDatum(_ context.Context, _ *pps.InspectDatumRequest) (*pps.DatumInfo, error) {
	return nil, errV1NotImplemented
}

func (a *apiServerV2) GetLogs(_ *pps.GetLogsRequest, _ pps.API_GetLogsServer) error {
	return errV1NotImplemented
}

func (a *apiServerV2) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (*types.Empty, error) {
	// TODO: catch unsupported features, apply transformations (always enable stats)
	return a.apiServer.CreatePipeline(ctx, request)
}

func (a *apiServerV2) GarbageCollect(_ context.Context, _ *pps.GarbageCollectRequest) (*pps.GarbageCollectResponse, error) {
	return nil, errV1NotImplemented
}
