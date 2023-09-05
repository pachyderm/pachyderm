package jobs

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type Artifact = any

// Job is work to do.
type Job interface {
	// Inputs are references to the inputs that Run desires.
	Inputs() []Reference
	// Outputs reference the objects produced by Run.  Every actual output must be referenced by
	// exactly one Outputs reference.
	Outputs() []Reference
	// Run produces the outputs from the inputs.
	Run(context.Context, *JobContext, []Artifact) ([]Artifact, error)
}

// JobContext is runtime information about the job system.
type JobContext struct {
	Cache      *Cache
	HTTPClient *http.Client
}

// PrintJob prints out a dump of a job.
func PrintJob(w io.Writer, job Job) {
	fmt.Fprintf(w, "job of type %T %v:\n", job, job)
	for i, in := range job.Inputs() {
		fmt.Fprintf(w, "    %d: %v\n", i, in)
	}
	fmt.Fprintf(w, " outputs ->\n")
	for i, out := range job.Outputs() {
		fmt.Fprintf(w, "    %d: %v\n", i, out)
	}
}
