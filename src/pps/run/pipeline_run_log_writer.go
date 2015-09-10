package run

import (
	"fmt"

	"go.pedge.io/proto/time"

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
	c := make([]byte, len(p))
	if n := copy(c, p); n != len(p) {
		return 0, fmt.Errorf("tried to copy %d bytes, only copied %d bytes", len(p), n)
	}
	if err := w.storeClient.AddPipelineRunLogs(
		&pps.PipelineRunLog{
			PipelineRunId: w.pipelineRunID,
			ContainerId:   w.containerID,
			Node:          w.node,
			OutputStream:  w.outputStream,
			Timestamp:     prototime.TimeToTimestamp(w.timer.Now()),
			Data:          c,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
