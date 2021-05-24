package pfsload

import "math/rand"

func shouldExecute(random *rand.Rand, prob float64) bool {
	return random.Float64() < prob
}
