package util

import (
	"math/rand"
	"time"
)

const (
	factor  int = 1000
	hundred int = 100 * factor
)

type RNG struct {
	*rand.Rand
}

// NewRNG returns a new Random Number Generator
func NewRNG() RNG {
	return RNG{rand.New(rand.NewSource(time.Now().UTC().UnixNano()))}
}

// PercentChance generates a random number generator, and returns true if it is strictly inferior
// to 100. In other terms, it randomizes a `Rate%` probability.
func (r RNG) PercentChance(rate int) bool {
	if rate < 0 {
		return false
	}
	if rate >= 100 {
		return true
	}
	n := r.Intn(hundred)
	return n < rate*factor
}
