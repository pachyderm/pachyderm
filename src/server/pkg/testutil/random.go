package testutil

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// SeedRand seeds the rand package with a seed based on the time and returns
// a string with the seed embedded in it.
func SeedRand(customSeed ...int64) string {
	seed := time.Now().UTC().UnixNano()
	if len(customSeed) > 0 {
		seed = customSeed[0]
	}
	rand.Seed(seed)
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}
