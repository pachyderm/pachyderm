package pps

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/pfs"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
)

var (
	// format strings for state name parsing errors
	errInvalidJobStateName      string
	errInvalidPipelineStateName string
)

func init() {
	// construct error messages from all current job and pipeline state names
	var states []string
	for i := int32(0); JobState_name[i] != ""; i++ {
		states = append(states, strings.ToLower(strings.TrimPrefix(JobState_name[i], "JOB_")))
	}
	errInvalidJobStateName = fmt.Sprintf("state %%s must be one of %s, or %s, etc", strings.Join(states, ", "), JobState_name[0])
	states = states[:0]
	for i := int32(0); PipelineState_name[i] != ""; i++ {
		states = append(states, strings.ToLower(strings.TrimPrefix(PipelineState_name[i], "PIPELINE_")))
	}
	errInvalidPipelineStateName = fmt.Sprintf("state %%s must be one of %s, or %s, etc", strings.Join(states, ", "), PipelineState_name[0])
}

func (j *Job) String() string {
	return fmt.Sprintf("%s@%s", j.Pipeline.Name, j.ID)
}

// VisitInput visits each input recursively in ascending order (root last)
func VisitInput(input *Input, f func(*Input) error) error {
	err := visitInput(input, f)
	if err != nil && errors.Is(err, errutil.ErrBreak) {
		return nil
	}
	return err
}

func visitInput(input *Input, f func(*Input) error) error {
	var source []*Input
	switch {
	case input == nil:
		return nil // Spouts may have nil input
	case input.Cross != nil:
		source = input.Cross
	case input.Join != nil:
		source = input.Join
	case input.Group != nil:
		source = input.Group
	case input.Union != nil:
		source = input.Union
	}
	for _, input := range source {
		if err := visitInput(input, f); err != nil {
			return err
		}
	}
	return f(input)
}

// InputName computes the name of an Input.
func InputName(input *Input) string {
	switch {
	case input == nil:
		return ""
	case input.Pfs != nil:
		return input.Pfs.Name
	case input.Cross != nil:
		if len(input.Cross) > 0 {
			return InputName(input.Cross[0])
		}
	case input.Join != nil:
		if len(input.Join) > 0 {
			return InputName(input.Join[0])
		}
	case input.Group != nil:
		if len(input.Group) > 0 {
			return InputName(input.Group[0])
		}
	case input.Union != nil:
		if len(input.Union) > 0 {
			return InputName(input.Union[0])
		}
	}
	return ""
}

// SortInput sorts an Input.
func SortInput(input *Input) {
	VisitInput(input, func(input *Input) error {
		SortInputs := func(inputs []*Input) {
			sort.SliceStable(inputs, func(i, j int) bool { return InputName(inputs[i]) < InputName(inputs[j]) })
		}
		switch {
		case input.Cross != nil:
			SortInputs(input.Cross)
		case input.Join != nil:
			SortInputs(input.Join)
		case input.Group != nil:
			SortInputs(input.Group)
		case input.Union != nil:
			SortInputs(input.Union)
		}
		return nil
	})
}

// InputBranches returns the branches in an Input.
func InputBranches(input *Input) []*pfs.Branch {
	var result []*pfs.Branch
	VisitInput(input, func(input *Input) error {
		if input.Pfs != nil {
			result = append(result, &pfs.Branch{
				Repo: &pfs.Repo{
					Name: input.Pfs.Repo,
					Type: input.Pfs.RepoType,
				},
				Name: input.Pfs.Branch,
			})
		}
		if input.Cron != nil {
			result = append(result, &pfs.Branch{
				Repo: &pfs.Repo{
					Name: input.Cron.Repo,
					Type: pfs.UserRepoType,
				},
				Name: "master",
			})
		}
		return nil
	})
	return result
}

// JobStateFromName attempts to interpret a string as a
// JobState, accepting either the enum names or the pretty printed state
// names
func JobStateFromName(name string) (JobState, error) {
	canonical := "JOB_" + strings.TrimPrefix(strings.ToUpper(name), "JOB_")
	if value, ok := JobState_value[canonical]; ok {
		return JobState(value), nil
	}
	return 0, fmt.Errorf(errInvalidJobStateName, name)
}

// PipelineStateFromName attempts to interpret a string as a PipelineState,
// accepting either the enum names or the pretty printed state names
func PipelineStateFromName(name string) (PipelineState, error) {
	canonical := "PIPELINE_" + strings.TrimPrefix(strings.ToUpper(name), "PIPELINE_")
	if value, ok := PipelineState_value[canonical]; ok {
		return PipelineState(value), nil
	}
	return 0, fmt.Errorf(errInvalidPipelineStateName, name)
}
