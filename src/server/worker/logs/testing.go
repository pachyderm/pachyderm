package logs

import (
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

type MockLogger struct {
	// These fields are exposed so that tests can fuck around with them or make assertions
	Writer   io.Writer
	Job      string
	Data     []*common.Input
	UserCode bool
}

// Not used - forces a compile-time error in this file if MockLogger does not
// implement TaggedLogger
var _ TaggedLogger = &MockLogger{}

func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

func (dl *MockLogger) Write(p []byte) (_ int, retErr error) {
	if dl.Writer != nil {
		return dl.Writer.Write(p)
	}
	return len(p), nil
}

func (dl *MockLogger) Logf(formatString string, args ...interface{}) {
	if dl.Writer != nil {
		str := fmt.Sprintf(formatString, args...)
		dl.Writer.Write([]byte(str))
	}
}

func (dl *MockLogger) Errf(formatString string, args ...interface{}) {
	if dl.Writer != nil {
		str := fmt.Sprintf(formatString, args...)
		dl.Writer.Write([]byte(str))
	}
}

func (dl *MockLogger) clone() *MockLogger {
	result := &MockLogger{}
	*result = *dl
	return result
}

func (dl *MockLogger) WithJob(jobID string) TaggedLogger {
	result := dl.clone()
	result.Job = jobID
	return result
}

func (dl *MockLogger) WithData(data []*common.Input) TaggedLogger {
	result := dl.clone()
	result.Data = data
	return result
}

func (dl *MockLogger) WithUserCode() TaggedLogger {
	result := dl.clone()
	result.UserCode = true
	return result
}

func (dl *MockLogger) JobID() string {
	return dl.Job
}

func (dl *MockLogger) Close() (*pfs.Object, int64, error) {
	// If you need an actual pfs.Object here, inherit and shadow this function
	return nil, 0, nil
}
