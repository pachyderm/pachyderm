// Package pjsdb exports functions and structs for creating, reading, writing,
// and deleting Job and Queue objects in PJS.
package pjsdb

import (
	"database/sql"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	JobErrorCodeFailed       = "failed"
	JobErrorCodeCanceled     = "canceled"
	JobErrorCodeDisconnected = "disconnected"
)

type JobID uint64

// QueueID is spec hash for now.
type QueueID []byte

// FilesetHandle is a type alias that exists purely for readability.
type FilesetHandle []byte

// Job is the internal representation of a job.
type Job struct {
	JobRow
	// additional fields or metadata that would be computed would go here.
}

// JobRow models a single row in the pjs.jobs table.
type JobRow struct {
	ID     JobID         `db:"id"`
	Parent sql.NullInt64 `db:"parent"`

	Input    *FilesetHandle `db:"input"`
	Spec     *FilesetHandle `db:"spec"`
	SpecHash *FilesetHandle `db:"spec_hash"`
	Output   *FilesetHandle `db:"output"`

	Error sql.NullString `db:"error"`

	Queued     time.Time    `db:"queued"`
	Processing sql.NullTime `db:"processing"`
	Done       sql.NullTime `db:"done"`
}

// Queue is the internal representation of a queue.
type Queue struct {
	ID    QueueID
	Specs []*FilesetHandle
	Jobs  []JobID
	Size  uint64
	// additional fields or metadata that would be computed would go here.
}

// queueRecord is derived from the pjs.jobs table.
type queueRecord struct {
	ID    QueueID `db:"id"`
	Specs string  `db:"specs"`
	Jobs  string  `db:"jobs"`
	Size  uint64  `db:"size"`
}

func (r *queueRecord) toQueue() (*Queue, error) {
	specs, err := r.parseSpecs()
	if err != nil {
		return nil, errors.Wrap(err, "to queue")
	}
	jobs, err := r.parseJobs()
	if err != nil {
		return nil, errors.Wrap(err, "to queue")
	}
	queue := &Queue{
		ID:    r.ID,
		Specs: specs,
		Jobs:  jobs,
	}
	return queue, nil
}

func inArrayAgg(array string) []string {
	return strings.Split(strings.Trim(array, "{}"), ",")
}

func (r *queueRecord) parseSpecs() ([]*FilesetHandle, error) {
	var specs []*FilesetHandle
	if r.Specs == "{NULL}" {
		return nil, nil
	}
	for _, spec := range inArrayAgg(r.Specs) {
		specHex := strings.Trim(strings.Trim(spec, "\""), "\\x")
		decodedSpec, err := hex.DecodeString(specHex)
		if err != nil {
			return nil, errors.Wrap(err, "parse specs")
		}
		handle := FilesetHandle(decodedSpec)
		specs = append(specs, &handle)
	}
	return specs, nil
}

func (r *queueRecord) parseJobs() ([]JobID, error) {
	var jobs []JobID
	if r.Jobs == "{NULL}" {
		return nil, nil
	}
	for _, job := range inArrayAgg(r.Jobs) {
		j, err := strconv.ParseUint(job, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse jobs")
		}
		jobs = append(jobs, JobID(j))
	}
	return jobs, nil
}
