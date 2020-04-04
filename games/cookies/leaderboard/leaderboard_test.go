package leaderboard_test

import (
	"github.com/stretchr/testify/assert"

	"sync"
	"testing"

	"github.com/x1m3/corona/games/cookies/leaderboard"
)

func TestLeaderBoard_Case1(t *testing.T) {

	sortedset := leaderboard.New()
	sortedset.AddOrUpdate(1, 89, "Kelly")
	sortedset.AddOrUpdate(2, 100, "Staley")
	sortedset.AddOrUpdate(3, 100, "Jordon")
	sortedset.AddOrUpdate(4, -321, "Park")
	sortedset.AddOrUpdate(5, 101, "Albert")
	sortedset.AddOrUpdate(6, 99, "Lyman")
	sortedset.AddOrUpdate(7, 99, nil)
	sortedset.AddOrUpdate(8, 70, "Audrey")

	sortedset.AddOrUpdate(5, 99, sync.Mutex{}) // Test a random interface{}

	sortedset.Remove(2)

	// Lower to greater
	assert.Equal(t, 7, sortedset.Count())

	nodes := sortedset.OrderedByRanking(1, -1)
	assert.Equal(t, sortedset.Count(), len(nodes))

	lastScore := int64(-100000)
	for i, n := range nodes {
		if lastScore > n.Score {
			t.Errorf("Leaderboad has wrong order. Check element %d", i)
		}
		lastScore = n.Score
	}

	// Greater to lower
	nodes = sortedset.OrderedByRanking(-1, 1)
	assert.Equal(t, sortedset.Count(), len(nodes))

	lastScore = int64(100000)
	for i, n := range nodes {
		if lastScore < n.Score {
			t.Errorf("Leaderboad has wrong order. Check element %d", i)
		}
		lastScore = n.Score
	}
}
