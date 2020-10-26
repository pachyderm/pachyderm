package load

import (
	"math/rand"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

type FuzzOperationSpec struct {
	OperationSpec *OperationSpec `yaml:"operation,omitempty"`
	Prob          float64        `yaml:"prob,omitempty"`
}

func FuzzOperation(c *client.APIClient, repo, commit string, specs []*FuzzOperationSpec, validator ...*Validator) error {
	var totalProb float64
	prob := rand.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return Operation(c, repo, commit, spec.OperationSpec, validator...)
		}
	}
	// TODO: Should not reach here, need to validate fuzzing probabilities (this applies to other fuzzers as well).
	return nil
}

type FuzzFileSpec struct {
	FileSpec *FileSpec `yaml:"file,omitempty"`
	Prob     float64   `yaml:"prob,omitempty"`
}

func FuzzFile(specs []*FuzzFileSpec) (tarutil.File, error) {
	var totalProb float64
	prob := rand.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return File(spec.FileSpec)
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
