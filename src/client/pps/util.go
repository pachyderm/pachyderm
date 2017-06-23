package pps

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"k8s.io/kubernetes/pkg/api"
	kube "k8s.io/kubernetes/pkg/client/unversioned"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// VisitInput visits each input recursively in ascending order (root last)
func VisitInput(input *Input, f func(*Input)) {
	switch {
	case input.Cross != nil:
		for _, input := range input.Cross {
			VisitInput(input, f)
		}
	case input.Union != nil:
		for _, input := range input.Union {
			VisitInput(input, f)
		}
	}
	f(input)
}

// InputName computes the name of an Input.
func InputName(input *Input) string {
	switch {
	case input.Atom != nil:
		return input.Atom.Name
	case input.Cross != nil:
		if len(input.Cross) > 0 {
			return InputName(input.Cross[0])
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
	VisitInput(input, func(input *Input) {
		SortInputs := func(inputs []*Input) {
			sort.SliceStable(inputs, func(i, j int) bool { return InputName(inputs[i]) < InputName(inputs[j]) })
		}
		switch {
		case input.Cross != nil:
			SortInputs(input.Cross)
		case input.Union != nil:
			SortInputs(input.Union)
		}
	})
}

// InputCommits returns the commits in an Input.
func InputCommits(input *Input) []*pfs.Commit {
	var result []*pfs.Commit
	VisitInput(input, func(input *Input) {
		if input.Atom != nil {
			result = append(result, &pfs.Commit{
				Repo: &pfs.Repo{input.Atom.Repo},
				ID:   input.Atom.Commit,
			})
		}
	})
	return result
}

// PipelineRcName generates the name of the k8s replication controller that
// manages a pipeline's workers
func PipelineRcName(name string, version uint64) string {
	// k8s won't allow RC names that contain upper-case letters
	// or underscores
	// TODO: deal with name collision
	name = strings.Replace(name, "_", "-", -1)
	return fmt.Sprintf("pipeline-%s-v%d", strings.ToLower(name), version)
}

// GetExpectedNumWorkers computes the expected number of workers that
// pachyderm will start given the ParallelismSpec 'spec'.
//
// This is only exported for testing
func GetExpectedNumWorkers(kubeClient *kube.Client, spec *ParallelismSpec) (int, error) {
	coefficient := 0.0 // Used if [spec.Strategy == PROPORTIONAL] or [spec.Constant == 0]
	if spec == nil {
		// Unset ParallelismSpec is handled here. Currently we start one worker per
		// node
		coefficient = 1.0
	} else if spec.Strategy == ParallelismSpec_CONSTANT {
		if spec.Constant > 0 {
			return int(spec.Constant), nil
		}
		// Zero-initialized ParallelismSpec is handled here. Currently we start one
		// worker per node
		coefficient = 1
	} else if spec.Strategy == ParallelismSpec_COEFFICIENT {
		coefficient = spec.Coefficient
	} else {
		return 0, fmt.Errorf("Unable to interpret ParallelismSpec strategy %s", spec.Strategy)
	}
	if coefficient == 0.0 {
		return 0, fmt.Errorf("Ended up with coefficient == 0 (no workers) after interpreting ParallelismSpec %s", spec.Strategy)
	}

	// Start ('coefficient' * 'nodes') workers. Determine number of workers
	nodeList, err := kubeClient.Nodes().List(api.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("unable to retrieve node list from k8s to determine parallelism: %v", err)
	}
	if len(nodeList.Items) == 0 {
		return 0, fmt.Errorf("pachyderm.pps.jobserver: no k8s nodes found")
	}
	result := math.Floor(coefficient * float64(len(nodeList.Items)))
	return int(math.Max(result, 1)), nil
}
