package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderBoardSort(t *testing.T) {
	a := assert.New(t)
	lb := leaderboard{
		{
			userID: "1111",
			user:   "lorem",
			score:  123,
			rank:   "1st",
		},
		{
			userID: "2222",
			user:   "ipsum",
			score:  120,
			rank:   "3rd",
		},
		{
			userID: "3333",
			user:   "dolor",
			score:  230,
			rank:   "2nd",
		},
	}
	lb.sort()
	a.Equal(230, lb[0].score)
	a.Equal("1st", lb[0].rank)
	a.Equal(123, lb[1].score)
	a.Equal("2nd", lb[1].rank)
	a.Equal(120, lb[2].score)
	a.Equal("3rd", lb[2].rank)
}
