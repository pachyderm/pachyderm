package load

import (
	"math/rand"
)

type FuzzOperationSpec struct {
	OperationSpec *OperationSpec `yaml:"operation,omitempty"`
	Prob          float64        `yaml:"prob,omitempty"`
}

func FuzzOperation(env *Env, repo, branch, commit string, specs []*FuzzOperationSpec) error {
	var totalProb float64
	prob := rand.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return Operation(env, repo, branch, commit, spec.OperationSpec)
		}
	}
	// TODO: Should not reach here, need to validate fuzzing probabilities (this applies to other fuzzers as well).
	return nil
}

type FuzzFileSpec struct {
	FileSpec *FileSpec `yaml:"file,omitempty"`
	Prob     float64   `yaml:"prob,omitempty"`
}

func FuzzFile(env *Env, specs []*FuzzFileSpec) (*MemFile, error) {
	var totalProb float64
	prob := rand.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return File(env, spec.FileSpec)
		}
	}
	return nil, nil
}

type FuzzSizeSpec struct {
	SizeSpec *SizeSpec `yaml:"size,omitempty"`
	Prob     float64   `yaml:"prob,omitempty"`
}

func FuzzSize(specs []*FuzzSizeSpec) *SizeSpec {
	var totalProb float64
	prob := rand.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return spec.SizeSpec
		}
	}
	return nil
}
