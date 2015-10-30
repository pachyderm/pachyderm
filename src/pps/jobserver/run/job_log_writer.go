package jobserverrun

import (
	"fmt"

	"golang.org/x/net/context"

	"go.pedge.io/pkg/time"

	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
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
	// TODO: probably do not need to copy at all, assuming we document the persist API
	// to not retain p in any way
	c := make([]byte, len(p))
	if n := copy(c, p); n != len(p) {
		return 0, fmt.Errorf("tried to copy %d bytes, only copied %d bytes", len(p), n)
	}
	if _, err := w.persistAPIClient.CreateJobLog(
		context.Background(),
		// TODO: do we want to set the timestamp in the persist API server implementation,
		// or have it be based on the actual log time? actual log time (which we do not propogate yet)
		// seems like a better option, but to be consistent while the persist API is not well documented,
		// setting it there for now
		&persist.JobLog{
			JobId:        w.jobID,
			OutputStream: w.outputStream,
			Value:        c,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
