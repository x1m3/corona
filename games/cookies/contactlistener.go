package cookies

import (
	"github.com/ByteArena/box2d"
	"log"
)

type contactListener struct {
}

func (l *contactListener) BeginContact(contact box2d.B2ContactInterface) {
	body1 := contact.GetFixtureA().GetBody()
	body2 := contact.GetFixtureB().GetBody()

	data1 := body1.GetUserData()
	data2 := body2.GetUserData()

	// Collission between cookies
	if cookie1, isACookie := data1.(*Cookie); isACookie {
		if cookie2, isACookie := data2.(*Cookie); isACookie {
			l.collisionBetweenCookies(cookie1, cookie2)
			return
		}
	}

	// Collission between cookie and food
	if cookie, isACookie := data1.(*Cookie); isACookie {
		if food, isFood := data2.(*Food); isFood {
			l.collisionBetweenCookiesAndFood(cookie, food)
			return
		}
	}

	// Collission between food and cookie
	if food, isFood := data1.(*Food); isFood {
		if cookie, isACookie := data2.(*Cookie); isACookie {
			l.collisionBetweenCookiesAndFood(cookie, food)
			return
		}
	}
}

func (l *contactListener) EndContact(contact box2d.B2ContactInterface) {
	return
}

func (l *contactListener) PreSolve(contact box2d.B2ContactInterface, oldManifold box2d.B2Manifold) {
	return
}

func (l *contactListener) PostSolve(contact box2d.B2ContactInterface, impulse *box2d.B2ContactImpulse) {
	return
}

func (l *contactListener) collisionBetweenCookies(cookie1 *Cookie, cookie2 *Cookie) {
	log.Printf("Collission between cookies <%d> and <%d>", cookie1.ID, cookie2.ID)
}

func (l *contactListener) collisionBetweenCookiesAndFood(cookie *Cookie, food *Food) {
	log.Printf("Collission between cookie<%d> and food <%d>", cookie.ID, food.ID)
}

func newContactListener() *contactListener {
	return &contactListener{}
}
