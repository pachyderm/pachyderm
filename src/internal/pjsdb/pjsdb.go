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

type JobID uint64

// QueueID is spec hash for now.
type QueueID []byte

// FilesetHandle is a type alias that exists purely for readability.
type FilesetHandle []byte

// Job is the internal representation of a job.
type Job struct {
	ID     JobID
	Parent JobID

	Input    FilesetHandle
	Spec     FilesetHandle
	SpecHash FilesetHandle
	Output   FilesetHandle

	Error string

	Queued     time.Time
	Processing time.Time
	Done       time.Time
	// additional fields or metadata that would be computed would go here.
}

// jobRow models a single row in the pjs.jobs table.
type jobRow struct {
	ID     JobID         `db:"id"`
	Parent sql.NullInt64 `db:"parent"`

	Input    []byte `db:"input"`
	Spec     []byte `db:"spec"`
	SpecHash []byte `db:"spec_hash"`
	Output   []byte `db:"output"`

	Error sql.NullString `db:"error"`

	Queued     time.Time    `db:"queued"`
	Processing sql.NullTime `db:"processing"`
	Done       sql.NullTime `db:"done"`
}

func (r jobRow) ToJob() Job {
	return Job{
		ID:     r.ID,
		Parent: JobID(r.Parent.Int64),

		Input:    r.Input,
		Spec:     r.Spec,
		SpecHash: r.SpecHash,
		Output:   r.Output,

		Error: r.Error.String,

		Queued:     r.Queued,
		Processing: r.Processing.Time,
		Done:       r.Done.Time,
	}
}

// Queue is the internal representation of a queue.
type Queue struct {
	ID    QueueID
	Specs []FilesetHandle
	Jobs  []JobID
	Size  uint64
	// additional fields or metadata that would be computed would go here.
}

// queueRecord is derived from the pjs.jobs table.
// note that this is a 'record' and not a row because queue rows don't exist.
// Instead, the API groups together job ids, specs by distinct spec hashes.
type queueRecord struct {
	ID    QueueID `db:"id"`
	Specs string  `db:"specs"`
	Jobs  string  `db:"jobs"`
	Size  uint64  `db:"size"`
}

func (r *queueRecord) toQueue() (Queue, error) {
	specs, err := r.parseSpecs()
	if err != nil {
		return Queue{}, errors.Wrap(err, "to queue")
	}
	jobs, err := r.parseJobs()
	if err != nil {
		return Queue{}, errors.Wrap(err, "to queue")
	}
	queue := Queue{
		ID:    r.ID,
		Specs: specs,
		Jobs:  jobs,
	}
	return queue, nil
}

// postgres returns array aggregates with the format '{item1, item2, item3, ...}'
// an empty aggregation is returned as {NULL}. In the case of queues, because
// each record returned is a distinct spec hash, that necessitates at least one job and spec
// per returned record.
func inArrayAgg(array string) []string {
	return strings.Split(strings.Trim(array, "{}"), ",")
}

// postgres formats hex strings with the prefix "\\x"
func withoutHexPrefix(spec string) string {
	return strings.Trim(strings.Trim(spec, "\""), "\\x")
}

func (r *queueRecord) parseSpecs() ([]FilesetHandle, error) {
	var specs []FilesetHandle
	for _, spec := range inArrayAgg(r.Specs) {
		decodedSpec, err := hex.DecodeString(withoutHexPrefix(spec))
		if err != nil {
			return nil, errors.Wrap(err, "parse specs")
		}
		specs = append(specs, decodedSpec)
	}
	return specs, nil
}

func (r *queueRecord) parseJobs() ([]JobID, error) {
	var jobs []JobID
	for _, job := range inArrayAgg(r.Jobs) {
		j, err := strconv.ParseUint(job, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "parse jobs")
		}
		jobs = append(jobs, JobID(j))
	}
	return jobs, nil
}
