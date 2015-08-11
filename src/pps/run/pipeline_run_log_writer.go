package run

import (
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/pachyderm/pachyderm/src/pkg/timing"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/store"
)

type pipelineRunLogWriter struct {
	pipelineRunID string
	containerID   string
	node          string
	outputStream  pps.OutputStream
	timer         timing.Timer
	storeClient   store.Client
}

func newPipelineRunLogWriter(
	pipelineRunID string,
	containerID string,
	node string,
	outputStream pps.OutputStream,
	timer timing.Timer,
	storeClient store.Client,
) *pipelineRunLogWriter {
	return &pipelineRunLogWriter{
		pipelineRunID,
		containerID,
		node,
		outputStream,
		timer,
		storeClient,
	}
}

func (w *pipelineRunLogWriter) Write(p []byte) (int, error) {
	if err := w.storeClient.AddPipelineRunLogs(
		&pps.PipelineRunLog{
			PipelineRunId: w.pipelineRunID,
			ContainerId:   w.containerID,
			Node:          w.node,
			OutputStream:  w.outputStream,
			Timestamp:     protoutil.TimeToTimestamp(w.timer.Now()),
			Data:          p,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
