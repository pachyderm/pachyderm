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

const (
	changeEventTypeCreate changeEventType = iota
	changeEventTypeDelete
)

type changeEventType int

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
	if _, err := a.persistAPIClient.CreatePipelineInfo(
		ctx,
		&persist.PipelineInfo{
			PipelineName: request.Pipeline.Name,
			Transform:    request.Transform,
			Input:        request.Input,
			Output:       request.Output,
		},
	); err != nil {
		return nil, err
	}
	if err := a.registerChangeEvent(
		ctx,
		&changeEvent{
			Type:         changeEventTypeCreate,
			PipelineName: request.Pipeline.Name,
		},
	); err != nil {
		// TODO(pedge): need to roll back the db create
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
	if err := a.registerChangeEvent(
		ctx,
		&changeEvent{
			Type:         changeEventTypeDelete,
			PipelineName: request.Pipeline.Name,
		},
	); err != nil {
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

type changeEvent struct {
	Type         changeEventType
	PipelineName string
}

// TODO(pedge): this is relateively out of date, we can just do this directly in the functions, and
// with the create at least, we avoid a db read
func (a *apiServer) registerChangeEvent(ctx context.Context, request *changeEvent) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	switch request.Type {
	case changeEventTypeCreate:
		if !a.pipelineRegistered(request.PipelineName) {
			pipelineInfo, err := a.InspectPipeline(
				ctx,
				&pps.InspectPipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: request.PipelineName,
					},
				},
			)
			if err != nil {
				return err
			}
			if err := a.addPipelineController(pipelineInfo); err != nil {
				return err
			}
			// TODO(pedge): what to do?
		} else {
			protolog.Warnf("pachyderm.pps.pipelineserver: had a create change event for an existing pipeline: %v", request)
			if err := a.removePipelineController(request.PipelineName); err != nil {
				return err
			}
			pipelineInfo, err := a.InspectPipeline(
				ctx,
				&pps.InspectPipelineRequest{
					Pipeline: &pps.Pipeline{
						Name: request.PipelineName,
					},
				},
			)
			if err != nil {
				return err
			}
			if err := a.addPipelineController(pipelineInfo); err != nil {
				return err
			}
		}
	case changeEventTypeDelete:
		if !a.pipelineRegistered(request.PipelineName) {
			protolog.Warnf("pachyderm.pps.pipelineserver: had a delete change event for a pipeline that was not registered: %v", request)
		} else {
			if err := a.removePipelineController(request.PipelineName); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("pachyderm.pps.pipelineserver: unknown change event type: %v", request.Type)
	}
	return nil
}

func (a *apiServer) pipelineRegistered(name string) bool {
	_, ok := a.pipelineNameToPipelineController[name]
	return ok
}

func (a *apiServer) addPipelineController(pipelineInfo *pps.PipelineInfo) error {
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
		return fmt.Errorf("pachyderm.pps.pipelineserver: no pipeline registered for name: %s", name)
	}
	pipelineController.Cancel()
	return nil
}
