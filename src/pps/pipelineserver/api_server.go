package pipelineserver

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/convert"
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

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	protorpclog.Logger
	pfsAPIClient     pfs.APIClient
	jobAPIClient     pps.JobAPIClient
	persistAPIClient persist.APIClient
	test             bool

	started                          bool
	pipelineNameToPipelineController map[string]*pipelineController
	lock                             *sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	jobAPIClient pps.JobAPIClient,
	persistAPIClient persist.APIClient,
	test bool,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.PipelineAPI"),
		pfsAPIClient,
		jobAPIClient,
		persistAPIClient,
		test,
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
	pipelines, err := a.GetAllPipelines(context.Background(), emptyInstance)
	if err != nil {
		return err
	}
	for _, pipeline := range pipelines.Pipeline {
		if err := a.addPipelineController(pipeline); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) CreatePipeline(ctx context.Context, request *pps.CreatePipelineRequest) (response *pps.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipeline, err := a.persistAPIClient.CreatePipeline(ctx, convert.PipelineToPersist(request.Pipeline))
	if err != nil {
		return nil, err
	}
	if err := a.registerChangeEvent(
		ctx,
		&changeEvent{
			Type:         changeEventTypeCreate,
			PipelineName: persistPipeline.Name,
		},
	); err != nil {
		// TODO(pedge): need to roll back the db create
		return nil, err
	}
	return convert.PersistToPipeline(persistPipeline), nil
}

func (a *apiServer) GetPipeline(ctx context.Context, request *pps.GetPipelineRequest) (response *pps.Pipeline, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelines, err := a.persistAPIClient.GetPipelinesByName(ctx, &google_protobuf.StringValue{Value: request.PipelineName})
	if err != nil {
		return nil, err
	}
	if len(persistPipelines.Pipeline) == 0 {
		return nil, fmt.Errorf("pachyderm.pps.server: no piplines for name %s", request.PipelineName)
	}
	return convert.PersistToPipeline(persistPipelines.Pipeline[0]), nil
}

func (a *apiServer) GetAllPipelines(ctx context.Context, request *google_protobuf.Empty) (response *pps.Pipelines, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	persistPipelines, err := a.persistAPIClient.GetAllPipelines(ctx, request)
	if err != nil {
		return nil, err
	}
	pipelineMap := make(map[string]*pps.Pipeline)
	for _, persistPipeline := range persistPipelines.Pipeline {
		// pipelines are ordered newest to oldest, so if we have already
		// seen a pipeline with the same name, it is newer
		if _, ok := pipelineMap[persistPipeline.Name]; !ok {
			pipelineMap[persistPipeline.Name] = convert.PersistToPipeline(persistPipeline)
		}
	}
	pipelines := make([]*pps.Pipeline, len(pipelineMap))
	i := 0
	for _, pipeline := range pipelineMap {
		pipelines[i] = pipeline
		i++
	}
	return &pps.Pipelines{
		Pipeline: pipelines,
	}, nil
}

type changeEvent struct {
	Type         changeEventType
	PipelineName string
}

func (a *apiServer) registerChangeEvent(ctx context.Context, request *changeEvent) error {
	a.lock.Lock()
	defer a.lock.Unlock()
	switch request.Type {
	case changeEventTypeCreate:
		if !a.pipelineRegistered(request.PipelineName) {
			pipeline, err := a.GetPipeline(ctx, &pps.GetPipelineRequest{PipelineName: request.PipelineName})
			if err != nil {
				return err
			}
			if err := a.addPipelineController(pipeline); err != nil {
				return err
			}
			// TODO(pedge): what to do?
		} else {
			protolog.Warnf("pachyderm.pps.pipelineserver: had a create change event for an existing pipeline: %v", request)
			if err := a.removePipelineController(request.PipelineName); err != nil {
				return err
			}
			pipeline, err := a.GetPipeline(ctx, &pps.GetPipelineRequest{PipelineName: request.PipelineName})
			if err != nil {
				return err
			}
			if err := a.addPipelineController(pipeline); err != nil {
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

func (a *apiServer) addPipelineController(pipeline *pps.Pipeline) error {
	pipelineController := newPipelineController(
		a.pfsAPIClient,
		a.jobAPIClient,
		pps.NewLocalPipelineAPIClient(a),
		a.test,
		pipeline,
	)
	a.pipelineNameToPipelineController[pipeline.Name] = pipelineController
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
