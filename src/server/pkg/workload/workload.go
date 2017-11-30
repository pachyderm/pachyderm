package workload

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// RunWorkload runs a test workload against a Pachyderm cluster.
func RunWorkload(
	client *client.APIClient,
	rand *rand.Rand,
	size int,
) error {
	worker := newWorker(rand)
	for i := 0; i < size; i++ {
		if err := worker.work(client); err != nil {
			return err
		}
	}
	for _, job := range worker.startedJobs {
		jobInfo, err := client.InspectJob(job.ID, true)
		if err != nil {
			return err
		}
		if jobInfo.State != pps.JobState_JOB_SUCCESS {
			return fmt.Errorf("job %s failed", job.ID)
		}
	}
	return nil
}

// Constants corresponding to operations that worker.work() may perform. For
// each constant c_i below, (c_i - c_{i-1}) is the probability that operation i
// is performed
const (
	repo     float64 = .02 // Create a repo
	commit           = .1  // Start or finish a commit
	file             = .9  // Put a file
	job              = .98 // Start a job, or wait for a job to finish
	pipeline         = 1.0 // Create a pipeline
)

type worker struct {
	// Repos created by the worker. Modified by the 'repo' operation.
	repos []*pfs.Repo

	// Commits that the worker has finished. Guaranteed to be a subset of
	// 'started'. Modified by 'repo' and 'commit'.
	finished []*pfs.Commit

	// Commits that the worker has started. Modified by 'repo' and 'commit'.
	started []*pfs.Commit

	// Files that the worker has created
	files []*pfs.File

	// Jobs that the worker has started (but may not have finished). Guaranteed to
	// be mutually exclusive with 'jobs'.
	startedJobs []*pps.Job

	// Worker jobs that have definitely finished. Modified by 'job'.
	finishedJobs []*pps.Job

	// Pipelines that the worker has created. Modified by 'pipeline'.
	pipelines []*pps.Pipeline

	// PRNG used to pick each operation
	rand *rand.Rand
}

func newWorker(rand *rand.Rand) *worker {
	return &worker{
		rand: rand,
	}
}

const maxStartedCommits = 6
const maxStartedJobs = 6

func (w *worker) work(c *client.APIClient) error {
	// Generate a random float, and use it to choose the next operation
	opt := w.rand.Float64()
	switch {
	case opt < repo:
		return w.createRepo(c)
	case opt < commit:
		return w.advanceCommit(c)
	case opt < file:
		return w.putFile(c)
	case opt < job:
		return w.advanceJob(c)
	case opt < pipeline:
		return w.createPipeline(c)
	}
	return nil
}

// createRepo creates a new repo in the cluster
func (w *worker) createRepo(c *client.APIClient) error {
	repoName := w.randString(10)
	if err := c.CreateRepo(repoName); err != nil {
		return err
	}
	w.repos = append(w.repos, &pfs.Repo{Name: repoName})

	// Start the first commit in the repo (no parent). This is critical to
	// advanceCommit(), which will try to finish a commit the first time it's
	// called, and therefore must have an open commit to finish.
	commit, err := c.StartCommit(repoName, "")
	if err != nil {
		return err
	}
	w.started = append(w.started, commit)
	return nil
}

// advanceCommit either starts or finishes a commit, depending on the state of
// the cluster.
func (w *worker) advanceCommit(c *client.APIClient) error {
	if len(w.started) >= maxStartedCommits || len(w.finished) == 0 {
		// Randomly select a commit that has been started and finish it
		if len(w.started) == 0 {
			return nil
		}
		i := w.rand.Intn(len(w.started))
		commit := w.started[i]

		// Before we finish a commit, we add a file. This assures that there
		// won't be any empty commits which will later crash jobs
		if _, err := c.PutFile(commit.Repo.Name, commit.ID, w.randString(10), w.reader()); err != nil {
			return err
		}
		if err := c.FinishCommit(commit.Repo.Name, commit.ID); err != nil {
			return err
		}
		// remove commit[i] from 'started' and add it to 'finished'
		w.started = append(w.started[:i], w.started[i+1:]...)
		w.finished = append(w.finished, commit)
	} else {
		// Start a new commmit (parented off of a commit that we've finished)
		commit := w.finished[w.rand.Intn(len(w.finished))]
		commit, err := c.StartCommitParent(commit.Repo.Name, "", commit.ID)
		if err != nil {
			return err
		}
		w.started = append(w.started, commit)
	}
	return nil
}

// putFile puts a file with random contents into a random open commit (or exits
// early if there are none)
func (w *worker) putFile(c *client.APIClient) error {
	if len(w.started) == 0 {
		return nil
	}
	commit := w.started[w.rand.Intn(len(w.started))]
	if _, err := c.PutFile(commit.Repo.Name, commit.ID, w.randString(10), w.reader()); err != nil {
		return err
	}
	return nil
}

// advanceJob either waits for the first job in 'startedJobs' to finish (if its
// job queue is full) or starts a new job
func (w *worker) advanceJob(c *client.APIClient) error {
	if len(w.startedJobs) >= maxStartedJobs {
		// Wait for the first job in 'w.startedJobs' to finish.

		// Pull first job off of 'startedJobs'
		job := w.startedJobs[0]
		copy(w.startedJobs, w.startedJobs[1:])
		w.startedJobs = w.startedJobs[:len(w.startedJobs)-1]

		// Wait for first started job to finish. Crash if the job failed
		jobInfo, err := c.InspectJob(job.ID, true /* block until job finishes */)
		if err != nil {
			return err
		}
		if jobInfo.State != pps.JobState_JOB_SUCCESS {
			return fmt.Errorf("job %s failed", job.ID)
		}
		w.startedJobs = append(w.startedJobs, job)
	} else {
		// Start a new job

		// If there are no finished commits, there will be nothing for the new job
		// to process. Just exit.
		if len(w.finished) == 0 {
			return nil
		}

		// Pick up to 5 finished commits from distinct repos: those are the job
		// inputs. Store the repo names in 'inputs' to generate the 'grep' command
		// that the job will run.
		inputs := [5]string{}
		var input []*pps.Input
		repoSet := make(map[string]bool)
		for i := range inputs {
			commit := w.finished[w.rand.Intn(len(w.finished))]
			if _, ok := repoSet[commit.Repo.Name]; ok {
				// skip commit if we already have one from this repo (job will have
				// one fewer inputs)
				continue
			}
			repoSet[commit.Repo.Name] = true
			inputs[i] = commit.Repo.Name
			input = append(input, client.NewAtomInputOpts("", commit.Repo.Name, commit.ID, "*", false, ""))
		}
		outFilename := w.randString(10)

		// Create a job to grep for a random string in the input files, and write
		// the results to an output file
		job, err := c.CreateJob(
			"", // Image (use default)
			[]string{"bash"},
			w.grepCmd(inputs, outFilename),
			&pps.ParallelismSpec{
				Constant: 1,
			},
			client.NewCrossInput(input...),
			0, // internal port
			0, // external port
		)
		if err != nil {
			return err
		}
		w.startedJobs = append(w.startedJobs, job)
	}
	return nil
}

func (w *worker) createPipeline(c *client.APIClient) error {
	// If we haven't created any repos yet, then a new pipeline won't have
	// anything to process. Just exit.
	if len(w.repos) == 0 {
		return nil
	}

	// Pick up to 5 distinct repos: those are the pipeline inputs. Store the repo
	// names in 'inputs' to generate the 'grep' command that the pipeline's jobs
	// will run.
	inputs := [5]string{}
	var input []*pps.Input
	repoSet := make(map[string]bool)
	for i := range inputs {
		repo := w.repos[w.rand.Intn(len(w.repos))]
		if _, ok := repoSet[repo.Name]; ok {
			// skip commit if we already have one from this repo (job will have
			// one fewer inputs)
			continue
		}
		repoSet[repo.Name] = true
		inputs[i] = repo.Name
		input = append(input, client.NewAtomInput(repo.Name, "*"))
	}

	// Create a pipeline to grep for a random string in the input files, and write
	// the results to an output file
	pipelineName := w.randString(10)
	outFilename := w.randString(10)
	if err := c.CreatePipeline(
		pipelineName,
		"",
		[]string{"bash"},
		w.grepCmd(inputs, outFilename),
		&pps.ParallelismSpec{
			Constant: 1,
		},
		client.NewCrossInput(input...),
		"master",
		false,
	); err != nil {
		return err
	}
	w.pipelines = append(w.pipelines, client.NewPipeline(pipelineName))
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyz"
const lettersAndSpaces = "abcdefghijklmnopqrstuvwxyz      "

// RandString returns a random alphabetical string of size n
func RandString(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func (w *worker) randString(n int) string {
	return RandString(w.rand, n)
}

func (w *worker) grepCmd(inputs [5]string, outFilename string) []string {
	return []string{
		fmt.Sprintf(
			"grep %s /pfs/{%s,%s,%s,%s,%s}/* >/pfs/out/%s; true",
			w.randString(4),
			inputs[0],
			inputs[1],
			inputs[2],
			inputs[3],
			inputs[4],
			outFilename,
		),
	}
}

func (w *worker) reader() io.Reader {
	return NewReader(w.rand, 1000)
}

type reader struct {
	rand      *rand.Rand
	bytes     int64
	bytesRead int64
}

// NewReader returns a Reader which generates strings of characters.
func NewReader(rand *rand.Rand, bytes int64) io.Reader {
	return &reader{
		rand:  rand,
		bytes: bytes,
	}
}

func (r *reader) Read(p []byte) (int, error) {
	var bytesReadThisTime int
	for i := range p {
		if r.bytesRead+int64(bytesReadThisTime) == r.bytes {
			break
		}
		if i%128 == 127 {
			p[i] = '\n'
		} else {
			p[i] = lettersAndSpaces[r.rand.Intn(len(lettersAndSpaces))]
		}
		bytesReadThisTime++
	}
	r.bytesRead += int64(bytesReadThisTime)
	if r.bytesRead == r.bytes {
		return bytesReadThisTime, io.EOF
	}
	return bytesReadThisTime, nil
}
