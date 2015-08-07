package run

import (
	"fmt"

	"go.pedge.io/protolog"

	"github.com/pachyderm/pachyderm/src/common"
	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
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
	_, pipeline, err := r.sourcer.GetDirPathAndPipeline(pipelineSource)
	if err != nil {
		return "", err
	}
	pipelineRunID := common.NewUUID()
	pipelineRun := &pps.PipelineRun{
		Id:             pipelineRunID,
		Pipeline:       pipeline,
		PipelineSource: pipelineSource,
	}
	if err := r.storeClient.AddPipelineRun(pipelineRun); err != nil {
		return "", err
	}
	protolog.Info(
		&AddedPipelineRun{
			PipelineRun: pipelineRun,
		},
	)
	nameToNode := pps.GetNameToNode(pipeline)
	nameToDockerService := pps.GetNameToDockerService(pipeline)
	nameToNodeInfo, err := getNameToNodeInfo(nameToNode)
	if err != nil {
		return "", err
	}
	nameToNodeFunc := make(map[string]func() error)
	for name, node := range nameToNode {
		nodeFunc, err := r.getNodeFunc(
			pipelineRunID,
			name,
			node,
			nameToDockerService,
			1,
		)
		if err != nil {
			return "", err
		}
		nameToNodeFunc[name] = nodeFunc
	}
	run, err := r.grapher.Build(
		newNodeErrorRecorder(
			pipelineRunID,
			r.storeClient,
		),
		nameToNodeInfo,
		nameToNodeFunc,
	)
	if err != nil {
		return "", err
	}
	if err := r.storeClient.AddPipelineRunStatus(
		&pps.PipelineRunStatus{
			PipelineRunId:         pipelineRunID,
			PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED,
		}); err != nil {
		return "", err
	}
	go func() {
		if err := run.Do(); err != nil {
			_ = r.storeClient.AddPipelineRunStatus(
				&pps.PipelineRunStatus{
					PipelineRunId:         pipelineRunID,
					PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED,
				})
		} else {
			_ = r.storeClient.AddPipelineRunStatus(
				&pps.PipelineRunStatus{
					PipelineRunId:         pipelineRunID,
					PipelineRunStatusType: pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED,
				})
		}
	}()
	return pipelineRunID, nil
}

func (r *runner) getNodeFunc(
	pipelineRunID string,
	name string,
	node *pps.Node,
	nameToDockerService map[string]*pps.DockerService,
	numContainers int,
) (func() error, error) {
	dockerService, ok := nameToDockerService[node.Service]
	if !ok {
		return nil, fmt.Errorf("no service for name %s", node.Service)
	}
	if dockerService.Build != "" || dockerService.Dockerfile != "" {
		return nil, fmt.Errorf("build/dockerfile not supported yet")
	}
	return func() (retErr error) {
		if err := r.containerClient.Pull(dockerService.Image, container.PullOptions{}); err != nil {
			return err
		}
		containers, err := r.containerClient.Create(
			dockerService.Image,
			container.CreateOptions{
				Input:         node.Input,
				Output:        node.Output,
				HasCommand:    len(node.Run) > 0,
				NumContainers: numContainers,
			},
		)
		if err != nil {
			return err
		}
		defer func() {
			for _, containerID := range containers {
				_ = r.containerClient.Kill(containerID, container.KillOptions{})
				_ = r.containerClient.Remove(containerID, container.RemoveOptions{})
			}
		}()
		for _, containerID := range containers {
			if err := r.containerClient.Start(
				containerID,
				container.StartOptions{
					Commands: node.Run,
				},
			); err != nil {
				return err
			}
		}
		for _, containerID := range containers {
			if err := r.containerClient.Wait(containerID, container.WaitOptions{}); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

type nodeErrorRecorder struct {
	pipelineRunID string
	storeClient   store.Client
}

func newNodeErrorRecorder(
	pipelineRunID string,
	storeClient store.Client,
) *nodeErrorRecorder {
	return &nodeErrorRecorder{
		pipelineRunID,
		storeClient,
	}
}

func (n *nodeErrorRecorder) Record(nodeName string, err error) {
	protolog.Error(
		&PipelineRunError{
			PipelineRunId: n.pipelineRunID,
			Node:          nodeName,
			Error:         errorString(err),
		},
	)
}

func errorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
