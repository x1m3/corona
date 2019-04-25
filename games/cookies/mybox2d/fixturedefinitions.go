package mybox2d

import (
	"github.com/ByteArena/box2d"
	"math"
)

const fixtureDefsByScoreSize = 100000

var cookieFixtureDefsByScorePool []box2d.B2FixtureDef
var foodFixtureDef *box2d.B2FixtureDef
var foodBodyDef *box2d.B2BodyDef

func init() {
	cookieFixtureDefsByScorePool = make([]box2d.B2FixtureDef, fixtureDefsByScoreSize)
	for i := uint64(0); i < fixtureDefsByScoreSize; i++ {
		cookieFixtureDefsByScorePool[i] = *newCookieFixtureDefByScore(i)
	}

	foodFixtureDef = newFoodFixtureDef()

	foodBodyDef = newFoodBodyDef()
}

func GetCookieFixtureDefByScore(score uint64) (def *box2d.B2FixtureDef) {
	if score < fixtureDefsByScoreSize {
		return &cookieFixtureDefsByScorePool[score]
	}
	return newCookieFixtureDefByScore(score)
}

func newCookieFixtureDefByScore(score uint64) *box2d.B2FixtureDef {
	// Shape
	shape := box2d.MakeB2CircleShape()
	sc := float64(score)
	shape.M_radius = (math.Log2(sc) + math.Sqrt(sc)) / 2

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 100 * math.Sqrt(sc)
	fd.Restitution = 1
	fd.Friction = 0.1
	return &fd
}

func GetFoodFixtureDef() *box2d.B2FixtureDef {
	return foodFixtureDef
}

func newFoodFixtureDef() *box2d.B2FixtureDef {
	// Shape
	shape := box2d.MakeB2CircleShape()
	shape.M_radius = 1

	// fixture
	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &shape
	fd.Density = 1
	fd.Restitution = 0
	fd.Friction = 1

	fd.Filter.GroupIndex = -1 // Food do not collide

	return &fd
}

func GetFoodBodyDef() *box2d.B2BodyDef {
	return foodBodyDef
}

func newFoodBodyDef() *box2d.B2BodyDef {
	def := box2d.MakeB2BodyDef()
	def.Type = box2d.B2BodyType.B2_dynamicBody
	def.LinearDamping = 1
	def.FixedRotation = true
	def.AllowSleep = true
	return &def
}
