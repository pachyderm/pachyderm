package transform

import (
	"sync"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/server/worker/common"
)

// Status is a struct representing the current status of the transform worker,
// its public interface only allows getting the status of a task and canceling
// the currently-processing datum.
type Status struct {
	mutex       sync.Mutex
	jobID       string
	datumStatus *pps.DatumStatus
	cancel      func()
}

func convertInputs(inputs []*common.Input) []*pps.InputFile {
	var result []*pps.InputFile
	for _, input := range inputs {
		result = append(result, &pps.InputFile{
			Path: input.FileInfo.File.Path,
			Hash: input.FileInfo.Details.Hash,
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

func (s *Status) withDatum(inputs []*common.Input, cancel func(), cb func() error) error {
	var err error
	s.withLock(func() {
		status := &pps.DatumStatus{
			Data: convertInputs(inputs),
		}
		status.Started, err = types.TimestampProto(time.Now())
		if err != nil {
			return
		}
		s.datumStatus = status
		s.cancel = cancel
	})
	if err != nil {
		return err
	}

	defer s.withLock(func() {
		s.datumStatus = nil
		s.cancel = nil
	})

	return cb()
}

// GetStatus returns the current WorkerStatus for the transform worker
func (s *Status) GetStatus() (*pps.WorkerStatus, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return &pps.WorkerStatus{
		JobID:       s.jobID,
		DatumStatus: s.datumStatus,
	}, nil
}

// Cancel cancels the currently running datum if it matches the specified job and inputs
func (s *Status) Cancel(jobID string, datumFilter []string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if jobID == s.jobID && common.MatchDatum(datumFilter, s.datumStatus.Data) {
		// Fields will be cleared as the worker stack unwinds
		s.cancel()
		return true
	}
	return false
}
