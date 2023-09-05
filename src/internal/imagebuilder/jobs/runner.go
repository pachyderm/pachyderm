package jobs

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"golang.org/x/exp/maps"
)

// runner runs jobs.
type runner struct {
	Cache *Cache // The artifact cache.

	artifacts     []Artifact
	artifactCache map[Reference]Artifact
	outputsToJobs map[Reference]Job
}

type jobResult struct {
	Job       Job
	Artifacts []Artifact
	Err       error
}

func newRunner() *runner {
	return &runner{
		Cache: &Cache{
			Path: "/tmp/builder-cache/",
		},
		outputsToJobs: make(map[Reference]Job),
		artifactCache: make(map[Reference]any),
	}
}

func (r *runner) NewJobContext() *JobContext {
	return &JobContext{
		Cache:      r.Cache,
		HTTPClient: http.DefaultClient,
	}
}

// resolveJobInputs resolves the inputs for a job.
func (r *runner) resolveJob(job Job) ([]Artifact, []Job, error) {
	var inputs []Artifact
	var jobs []Job
	var inputErr error
	for _, ref := range job.Inputs() {
		var foundArtifact bool
		for _, a := range r.artifacts {
			if ok := ref.Match(a); ok && foundArtifact {
				errors.JoinInto(&inputErr, fmt.Errorf("input %q matches multiple artifacts", ref))
			} else if ok && !foundArtifact {
				foundArtifact = true
				inputs = append(inputs, a)
			}
		}
		if !foundArtifact {
			var foundJob Job
			for output, job := range r.outputsToJobs {
				if ref.Match(output) {
					if foundJob != nil {
						errors.JoinInto(&inputErr, fmt.Errorf("input %q can be produced by multiple jobs", ref))
					} else {
						foundJob = job
						jobs = append(jobs, job)
					}
				}
			}
			if foundJob == nil {
				errors.JoinInto(&inputErr, fmt.Errorf("input %q matches no artifacts and is produced by no jobs", ref))
			}
			inputs = append(inputs, nil)
		} else {
			jobs = append(jobs, nil)
		}
	}
	if inputErr != nil {
		return inputs, jobs, errors.Wrapf(inputErr, "invalid input sequence %v", inputs)
	}
	return inputs, jobs, nil
}

func dedupJobs(jobs ...Job) []Job {
	result := make(map[Job]struct{})
	for _, j := range jobs {
		if j != nil {
			result[j] = struct{}{}
		}
	}
	return maps.Keys(result)
}

// AddJob adds a job to potentially run.
func (r *runner) addJob(j Job) error {
	inputs, jobs, err := r.resolveJob(j)
	if err != nil {
		// This is not allowed; everything must resolve at the time the job is added.  This
		// avoids cycles.
		return errors.Wrapf(err, "job has unresolvable dependencies [computed inputs: %v, required jobs: %v]; add dependencies before dependents, no cycles", inputs, dedupJobs(jobs...))
	}
	for _, o := range j.Outputs() {
		r.outputsToJobs[o] = j
	}
	return nil
}

func (r *runner) getArtifact(ref Reference) (Artifact, bool) {
	if a, ok := r.artifactCache[ref]; ok {
		return a, true
	}
	for _, a := range r.artifacts {
		if ref.Match(a) {
			r.artifactCache[ref] = a
			return a, true
		}
	}
	return nil, false
}

// Plan returns an execution plan.
func (r *runner) Plan(ctx context.Context, want ...Reference) ([][]string, error) {
	var steps [][]string
	var firstStep []string

	need := map[Reference]struct{}{}
	for _, ref := range want {
		need[ref] = struct{}{}
		firstStep = append(firstStep, fmt.Sprintf("initial dependency on %v", ref))
	}
	steps = append(steps, firstStep)

	jc := &JobContext{
		Cache:      r.Cache,
		HTTPClient: http.DefaultClient,
	}
	runningJobs := make(map[Job]struct{})
	waitingJobs := make(map[Job]struct{})
	resultCh := make(chan jobResult)

	// Breadth-first search over the dependency graph.  We start with what we need, add jobs to
	// satisfy those dependencies, until `want` is produced.  After each iteration that doesn't
	// make progress resolving dependencies, we stop and wait for a job to finish.
	//
	// The 1_000_000 here is the maximum edge count between dependencies plus the total number of
	// jobs plus the total number of outputs.
	for i := 0; i < 1_000_000 && len(need) > 0; i++ {
		var step []string
		var madeProgress bool

		// See what we have to start with.
		for ref := range need {
			if a, ok := r.getArtifact(ref); ok {
				step = append(step, fmt.Sprintf("yield ref %v -> %v", ref, a))
				delete(need, ref)
				madeProgress = true
			}
		}

		// See which jobs we need to run in order to create the artifacts in "need".
		needJobs := make(map[Job]struct{})
		for ref := range need {
			for product, job := range r.outputsToJobs {
				if ref.Match(product) {
					needJobs[job] = struct{}{}
				}
			}
		}

		// See which jobs are possible to run given the current state of "have".
		for job := range needJobs {
			if _, ok := runningJobs[job]; ok {
				// It's already running.
				continue
			}
			_, jobs, err := r.resolveJob(job)
			if err != nil {
				return steps, errors.Wrapf(err, "internal error: resolveJob(%v)", job)
			}
			if len(dedupJobs(jobs...)) == 0 {
				// We can run this job right now.
				if err := r.startJob(ctx, jc, job, runFakeJob, resultCh); err != nil {
					return steps, errors.Wrapf(err, "problem starting job %v", job)
				}
				runningJobs[job] = struct{}{}
				delete(waitingJobs, job)
				step = append(step, fmt.Sprintf("start job %v", job))
				madeProgress = true
				continue
			}

			if _, ok := waitingJobs[job]; ok {
				continue
			}

			// The job isn't running, wasn't started, and isn't waiting to run,
			for _, input := range job.Inputs() {
				need[input] = struct{}{}
				waitingJobs[job] = struct{}{}
				step = append(step, fmt.Sprintf("discovered dependency on %v from %v", input, job))
				madeProgress = true
			}
		}

		// If we didn't do anything this iteration, wait for a job to return.
		if !madeProgress {
			step = append(step, fmt.Sprintf("wait for a job to finish; %v", maps.Keys(runningJobs)))
			select {
			case <-ctx.Done():
				steps = append(steps, step)
				return steps, ctx.Err()
			case result := <-resultCh:
				if err := result.Err; err != nil {
					steps = append(steps, step)
					return steps, errors.Wrapf(err, "received error result from job %v", result.Job)
				}
				r.artifacts = append(r.artifacts, result.Artifacts...)
				delete(runningJobs, result.Job)
				step = append(step, fmt.Sprintf("job %v finished", result.Job))
			}
		}
		steps = append(steps, step)

	}
	if len(need) > 0 {
		return steps, errors.New("unresolved outputs after max iteration depth; probably a bug")
	}
	return steps, nil

}

type runJobFn = func(ctx context.Context, job Job, jc *JobContext, inputs []Artifact) ([]Artifact, error)

func runRealJob(ctx context.Context, job Job, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	var err error
	for i := 0; i < 10; i++ {
		var outputs []Artifact
		outputs, err = job.Run(ctx, jc, inputs)
		if err != nil {
			if errors.Is(err, &Retryable{}) {
				continue
			}
			return nil, errors.Wrapf(err, "job failed after %d attempts", i+1)
		}
		return outputs, nil
	}
	return nil, errors.New("job failed after all retries")
}

func runFakeJob(ctx context.Context, job Job, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	var out []Artifact
	for _, o := range job.Outputs() {
		out = append(out, o)
	}
	return out, nil
}

// runJob starts a job running, eventually sending validated results to doneCh.
func (r *runner) startJob(ctx context.Context, jc *JobContext, job Job, runJob runJobFn, doneCh chan<- jobResult) error {
	inputs, jobs, err := r.resolveJob(job)
	if err != nil {
		return errors.Wrap(err, "unable to resolve job inputs")
	}
	if len(dedupJobs(jobs...)) != 0 {
		return errors.New("job is not ready to run")
	}
	go func() {
		do := func() ([]Artifact, error) {
			outputs, err := runJob(ctx, job, jc, inputs)
			if err != nil {
				return nil, errors.Wrap(err, "job failed")
			}
			declaredOutputs := job.Outputs()
			if got, want := len(outputs), len(declaredOutputs); got != want {
				return nil, errors.Errorf("job %v produced wrong number of outputs; got %v (%v) want %v (%v)", job, got, outputs, want, declaredOutputs)
			}
			var outputErr error
			for i, ref := range declaredOutputs {
				if !ref.Match(outputs[i]) {
					errors.JoinInto(&outputErr, fmt.Errorf("produced artifact %v (index %d) should match %q", outputs[i], i, ref))
				}
			}
			if outputErr != nil {
				return nil, errors.Wrapf(outputErr, "job %v produced invalid output %v", job, outputs)
			}
			return outputs, nil
		}
		outputs, err := do()
		select {
		case <-ctx.Done():
			return
		case doneCh <- jobResult{Job: job, Artifacts: outputs, Err: err}:
		}
	}()
	return nil

}

func Plan(ctx context.Context, jobs []Job, requirements []Reference) ([][]string, error) {
	r := newRunner()
	for _, job := range jobs {
		if err := r.addJob(job); err != nil {
			return nil, err
		}
	}
	return r.Plan(ctx, requirements...)
}
