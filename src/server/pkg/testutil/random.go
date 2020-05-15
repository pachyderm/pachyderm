package testutil

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func SeedRand() string {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	return fmt.Sprint("seed: ", strconv.FormatInt(seed, 10))
}
