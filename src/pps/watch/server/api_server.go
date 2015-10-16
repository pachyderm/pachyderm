package server

import (
	"errors"
	"sync"

	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pps/persist"
	"go.pachyderm.com/pachyderm/src/pps/watch"
	"go.pedge.io/google-protobuf"
	"golang.org/x/net/context"
)

var (
	emptyInstance = &google_protobuf.Empty{}
)

type apiServer struct {
	pfsAPIClient     pfs.ApiClient
	persistAPIClient persist.APIClient

	started                          bool
	pipelineNameToPipelineController map[string]*pipelineController
	lock                             *sync.Mutex
}

func newAPIServer(
	pfsAPIClient pfs.ApiClient,
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
	pipelines, err := a.getAllPipelines()
	if err != nil {
		return nil, err
	}
	for _, pipeline := range pipelines {
		pipelineController := newPipelineController(
			a.pfsAPIClient,
			a.persistAPIClient,
			pipeline,
		)
		a.pipelineNameToPipelineController[pipeline.Name] = pipelineController
		if err := pipelineController.Start(); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *apiServer) RegisterChangeEvent(ctx context.Context, request *watch.ChangeEvent) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}

func (a *apiServer) getAllPipelines() ([]*persist.Pipeline, error) {
	protoPipelines, err := a.persistAPIClient.GetAllPipelines(context.Background(), emptyInstance)
	if err != nil {
		return nil, err
	}
	pipelineMap := make(map[string]*persist.Pipeline)
	for _, pipeline := range protoPipelines.Pipeline {
		// pipelines are ordered newest to oldest, so if we have already
		// seen a pipeline with the same name, it is newer
		if _, ok := pipelineMap[pipeline.Name]; !ok {
			pipelineMap[pipeline.Name] = pipeline
		}
	}
	pipelines := make([]*persist.Pipeline, len(pipelineMap))
	i := 0
	for _, pipeline := range pipelineMap {
		pipelines[i] = pipeline
		i++
	}
	return pipelines, nil
}
