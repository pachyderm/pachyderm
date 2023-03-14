package random

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
	rand.Seed(seed) //nolint:staticcheck // CORE-1512
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}

// String returns a random string with length n.
func String(n int) string {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck // CORE-1512
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}
