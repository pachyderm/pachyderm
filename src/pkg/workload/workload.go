package workload

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/pfsutil"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/ppsutil"
)

func RunWorkload(
	pfsClient pfs.APIClient,
	ppsClient pps.APIClient,
	rand *rand.Rand,
	size int,
) error {
	worker := newWorker(rand)
	for i := 0; i < size; i++ {
		if err := worker.work(pfsClient, ppsClient); err != nil {
			return err
		}
	}
	return nil
}

type worker struct {
	repos     []*pfs.Repo
	finished  []*pfs.Commit
	started   []*pfs.Commit
	files     []*pfs.File
	jobs      []*pps.Job
	pipelines []*pps.Pipeline
	rand      *rand.Rand
}

func newWorker(rand *rand.Rand) *worker {
	return &worker{
		rand: rand,
	}
}

const (
	repo     float64 = .1
	commit           = .3
	file             = .6
	job              = .9
	pipeline         = 1.0
)

const maxStartedCommits = 6

func (w *worker) work(pfsClient pfs.APIClient, ppsClient pps.APIClient) error {
	opt := w.rand.Float64()
	switch {
	case opt < repo:
		repoName := w.name()
		if err := pfsutil.CreateRepo(pfsClient, repoName); err != nil {
			return err
		}
		w.repos = append(w.repos, &pfs.Repo{Name: repoName})
		commitID, err := pfsutil.StartCommit(pfsClient, repoName, "")
		if err != nil {
			return err
		}
		w.started = append(w.started, commitID)
	case opt < commit:
		if len(w.started) >= maxStartedCommits {
			i := w.rand.Intn(len(w.started))
			commit := w.started[i]
			if err := pfsutil.FinishCommit(pfsClient, commit.Repo.Name, commit.Id); err != nil {
				return err
			}
			w.started = append(w.started[:i], w.started[i+1:]...)
			w.finished = append(w.finished, commit)
		} else {
			commit := w.finished[w.rand.Intn(len(w.finished))]
			commitID, err := pfsutil.StartCommit(pfsClient, commit.Repo.Name, commit.Id)
			if err != nil {
				return err
			}
			w.started = append(w.started, commitID)
		}
	case opt < file:
		commit := w.started[w.rand.Intn(len(w.started))]
		if _, err := pfsutil.PutFile(pfsClient, commit.Repo.Name, commit.Id, w.name(), 0, w.reader()); err != nil {
			return err
		}
	case opt < job:
		inputs := [5]string{}
		var inputCommits []*pfs.Commit
		for i := range inputs {
			randI := w.rand.Intn(len(w.finished))
			inputs[i] = w.finished[randI].Repo.Name
			inputCommits = append(inputCommits, w.finished[randI])
		}
		var parentJobID string
		if len(w.jobs) > 0 {
			parentJobID = w.jobs[w.rand.Intn(len(w.jobs))].Id
		}
		outFilename := w.name()
		job, err := ppsutil.CreateJob(
			ppsClient,
			"",
			[]string{"sh"},
			w.grepCmd(inputs, outFilename),
			1,
			inputCommits,
			parentJobID,
		)
		if err != nil {
			return err
		}
		w.jobs = append(w.jobs, job)
	case opt < pipeline:
		inputs := [5]string{}
		var inputRepos []*pfs.Repo
		for i := range inputs {
			randI := w.rand.Intn(len(w.repos))
			inputs[i] = w.repos[randI].Name
			inputRepos = append(inputRepos, w.repos[randI])
		}
		pipelineName := w.name()
		outFilename := w.name()
		if err := ppsutil.CreatePipeline(
			ppsClient, pipelineName,
			"",
			[]string{"sh"},
			w.grepCmd(inputs, outFilename),
			1,
			inputRepos,
		); err != nil {
			return err
		}

	}
	return nil
}

const letters = "abcdefghijklmnopqrstuvwxyz"
const lettersAndSpaces = "abcdefghijklmnopqrstuvwxyz      \n\n"

func (w *worker) name() string {
	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

type reader struct {
	rand *rand.Rand
}

func (r *reader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = lettersAndSpaces[r.rand.Intn(len(lettersAndSpaces))]
	}
	if rand.Intn(10) == 0 {
		return len(p), io.EOF
	}
	return len(p), nil
}

func (w *worker) reader() io.Reader {
	return &reader{w.rand}
}

func (w *worker) grepCmd(inputs [5]string, outFilename string) string {
	pattern := make([]byte, 5)
	_, _ = w.reader().Read(pattern)
	return fmt.Sprintf(
		"grep %s /pfs/{%s,%s,%s,%s,%s}/* >/pfs/out/%s",
		pattern,
		inputs[0],
		inputs[1],
		inputs[2],
		inputs[3],
		inputs[4],
		outFilename,
	)
}
