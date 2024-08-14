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
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

type JobID uint64

// QueueID is program hash for now.
type QueueID []byte

// Job is the internal representation of a job.
type Job struct {
	ID     JobID
	Parent JobID

	Inputs      []fileset.ID
	Program     fileset.ID
	ProgramHash []byte
	ContextHash []byte
	Outputs     []fileset.ID

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

	Program     []byte `db:"program"`
	ProgramHash []byte `db:"program_hash"`
	ContextHash []byte `db:"context_hash"`

	Error sql.NullString `db:"error"`

	Queued     time.Time    `db:"queued"`
	Processing sql.NullTime `db:"processing"`
	Done       sql.NullTime `db:"done"`
}

// jobFilesetsRow models a single row in the pjs.job_filesets table.
type jobFilesetsRow struct {
	JobID         JobID  `db:"job_id"`
	Type          string `db:"fileset_type"`
	ArrayPosition int    `db:"array_position"`
	Fileset       []byte `db:"fileset"`
}

// jobRecord is derived from the pjs.jobs and pjs.job_filesets tables.
// note that this is a 'record' and not a row, because it is the result of joining tables together.
type jobRecord struct {
	jobRow
	Inputs  string `db:"inputs"`
	Outputs string `db:"outputs"`
}

func (r jobRecord) toJob() (Job, error) {
	job := Job{
		ID:          r.ID,
		Parent:      JobID(r.Parent.Int64),
		ProgramHash: r.ProgramHash,
		ContextHash: r.ContextHash,
		Error:       r.Error.String,
		Queued:      r.Queued,
		Processing:  r.Processing.Time,
		Done:        r.Done.Time,
	}
	var err error
	if job.Inputs, err = parseFileset(r.Inputs); err != nil {
		return Job{}, errors.Wrap(err, "to job")
	}
	if job.Outputs, err = parseFileset(r.Outputs); err != nil {
		return Job{}, errors.Wrap(err, "to job")
	}
	program, err := fileset.ParseID(string(r.Program))
	if err != nil {
		return Job{}, errors.Wrap(err, "to job")
	}
	job.Program = *program
	return job, nil
}

// Queue is the internal representation of a queue.
type Queue struct {
	ID       QueueID
	Programs []fileset.ID
	Jobs     []JobID
	Size     uint64
	// additional fields or metadata that would be computed would go here.
}

// queueRecord is derived from the pjs.jobs table.
// note that this is a 'record' and not a row because queue rows don't exist.
// Instead, the API groups together job ids, programs by distinct program hashes.
type queueRecord struct {
	ID       QueueID `db:"id"`
	Programs string  `db:"programs"`
	Jobs     string  `db:"jobs"`
	Size     uint64  `db:"size"`
}

func (r *queueRecord) toQueue() (Queue, error) {
	programs, err := r.parsePrograms()
	if err != nil {
		return Queue{}, errors.Wrap(err, "to queue")
	}
	jobs, err := r.parseJobs()
	if err != nil {
		return Queue{}, errors.Wrap(err, "to queue")
	}
	queue := Queue{
		ID:       r.ID,
		Programs: programs,
		Jobs:     jobs,
	}
	return queue, nil
}
func (r *queueRecord) parsePrograms() ([]fileset.ID, error) {
	return parseFileset(r.Programs)
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

// postgres returns array aggregates with the format '{item1, item2, item3, ...}'
// an empty aggregation is returned as '{}' or contains a single null element: '{NULL}'.
// In the case of queues, because each record returned is a distinct program hash,
// that necessitates at least one job and program per returned record.
// In the case of jobRecords, the '{}' is trimmed away, leaving 0 items in the array.
func inArrayAgg(array string) []string {
	return strings.Split(strings.Trim(array, "{}"), ",")
}

// postgres formats hex strings with the prefix "\\x"
func withoutHexPrefix(fs string) string {
	return strings.Trim(strings.Trim(fs, "\""), "\\x")
}

// the input filesets should be hex code of a string, in string format :)
// Use this function when the input string is read with functions like ARRAY_AGG.
// For example, if the fileset handle looks like e4fddb03882481d2fd47cff01b0ca753, you should use fileset.ParseID.
// If handle looks like \\x6534666464623033383832343831643266643437636666303162306361373533
// then use this function instead. The latter is the hex interpretation of the former.
func parseFileset(filesets string) ([]fileset.ID, error) {
	var ids []fileset.ID
	for _, fs := range inArrayAgg(filesets) {
		if fs == "" { // can happen if the array aggregate was empty '{}'
			continue
		}
		decoded, err := hex.DecodeString(withoutHexPrefix(fs))
		if err != nil {
			return nil, errors.Wrap(err, "parse ids")
		}
		id, err := fileset.ParseID(string(decoded))
		if err != nil {
			return nil, errors.Wrap(err, "parse ids")
		}
		ids = append(ids, *id)
	}
	return ids, nil
}
