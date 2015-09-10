package run

import (
	"fmt"

	"go.pedge.io/protolog"

	"github.com/pachyderm/pachyderm/src/pkg/graph"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/container"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type runner struct {
	grapher         graph.Grapher
	containerClient container.Client
	storeClient     store.Client
	timer           timing.Timer
}

func newRunner(
	grapher graph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
	timer timing.Timer,
) *runner {
	return &runner{
		grapher,
		containerClient,
		storeClient,
		timer,
	}
}

func (r *runner) Start(pipelineRun *pps.PipelineRun) error {
	nameToNodeInfo, err := getNameToNodeInfo(pipelineRun.Pipeline.NameToNode)
	if err != nil {
		return err
	}
	nameToNodeFunc := make(map[string]func() error)
	for name, node := range pipelineRun.Pipeline.NameToNode {
		nodeFunc, err := r.getNodeFunc(
			pipelineRun.Id,
			name,
			node,
			pipelineRun.Pipeline.NameToDockerService,
			dirPath,
			1,
		)
		if err != nil {
			return err
		}
		nameToNodeFunc[name] = nodeFunc
	}
	run, err := r.grapher.Build(
		nameToNodeInfo,
		nameToNodeFunc,
	)
	if err != nil {
		return err
	}
	if err := r.storeClient.AddPipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED); err != nil {
		return err
	}
	go func() {
		if err := run.Do(); err != nil {
			if storeErr := r.storeClient.AddPipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ERROR); storeErr != nil {
				protolog.Errorln(storeErr.Error())
			}
		} else {
			if storeErr := r.storeClient.AddPipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS); storeErr != nil {
				protolog.Errorln(storeErr.Error())
			}
		}
	}()
	return nil
}

func (r *runner) getNodeFunc(
	pipelineRunID string,
	name string,
	node *pps.Node,
	nameToDockerService map[string]*pps.DockerService,
	dirPath string,
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
		if err := r.containerClient.Pull(
			dockerService.Image,
			container.PullOptions{
			//OutputStream: newPipelineRunLogWriter(
			//pipelineRunID,
			//"",
			//name,
			//pps.OutputStream_OUTPUT_STREAM_NONE,
			//r.timer,
			//r.storeClient,
			//),
			},
		); err != nil {
			return err
		}
		containers, err := r.containerClient.Create(
			dockerService.Image,
			container.CreateOptions{
				Binds:         append(getInputBinds(node.Input), getOutputBinds(node.Output)...),
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
		errC := make(chan error, len(containers))
		for _, containerID := range containers {
			containerID := containerID
			go func() {
				errC <- r.containerClient.Logs(
					containerID,
					container.LogsOptions{
						Stdout: newPipelineRunLogWriter(
							pipelineRunID,
							containerID,
							name,
							pps.OutputStream_OUTPUT_STREAM_STDOUT,
							r.timer,
							r.storeClient,
						),
						Stderr: newPipelineRunLogWriter(
							pipelineRunID,
							containerID,
							name,
							pps.OutputStream_OUTPUT_STREAM_STDERR,
							r.timer,
							r.storeClient,
						),
					},
				)
			}()
		}
		for _, containerID := range containers {
			if err := r.containerClient.Wait(containerID, container.WaitOptions{}); err != nil {
				return err
			}
		}
		err = nil
		for _ = range containers {
			if logsErr := <-errC; logsErr != nil && err == nil {
				err = logsErr
			}
		}
		return err
	}, nil
}
