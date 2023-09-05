package jobs

import "context"

type TestJob struct {
	Name      string
	Ins, Outs []Reference
	F         func(context.Context, *JobContext, []Artifact) ([]Artifact, error)
}

var _ Job = (*TestJob)(nil)

func (j *TestJob) String() string       { return j.Name }
func (j *TestJob) Inputs() []Reference  { return j.Ins }
func (j *TestJob) Outputs() []Reference { return j.Outs }
func (j *TestJob) Run(ctx context.Context, jc *JobContext, inputs []Artifact) ([]Artifact, error) {
	return j.F(ctx, jc, inputs)
}
