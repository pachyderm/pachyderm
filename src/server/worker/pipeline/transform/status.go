package transform

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/worker/common"
)

// Status is a struct representing the current status of the transform worker,
// its public interface only allows getting the status of a task and canceling
// the currently-processing datum.
type Status struct {
	mutex         sync.Mutex
	jobID         string
	stats         *pps.ProcessStats
	queueSize     *int64
	dataProcessed *int64
	dataRecovered *int64
	datum         []*pps.InputFile
	cancel        func()
	started       time.Time
}

func convertInputs(inputs []*common.Input) []*pps.InputFile {
	var result []*pps.InputFile
	for _, input := range inputs {
		result = append(result, &pps.InputFile{
			Path: input.FileInfo.File.Path,
			Hash: input.FileInfo.Hash,
		})
	}
	return result
}

func (s *Status) withLock(cb func()) {
	s.mutex.Lock()
	cb()
	s.mutex.Unlock()
}

func (s *Status) withJob(jobID string, cb func() error) error {
	s.withLock(func() {
		s.jobID = jobID
	})

	defer s.withLock(func() {
		s.jobID = ""
	})

	return cb()
}

func (s *Status) withStats(stats *pps.ProcessStats, queueSize *int64, dataProcessed, dataRecovered *int64, cb func() error) error {
	s.withLock(func() {
		s.stats = stats
		s.queueSize = queueSize
		s.dataProcessed = dataProcessed
		s.dataRecovered = dataRecovered
	})

	defer s.withLock(func() {
		s.stats = nil
		s.queueSize = nil
		s.dataProcessed = nil
		s.dataRecovered = nil
	})

	return cb()
}

func (s *Status) withDatum(inputs []*common.Input, cancel func(), cb func() error) error {
	s.withLock(func() {
		s.datum = convertInputs(inputs)
		s.cancel = cancel
		s.started = time.Now()
	})

	defer s.withLock(func() {
		s.datum = nil
		s.cancel = nil
		s.started = time.Time{}
	})

	return cb()
}

// GetStatus returns the current WorkerStatus for the transform worker
func (s *Status) GetStatus() (*pps.WorkerStatus, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	started, err := types.TimestampProto(s.started)
	if err != nil {
		return nil, err
	}
	result := &pps.WorkerStatus{
		JobID:   s.jobID,
		Data:    s.datum,
		Started: started,
	}
	if s.queueSize != nil {
		result.QueueSize = atomic.LoadInt64(s.queueSize)
	}
	if s.dataProcessed != nil {
		result.DataProcessed = atomic.LoadInt64(s.dataProcessed)
	}
	if s.dataRecovered != nil {
		result.DataRecovered = atomic.LoadInt64(s.dataRecovered)
	}
	return result, nil
}

// Cancel cancels the currently running datum if it matches the specified job and inputs
func (s *Status) Cancel(jobID string, datumFilter []string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if jobID == s.jobID && common.MatchDatum(datumFilter, s.datum) {
		// Fields will be cleared as the worker stack unwinds
		s.cancel()
		return true
	}
	return false
}
