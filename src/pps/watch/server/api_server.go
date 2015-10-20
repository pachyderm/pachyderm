package server

import (
	"errors"
	"fmt"
	"sync"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pachyderm.com/pachyderm/src/pps/watch"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	pfsAPIClient     pfs.APIClient
	persistAPIClient persist.APIClient

	started                          bool
	pipelineNameToPipelineController map[string]*pipelineController
	lock                             *sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.APIClient,
	persistAPIClient persist.APIClient,
) *apiServer {
	return &apiServer{
		pfsAPIClient,
		persistAPIClient,
		false,
		make(map[string]*pipelineController),
		&sync.Mutex{},
	}
}

func (a *apiServer) Start(ctx context.Context, request *google_protobuf.Empty) (*google_protobuf.Empty, error) {
	a.lock.Lock()
	defer a.lock.Unlock()
	// TODO(pedge): volatile bool?
	if a.started {
		// TODO(pedge): abstract error to public variable
		return nil, errors.New("pachyderm.pps.watch.server: already started")
	}
	a.started = true
	pipelines, err := getAllPipelines(a.persistAPIClient)
	if err != nil {
		return nil, err
	}
	for _, pipeline := range pipelines {
		if err := a.addPipelineController(pipeline); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *apiServer) RegisterChangeEvent(ctx context.Context, request *watch.ChangeEvent) (*google_protobuf.Empty, error) {
	if request.Type == watch.ChangeEvent_CHANGE_EVENT_TYPE_NONE {
		return nil, fmt.Errorf("pachyderm.pps.watch.server: change event type not set for %v", request)
	}
	if request.PipelineName == "" {
		return nil, fmt.Errorf("pachyderm.pps.watch.server: pipeline name not set for %v", request)
	}
	a.lock.Lock()
	defer a.lock.Unlock()
	switch request.Type {
	case watch.ChangeEvent_CHANGE_EVENT_TYPE_CREATE:
		if !a.pipelineRegistered(request.PipelineName) {
			pipeline, err := getPipeline(a.persistAPIClient, request.PipelineName)
			if err != nil {
				return nil, err
			}
			if err := a.addPipelineController(pipeline); err != nil {
				return nil, err
			}
			// TODO(pedge): what to do?
		} else {
			protolog.Warnf("pachyderm.pps.watch.server: had a create change event for an existing pipeline: %v", request)
			if err := a.removePipelineController(request.PipelineName); err != nil {
				return nil, err
			}
			pipeline, err := getPipeline(a.persistAPIClient, request.PipelineName)
			if err != nil {
				return nil, err
			}
			if err := a.addPipelineController(pipeline); err != nil {
				return nil, err
			}
		}
	case watch.ChangeEvent_CHANGE_EVENT_TYPE_DELETE:
		if !a.pipelineRegistered(request.PipelineName) {
			protolog.Warnf("pachyderm.pps.watch.server: had a delete change event for a pipeline that was not registered: %v", request)
		} else {
			if err := a.removePipelineController(request.PipelineName); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("pachyderm.pps.watch.server: unknown change event type: %v", request.Type)
	}
	return emptyInstance, nil
}

func (a *apiServer) pipelineRegistered(name string) bool {
	_, ok := a.pipelineNameToPipelineController[name]
	return ok
}

func (a *apiServer) addPipelineController(pipeline *persist.Pipeline) error {
	pipelineController := newPipelineController(
		a.pfsAPIClient,
		a.persistAPIClient,
		pipeline,
	)
	a.pipelineNameToPipelineController[pipeline.Name] = pipelineController
	return pipelineController.Start()
}

func (a *apiServer) removePipelineController(name string) error {
	pipelineController, ok := a.pipelineNameToPipelineController[name]
	if !ok {
		return fmt.Errorf("pachyderm.pps.watch.server: no pipeline registered for name: %s", name)
	}
	pipelineController.Cancel()
	return nil
}
