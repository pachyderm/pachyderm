package server

import (
	"fmt"

	"golang.org/x/net/context"

	"go.pedge.io/pkg/time"
	"go.pedge.io/proto/time"

	"go.pachyderm.com/pachyderm/src/pps"
	"go.pachyderm.com/pachyderm/src/pps/persist"
)

type jobLogWriter struct {
	jobID            string
	outputStream     pps.OutputStream
	persistAPIClient persist.APIClient
	timer            pkgtime.Timer
}

func newJobLogWriter(
	jobID string,
	outputStream pps.OutputStream,
	persistAPIClient persist.APIClient,
) *jobLogWriter {
	return &jobLogWriter{
		jobID,
		outputStream,
		persistAPIClient,
		pkgtime.NewSystemTimer(),
	}
}

func (w *jobLogWriter) Write(p []byte) (int, error) {
	c := make([]byte, len(p))
	if n := copy(c, p); n != len(p) {
		return 0, fmt.Errorf("tried to copy %d bytes, only copied %d bytes", len(p), n)
	}
	if _, err := w.persistAPIClient.CreateJobLog(
		context.Background(),
		&persist.JobLog{
			JobId:        w.jobID,
			Timestamp:    prototime.TimeToTimestamp(w.timer.Now()),
			OutputStream: w.outputStream,
			Value:        c,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
