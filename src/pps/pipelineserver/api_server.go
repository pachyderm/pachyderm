package pipelineserver

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
)

type apiServer struct {
	protorpclog.Logger
	pfsAPIClient     pfs.APIClient
	jobAPIClient     pps.JobAPIClient
	persistAPIClient persist.APIClient

	started                          bool
	pipelineNameToPipelineController map[string]*pipelineController
	lock                             *sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	persistAPIClient persist.APIClient,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.PipelineAPI"),
		pfsAPIClient,
		jobAPIClient,
		persistAPIClient,
		false,
		make(map[string]*pipelineController),
		&sync.Mutex{},
	}
}

func (a *apiServer) Start() error {
	a.lock.Lock()
	defer a.lock.Unlock()
	// TODO(pedge): volatile bool?
	if a.started {
		// TODO(pedge): abstract error to public variable
		return errors.New("pachyderm.pps.pipelineserver: already started")
	}
	a.started = true
	pipelineInfos, err := a.ListPipeline(context.Background(), &pps.ListPipelineRequest{})
	if err != nil {
		return err
	}
	for _, pipelineInfo := range pipelineInfos.PipelineInfo {
		if err := a.addPipelineController(pipelineInfo); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfo := &persist.PipelineInfo{
		PipelineName: request.Pipeline.Name,
		Transform:    request.Transform,
		Input:        request.Input,
		Output:       request.Output,
	}
	if _, err := a.persistAPIClient.CreatePipelineInfo(ctx, persistPipelineInfo); err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	if err := a.addPipelineController(persistPipelineInfoToPipelineInfo(persistPipelineInfo)); err != nil {
		// TODO(pedge): proper create/rollback (do not commit transaction with create)
		if _, rollbackErr := a.persistAPIClient.DeletePipelineInfo(ctx, request.Pipeline); rollbackErr != nil {
			return nil, fmt.Errorf("%v", err, rollbackErr)
		}
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectPipeline(ctx context.Context, request *pps.InspectPipelineRequest) (response *pps.PipelineInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfo, err := a.persistAPIClient.GetPipelineInfo(ctx, request.Pipeline)
	if err != nil {
		return nil, err
	}
	return persistPipelineInfoToPipelineInfo(persistPipelineInfo), nil
}

func (a *apiServer) ListPipeline(ctx context.Context, request *pps.ListPipelineRequest) (response *pps.PipelineInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelineInfos, err := a.persistAPIClient.ListPipelineInfos(ctx, google_protobuf.EmptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineInfos := make([]*pps.PipelineInfo, len(persistPipelineInfos.PipelineInfo))
	for i, persistPipelineInfo := range persistPipelineInfos.PipelineInfo {
		pipelineInfos[i] = persistPipelineInfoToPipelineInfo(persistPipelineInfo)
	}
	return &pps.PipelineInfos{
		PipelineInfo: pipelineInfos,
	}, nil
}

func (a *apiServer) DeletePipeline(ctx context.Context, request *pps.DeletePipelineRequest) (response *google_protobuf.Empty, err error) {
	if _, err := a.persistAPIClient.DeletePipelineInfo(ctx, request.Pipeline); err != nil {
		return nil, err
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	if err := a.removePipelineController(request.Pipeline.Name); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func persistPipelineInfoToPipelineInfo(persistPipelineInfo *persist.PipelineInfo) *pps.PipelineInfo {
	return &pps.PipelineInfo{
		Pipeline: &pps.Pipeline{
			Name: persistPipelineInfo.PipelineName,
		},
		Transform: persistPipelineInfo.Transform,
		Input:     persistPipelineInfo.Input,
		Output:    persistPipelineInfo.Output,
	}
}

func (a *apiServer) addPipelineController(pipelineInfo *pps.PipelineInfo) error {
	if _, ok := a.pipelineNameToPipelineController[pipelineInfo.Pipeline.Name]; ok {
		protolog.Warnf("pachyderm.pps.pipelineserver: had a create change event for an existing pipeline: %v", pipelineInfo)
		if err := a.removePipelineController(pipelineInfo.Pipeline.Name); err != nil {
			return err
		}
	}
	pipelineController := newPipelineController(
		a.pfsAPIClient,
		a.jobAPIClient,
		pps.NewLocalPipelineAPIClient(a),
		pipelineInfo,
	)
	a.pipelineNameToPipelineController[pipelineInfo.Pipeline.Name] = pipelineController
	return pipelineController.Start()
}

func (a *apiServer) removePipelineController(name string) error {
	pipelineController, ok := a.pipelineNameToPipelineController[name]
	if !ok {
		protolog.Warnf("pachyderm.pps.pipelineserver: no pipeline registered for name: %s", name)
	} else {
		pipelineController.Cancel()
	}
	return nil
}
