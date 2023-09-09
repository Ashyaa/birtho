package util_test

import (
	"testing"

	"github.com/ashyaa/birtho/util"
	"github.com/stretchr/testify/assert"
)

func TestProbability(t *testing.T) {
	a := assert.New(t)
	rng := util.NewRNG()

	for i := 0; i < 10000; i++ {
		n := rng.Intn(5)
		a.True(n >= 0)
		a.True(n < 5)
	}
}
