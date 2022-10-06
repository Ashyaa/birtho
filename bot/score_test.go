package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderBoardSort(t *testing.T) {
	a := assert.New(t)
	lb := Leaderboard{
		{
			UID:   "1111",
			Name:  "lorem",
			Score: 123,
			Rank:  "1st",
		},
		{
			UID:   "2222",
			Name:  "ipsum",
			Score: 120,
			Rank:  "3rd",
		},
		{
			UID:   "3333",
			Name:  "dolor",
			Score: 230,
			Rank:  "2nd",
		},
	}
	exp := Leaderboard{
		{
			UID:   "3333",
			Name:  "dolor",
			Score: 230,
			Rank:  "1st",
		},
		{
			UID:   "1111",
			Name:  "lorem",
			Score: 123,
			Rank:  "2nd",
		},
		{
			UID:   "2222",
			Name:  "ipsum",
			Score: 120,
			Rank:  "3rd",
		},
	}
	lb.sort()
	a.Equal(exp, lb)
}
