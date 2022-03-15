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

func TestStripUserTag(t *testing.T) {
	a := assert.New(t)
	t.Run("invalid", func(t *testing.T) {
		res, ok := StripUserTag("lorem")
		a.False(ok)
		a.Equal("", res)
	})
	t.Run("nominal", func(t *testing.T) {
		res, ok := StripUserTag("<@!951792639001366558>")
		a.True(ok)
		a.Equal("951792639001366558", res)
	})
}

func TestBuildUserTag(t *testing.T) {
	a := assert.New(t)
	t.Run("nominal", func(t *testing.T) {
		res := BuildUserTag("951792639001366558")
		a.Equal("<@!951792639001366558>", res)
	})
}
