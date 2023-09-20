package transform

import (
	"context"
	"sync"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
	batchMutex  sync.Mutex
	nextChan    chan error
	setupChan   chan []string
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
	defer s.mutex.Unlock()
	cb()
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
	s.withLock(func() {
		status := &pps.DatumStatus{
			Data:    convertInputs(inputs),
			Started: timestamppb.Now(),
		}
		s.datumStatus = status
		s.cancel = cancel
	})
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
		JobId:       s.jobID,
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

func (s *Status) withDatumBatch(cb func(<-chan error, chan<- []string) error) error {
	s.withBatchLock(func() {
		if s.nextChan != nil {
			panic("multiple goroutines attempting to set up a datum batch")
		}
		s.nextChan, s.setupChan = make(chan error), make(chan []string)
	})
	defer s.withBatchLock(func() {
		s.nextChan, s.setupChan = nil, nil
	})
	return cb(s.nextChan, s.setupChan)
}

func (s *Status) withBatchLock(cb func()) {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()
	cb()
}

func (s *Status) NextDatum(ctx context.Context, err error) ([]string, error) {
	s.batchMutex.Lock()
	defer s.batchMutex.Unlock()
	if s.nextChan == nil {
		return nil, errors.New("datum batching not enabled")
	}
	select {
	case s.nextChan <- err:
	case <-ctx.Done():
		return nil, errors.EnsureStack(context.Cause(ctx))
	}
	select {
	case env := <-s.setupChan:
		return env, nil
	case <-ctx.Done():
		return nil, errors.EnsureStack(context.Cause(ctx))
	}
}
