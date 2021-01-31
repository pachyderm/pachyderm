package load

import "math/rand"

func shouldExecute(prob float64) bool {
	return rand.Float64() < prob
}
