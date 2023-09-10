package jobs

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/httpclient"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

// runner runs jobs.
type runner struct {
	Cache *Cache // The artifact cache.

	artifacts     []Artifact
	artifactCache map[Reference]Artifact
	outputsToJobs map[Reference]Job
	jobs          map[uint64]struct{}
}

type jobResult struct {
	Job       Job
	Artifacts []Artifact
	Err       error
}

type RunnerOption struct {
	Artifacts []Artifact
}

func newRunner(opts ...RunnerOption) *runner {
	r := &runner{
		Cache: &Cache{
			Path: "/tmp/builder-cache/",
		},
		outputsToJobs: make(map[Reference]Job),
		artifactCache: make(map[Reference]any),
		jobs:          make(map[uint64]struct{}),
	}
	for _, opt := range opts {
		r.artifacts = append(r.artifacts, opt.Artifacts...)
	}
	return r
}

// resolveInputs builds a plan for generating the desired inputs.  For each element in i, we add the
// artifact to the artifacts slice if we have it, or nil if not, and add the job required to build
// the artifact to job, or nil if we already have the artifact.  len(want) == len(artifacts) ==
// len(jobs).
func (r *runner) resolveInputs(want []Reference) ([]Artifact, []Job, error) {
	var artifacts []Artifact
	var jobs []Job
	var inputErr error
	for _, ref := range want {
		var foundArtifact bool
		for _, a := range r.artifacts {
			if ok := ref.Match(a); ok && foundArtifact {
				errors.JoinInto(&inputErr, fmt.Errorf("input %q matches multiple artifacts", ref))
			} else if ok && !foundArtifact {
				foundArtifact = true
				artifacts = append(artifacts, a)
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
			artifacts = append(artifacts, nil)
		} else {
			jobs = append(jobs, nil)
		}
	}
	if inputErr != nil {
		return artifacts, jobs, errors.Wrapf(inputErr, "cannot resolve inputs; artifacts=%v jobs=%v", artifacts, jobs)
	}
	return artifacts, jobs, nil
}

func dedupJobs(jobs ...Job) []Job {
	result := make(map[uint64]Job)
	for _, j := range jobs {
		if j != nil {
			result[j.ID()] = j
		}
	}
	return maps.Values(result)
}

// AddJob adds a job to potentially run.
func (r *runner) addJob(j Job) error {
	if _, ok := r.jobs[j.ID()]; ok {
		return errors.Errorf("job %v already added (hash collision?)", j)
	}
	inputs, jobs, err := r.resolveInputs(j.Inputs())
	if err != nil {
		// This is not allowed; everything must resolve at the time the job is added.  This
		// avoids cycles.
		return errors.Wrapf(err, "job has unresolvable dependencies [computed inputs: %v, required jobs: %v]; add dependencies before dependents, no cycles", inputs, dedupJobs(jobs...))
	}
	for _, o := range j.Outputs() {
		r.outputsToJobs[o] = j
	}
	r.jobs[j.ID()] = struct{}{}
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

type runJobFn = func(ctx context.Context, job Job, jc *JobContext, inputs []Artifact) ([]Artifact, error)

func runRealJob(ctx context.Context, job Job, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	var err error
	for i := 0; i < 10; i++ {
		var outputs []Artifact
		outputs, err = job.Run(ctx, jc, inputs)
		if err != nil {
			retryable := &Retryable{}
			if errors.As(err, &retryable) {
				select {
				case <-ctx.Done():
					// Don't retry if the context is done.
					return nil, errors.Join(context.Cause(ctx), err)
				case <-time.After(retryable.Delay):
				}
				continue
			}
			s := "s"
			if i == 0 {
				s = ""
			}
			return nil, errors.Wrapf(err, "run failed after %d attempt%s", i+1, s)
		}
		return outputs, nil
	}
	return nil, errors.Wrap(err, "job failed after all retries")
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
	inputs, jobs, err := r.resolveInputs(job.Inputs())
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

type stepReceiver interface {
	Reportf(ctx context.Context, format string, args ...any)
	Next(ctx context.Context)
}

type PlanOutput [][]string

func (p PlanOutput) String() string {
	b := new(strings.Builder)
	for i, paragraph := range p {
		if i == len(p)-1 && len(paragraph) == 0 {
			break
		}
		fmt.Fprintf(b, "step %d:\n", i)
		for _, line := range paragraph {
			fmt.Fprintf(b, "    %v\n", line)
		}
	}
	return b.String()
}

func (p *PlanOutput) Reportf(_ context.Context, format string, args ...any) {
	if len(*p) == 0 {
		*p = append(*p, nil)
	}
	(*p)[len(*p)-1] = append((*p)[len(*p)-1], fmt.Sprintf(format, args...))
}
func (p *PlanOutput) Next(_ context.Context) {
	*p = append(*p, nil)
}

type logStepReceiver struct{}

func (logStepReceiver) Reportf(ctx context.Context, format string, args ...any) {
	log.Debug(pctx.Child(ctx, "", pctx.WithOptions(zap.AddCallerSkip(1))), fmt.Sprintf(format, args...))
}
func (logStepReceiver) Next(_ context.Context) {}

func (r *runner) run(rctx context.Context, want []Reference, runFn runJobFn, s stepReceiver) error {
	if _, _, err := r.resolveInputs(want); err != nil {
		return errors.Wrap(err, "resolve arguments")
	}
	need := map[Reference]struct{}{}
	for _, ref := range want {
		need[ref] = struct{}{}
		s.Reportf(rctx, "initial dependency on %v", ref)
	}
	s.Next(rctx)

	jc := &JobContext{
		allowedPushPrefixes: []string{"http://localhost:"},
		Cache:               r.Cache,
		HTTPClient: &http.Client{
			Transport: &httpclient.RoundTripper{
				RoundTripper: http.DefaultTransport,
			},
		},
	}
	runningJobs := make(map[uint64]Job)
	jobStart := make(map[uint64]time.Time)
	waitingJobs := make(map[uint64]Job)
	resultCh := make(chan jobResult)
	var completed, total int

	// Breadth-first search over the dependency graph.  We start with what we need, add jobs to
	// satisfy those dependencies, until `want` is produced.  After each iteration that doesn't
	// make progress resolving dependencies, we stop and wait for a job to finish.
	runOnce := func(i int) (retErr error) {
		ctx, done := log.SpanContext(rctx, fmt.Sprintf("step.%d", i))
		defer done(log.Errorp(&retErr))
		var madeProgress bool

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// See what we have to start with.
		for ref := range need {
			if _, ok := r.getArtifact(ref); ok {
				s.Reportf(ctx, "yield ref %v", ref)
				delete(need, ref)
				madeProgress = true
			}
		}

		// See which jobs we need to run in order to create the artifacts in "need".
		needJobs := make(map[uint64]Job)
		for ref := range need {
			for product, job := range r.outputsToJobs {
				if ref.Match(product) {
					needJobs[job.ID()] = job
				}
			}
		}

		// See which jobs are possible to run given the current state of "have".
		for _, job := range needJobs {
			if _, ok := runningJobs[job.ID()]; ok {
				// It's already running.
				continue
			}
			_, jobs, err := r.resolveInputs(job.Inputs())
			if err != nil {
				return errors.Wrapf(err, "problem resolving inputs for job %v", job)
			}
			if len(dedupJobs(jobs...)) == 0 {
				// We can run this job right now.  This job uses the root context;
				// it has nothing to do with this step in particular.
				if err := r.startJob(rctx, jc, job, runFn, resultCh); err != nil {
					return errors.Wrapf(err, "problem starting job %v", job)
				}
				total++
				runningJobs[job.ID()] = job
				jobStart[job.ID()] = time.Now()
				delete(waitingJobs, job.ID())
				s.Reportf(ctx, "start job %v", job)
				madeProgress = true
				continue
			}

			if _, ok := waitingJobs[job.ID()]; ok {
				continue
			}

			// The job isn't running, wasn't started, and isn't waiting to run,
			for _, input := range job.Inputs() {
				need[input] = struct{}{}
				waitingJobs[job.ID()] = job
				s.Reportf(ctx, "discovered dependency on %v from %v", input, job)
				madeProgress = true
			}
		}

		// If we didn't do anything this iteration, wait for a job to return.
		if !madeProgress {
			running := maps.Values(runningJobs)
			s.Reportf(ctx, "%v jobs in progress", len(running))
			if len(running) == 0 {
				s.Reportf(ctx, "panic: deadlock; no jobs to wait for")
				return errors.New("panic: deadlock")
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
				var runtime time.Duration
				for _, v := range jobStart {
					runtime = max(runtime, time.Since(v))
				}
				runtime = runtime.Round(time.Millisecond)
				if len(running) == 1 {
					log.Info(rctx, fmt.Sprintf("%v running (%s), %v/%v done", running, runtime, completed, total))
				} else {
					log.Info(rctx, fmt.Sprintf("%v running (%s), %v/%v done", len(running), runtime, completed, total))
				}
			case result := <-resultCh:
				if err := result.Err; err != nil {
					return errors.Wrapf(err, "received error result from job %v", result.Job)
				}
				r.artifacts = append(r.artifacts, result.Artifacts...)
				delete(runningJobs, result.Job.ID())
				delete(jobStart, result.Job.ID())
				s.Reportf(ctx, "job %v finished", result.Job)
				completed++
			}
		}
		s.Next(ctx)
		return nil
	}

	// The 1_000_000 here is the maximum edge count between dependencies plus the total number of
	// jobs plus the total number of outputs.
	for i := 0; i < 1_000_000 && len(need) > 0; i++ {
		if err := runOnce(i); err != nil {
			return err
		}
	}
	if len(need) > 0 {
		return errors.New("unresolved outputs after max iteration depth; probably a bug")
	}
	log.Info(rctx, fmt.Sprintf("%v/%v done", completed, total))
	return nil

}

// Plan returns a human-readable plan of what the job runner will do to build want.
func Plan(ctx context.Context, jobs []Job, want []Reference, opts ...RunnerOption) (plan PlanOutput, _ error) {
	r := newRunner(opts...)
	for _, job := range jobs {
		if err := r.addJob(job); err != nil {
			return nil, errors.Wrapf(err, "add job %v", job)
		}
	}
	err := r.run(ctx, want, runFakeJob, &plan)
	return plan, err
}

// Resolve, using the provided jobs, builds the objects requested in want and returns the built
// artifacts (in the same order as want).
func Resolve(ctx context.Context, jobs []Job, want []Reference, opts ...RunnerOption) ([]Artifact, error) {
	r := newRunner(opts...)
	for _, job := range jobs {
		if err := r.addJob(job); err != nil {
			return nil, errors.Wrapf(err, "add job %v", job)
		}
	}
	if err := r.run(ctx, want, runRealJob, logStepReceiver{}); err != nil {
		return nil, errors.Wrap(err, "run")
	}
	var output []Artifact
	for _, ref := range want {
		a, ok := r.getArtifact(ref)
		if !ok {
			return nil, errors.Errorf("unexpectedly missing artifact; no match for ref %q", ref)
		}
		output = append(output, a)
	}
	return output, nil
}

// Outputs returns the names of all outputs generated by the listed jobs.
func Outputs(jobs []Job) ([]Reference, error) {
	r := newRunner()
	for _, job := range jobs {
		if err := r.addJob(job); err != nil {
			return nil, errors.Wrapf(err, "add job %v", job)
		}
	}
	return maps.Keys(r.outputsToJobs), nil
}
