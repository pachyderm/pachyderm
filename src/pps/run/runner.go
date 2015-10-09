package run

import (
	"fmt"
	"io/ioutil"

	"go.pedge.io/pkg/graph"
	"go.pedge.io/pkg/time"
	"go.pedge.io/protolog"

	"go.pachyderm.com/pachyderm/src/pkg/container"
	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/store"
)

type runner struct {
	grapher         pkggraph.Grapher
	containerClient container.Client
	storeClient     store.Client
	timer           pkgtime.Timer
}

func newRunner(
	grapher pkggraph.Grapher,
	containerClient container.Client,
	storeClient store.Client,
	timer pkgtime.Timer,
) *runner {
	return &runner{
		grapher,
		containerClient,
		storeClient,
		timer,
	}
}

func (r *runner) Start(pipelineRunID string) error {
	pipelineRun, err := r.storeClient.GetPipelineRun(pipelineRunID)
	if err != nil {
		return err
	}
	pipeline, err := r.storeClient.GetPipeline(pipelineRun.PipelineId)
	if err != nil {
		return err
	}
	nameToNodeInfo, err := getNameToNodeInfo(pipeline.NameToNode)
	if err != nil {
		return err
	}
	nameToNodeFunc := make(map[string]func() error)
	for name, node := range pipeline.NameToNode {
		nodeFunc, err := r.getNodeFunc(
			pipelineRun.Id,
			name,
			node,
			pipeline.NameToDockerService,
			//dirPath,
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
	if err := r.storeClient.CreatePipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_STARTED); err != nil {
		return err
	}
	go func() {
		if err := run.Do(); err != nil {
			protolog.Errorln(err.Error())
			if storeErr := r.storeClient.CreatePipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_ERROR); storeErr != nil {
				protolog.Errorln(storeErr.Error())
			}
		} else {
			if storeErr := r.storeClient.CreatePipelineRunStatus(pipelineRunID, pps.PipelineRunStatusType_PIPELINE_RUN_STATUS_TYPE_SUCCESS); storeErr != nil {
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
	//dirPath string,
	numContainers int,
) (func() error, error) {
	dockerService, ok := nameToDockerService[node.Service]
	if !ok {
		return nil, fmt.Errorf("no service for name %s", node.Service)
	}
	return func() (retErr error) {
		image := dockerService.Image
		if dockerService.Build != "" {
			image = node.Service
			if err := r.containerClient.Build(
				node.Service,
				dockerService.Build,
				container.BuildOptions{
					Dockerfile:   dockerService.Dockerfile,
					OutputStream: ioutil.Discard,
				},
			); err != nil {
				return err
			}
		} else if err := r.containerClient.Pull(
			dockerService.Image,
			container.PullOptions{},
		); err != nil {
			return err
		}
		var containers []string
		defer func() {
			for _, containerID := range containers {
				_ = r.containerClient.Kill(containerID, container.KillOptions{})
				_ = r.containerClient.Remove(containerID, container.RemoveOptions{})
			}
		}()
		for i := 0; i < numContainers; i++ {
			container, err := r.containerClient.Create(
				image,
				container.CreateOptions{
					Binds:      append(getInputBinds(node.Input), getOutputBinds(node.Output)...),
					HasCommand: len(node.Run) > 0,
				},
			)
			if err != nil {
				return err
			}
			containers = append(containers, container)
		}
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
		var err error
		for _ = range containers {
			if logsErr := <-errC; logsErr != nil && err == nil {
				err = logsErr
			}
		}
		return err
	}, nil
}
