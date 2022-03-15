package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripChannelTag(t *testing.T) {
	a := assert.New(t)
	t.Run("invalid", func(t *testing.T) {
		res, ok := StripChannelTag("toto")
		a.False(ok)
		a.Equal("", res)
	})
	t.Run("nominal", func(t *testing.T) {
		res, ok := StripChannelTag("<#739949027776266260>")
		a.True(ok)
		a.Equal("739949027776266260", res)
	})
}

func TestBuildChannelTag(t *testing.T) {
	a := assert.New(t)
	t.Run("nominal", func(t *testing.T) {
		res := BuildChannelTag("739949027776266260")
		a.Equal("<#739949027776266260>", res)
	})
}
