package run

import (
	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/log"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/graph"
	"github.com/pachyderm/pachyderm/src/pps/source"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type runner struct {
	sourcer         source.Sourcer
	grapher         graph.Grapher
	containerClient container.Client
	storeClient     store.Client
}

func newRunner(
	sourcer source.Sourcer,
	grapher graph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
) *runner {
	return &runner{
		sourcer,
		grapher,
		containerClient,
		storeClient,
	}
}

func (r *runner) Start(pipelineSource *pps.PipelineSource) (string, error) {
	dirPath, pipeline, err := r.sourcer.GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return "", err
	}
	pipelineRunID := common.NewUUID()
	if err := r.storeClient.AddPipelineRun(
		pipelineRunID,
		pipelineSource,
		pipeline,
	); err != nil {
		return "", err
	}
	log.Printf("%v %s %v\n", dirPath, pipelineRunID, pipeline)
	nameToNode := pps.GetNameToNode(pipeline)
	nameToNodeInfo, err := graph.GetNameToNodeInfo(nameToNode)
	if err != nil {
		return "", err
	}
	nameToNodeFunc := make(map[string]func() error)
	for name := range nameToNode {
		name := name
		nameToNodeFunc[name] = func() error {
			log.Printf("RUNNING %s\n", name)
			return nil
		}
	}
	run, err := r.grapher.Build(
		&dummyNodeErrorRecorder{},
		nameToNodeInfo,
		nameToNodeFunc,
	)
	if err != nil {
		return "", err
	}
	go run.Do()
	if err := r.storeClient.AddPipelineRunStatus(
		pipelineRunID,
		pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED,
	); err != nil {
		return "", err
	}
	return pipelineRunID, nil
}

type dummyNodeErrorRecorder struct{}

func (d *dummyNodeErrorRecorder) Record(nodeName string, err error) {
	log.Printf("%s HAD ERROR %v\n", nodeName, err)
}
