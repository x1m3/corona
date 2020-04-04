package leaderboard

import (
	"github.com/wangjia184/sortedset"
	"strconv"
)

type Item struct {
	ID    uint64
	Score int64
	Value interface{}
}

type LeaderBoard struct {
	sortedSet *sortedset.SortedSet
}

func New() *LeaderBoard {
	return &LeaderBoard{
		sortedSet: sortedset.New(),
	}
}

func (b *LeaderBoard) AddOrUpdate(ID uint64, score int64, value interface{}) {
	id := strconv.FormatUint(ID, 10)
	b.sortedSet.AddOrUpdate(id, sortedset.SCORE(score), value)
}

func (b *LeaderBoard) Count() int {
	return b.sortedSet.GetCount()
}

func (b *LeaderBoard) GetByID(ID uint64) *Item {
	id := strconv.FormatUint(ID, 10)
	n := b.sortedSet.GetByKey(id)
	if n == nil {
		return nil
	}
	return &Item{ID: ID, Score: int64(n.Score()), Value: n.Value}
}

func (b *LeaderBoard) OrderedByRanking(start int, end int) []*Item {
	ldb := make([]*Item, 0)
	for _, v := range b.sortedSet.GetByRankRange(start, end, false) {
		id, _ := strconv.ParseUint(v.Key(),10,64)
		ldb = append(ldb, &Item{ID:id, Score: int64(v.Score()), Value: v.Value})
	}
	return ldb
}


func (b *LeaderBoard) Remove(ID uint64) {
	id := strconv.FormatUint(ID, 10)
	b.sortedSet.Remove(id)
}
