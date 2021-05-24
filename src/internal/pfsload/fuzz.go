package pfsload

import "math/rand"

func FuzzOperation(env *Env, repo, branch, commit string, specs []*OperationSpec) error {
	var totalProb float64
	prob := env.Rand().Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return Operation(env, repo, branch, commit, spec)
		}
	}
	// TODO: Should not reach here, need to validate fuzzing probabilities (this applies to other fuzzers as well).
	return nil
}

func FuzzFile(env *Env, specs []*FileSpec) (*MemFile, error) {
	var totalProb float64
	prob := env.Rand().Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return File(env, spec)
		}
	}
	return nil, nil
}

func FuzzSize(specs []*SizeSpec, random *rand.Rand) *SizeSpec {
	var totalProb float64
	prob := random.Float64()
	for _, spec := range specs {
		totalProb += spec.Prob
		if prob <= totalProb {
			return spec
		}
	}
	return nil
}
