package mybox2d

import (
	"math/rand"
	"testing"

	"github.com/ByteArena/box2d"
)

func BenchmarkWithMap(b *testing.B) {
	var def *box2d.B2FixtureDef
	for n := 0; n < b.N; n++ {
		def = GetCookieFixtureDefByScore(uint64(rand.Intn(fixtureDefsByScoreSize)))
	}
	_ = def
}

func BenchmarkNoMap(b *testing.B) {
	var def *box2d.B2FixtureDef
	for n := 0; n < b.N; n++ {
		def = newCookieFixtureDefByScore(uint64(rand.Intn(fixtureDefsByScoreSize)))
	}
	_ = def
}
