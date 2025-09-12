package utils

import (
	"math/rand"
	"time"
)

// RandomValue if seedOptional int64 is not provided; time.Now().UnixNano() is used as seed value
func RandomValue(seedOptional ...int64) int {
	r := Random(seedOptional...)
	return r.Intn(900000) + 100000
}

// Random returns a rand instance with seed
func Random(seedOptional ...int64) *rand.Rand {
	var seed int64
	if len(seedOptional) == 0 {
		seed = time.Now().UnixNano()
	} else {
		seed = seedOptional[0]
	}

	return rand.New(rand.NewSource(seed))
}
