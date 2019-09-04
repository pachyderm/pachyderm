package logs

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

type DummyLogger struct {
	// These fields are exposed so that tests can fuck around with them or make assertions
	Writer   io.Writer
	Job      string
	Data     []*common.Input
	UserCode bool
}

// Not used - forces a compile-time error in this file if DummyLogger does not
// implement TaggedLogger
func ensureInterface() TaggedLogger {
	return &DummyLogger{}
}

func NewDummyLogger() *DummyLogger {
	return &DummyLogger{}
}

func (dl *DummyLogger) Write(p []byte) (_ int, retErr error) {
	if dl.Writer != nil {
		return dl.Writer.Write(p)
	}
	return len(p), nil
}

func (dl *DummyLogger) Logf(formatString string, args ...interface{}) {
	if dl.Writer != nil {
		str := fmt.Sprintf(formatString, args...)
		dl.Writer.Write([]byte(str))
	}
}

func (dl *DummyLogger) Errf(formatString string, args ...interface{}) {
	if dl.Writer != nil {
		str := fmt.Sprintf(formatString, args...)
		dl.Writer.Write([]byte(str))
	}
}

func (dl *DummyLogger) clone() *DummyLogger {
	result := &DummyLogger{}
	*result = *dl
	return result
}

func (dl *DummyLogger) WithJob(jobID string) TaggedLogger {
	result := dl.clone()
	result.Job = jobID
	return result
}

func (dl *DummyLogger) WithData(data []*common.Input) TaggedLogger {
	result := dl.clone()
	result.Data = data
	return result
}

func (dl *DummyLogger) WithUserCode() TaggedLogger {
	result := dl.clone()
	result.UserCode = true
	return result
}

func (dl *DummyLogger) JobID() string {
	return dl.Job
}

func (dl *DummyLogger) Close() (*pfs.Object, int64, error) {
	// If you need an actual pfs.Object here, inherit and shadow this function
	return nil, 0, nil
}
