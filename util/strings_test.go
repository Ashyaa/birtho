package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	a := assert.New(t)
	t.Run("false", func(t *testing.T) {
		a.False(Contains([]string{"abc", "def", "ghi"}, "lorem"))
	})
	t.Run("true", func(t *testing.T) {
		a.True(Contains([]string{"abc", "def", "ghi"}, "abc"))
		a.True(Contains([]string{"abc", "def", "ghi"}, "def"))
		a.True(Contains([]string{"abc", "def", "ghi"}, "ghi"))
	})
}

func TestAppendUnique(t *testing.T) {
	a := assert.New(t)
	list := []string{"abc", "def", "ghi"}
	t.Run("new element", func(t *testing.T) {
		res := AppendUnique(list, "lorem")
		a.Equal(append(list, "lorem"), res)
	})
	t.Run("existing element", func(t *testing.T) {
		a.Equal(list, AppendUnique(list, "abc"))
		a.Equal(list, AppendUnique(list, "def"))
		a.Equal(list, AppendUnique(list, "ghi"))
	})
}

func TestIndex(t *testing.T) {
	a := assert.New(t)
	list := []string{"abc", "def", "ghi"}
	t.Run("element doesn't exist", func(t *testing.T) {
		a.Equal(-1, Index(list, "lorem"))
	})
	t.Run("existing element", func(t *testing.T) {
		a.Equal(0, Index(list, "abc"))
		a.Equal(1, Index(list, "def"))
		a.Equal(2, Index(list, "ghi"))
	})
}

func TestRemove(t *testing.T) {
	a := assert.New(t)
	list := []string{"abc", "def", "ghi"}
	t.Run("element doesn't exist", func(t *testing.T) {
		a.Equal(list, Remove(list, "lorem"))
	})
	t.Run("element exists", func(t *testing.T) {
		a.Equal([]string{"def", "ghi"}, Remove(list, "abc"))
		a.Equal([]string{"abc", "ghi"}, Remove(list, "def"))
		a.Equal([]string{"abc", "def"}, Remove(list, "ghi"))
	})
}
