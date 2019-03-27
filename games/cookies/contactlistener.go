package cookies

import (
	"github.com/ByteArena/box2d"
	"time"
)

type contactListener struct {
	chColl2Cookies   chan *collision2CookiesDTO
	chCollCookieFood chan *collissionCookieFoodDTO
}

func newContactListener(chCkCk chan *collision2CookiesDTO, chCkFd chan *collissionCookieFoodDTO) *contactListener {
	return &contactListener{chColl2Cookies: chCkCk, chCollCookieFood: chCkFd}
}

func (l *contactListener) BeginContact(contact box2d.B2ContactInterface) {

}

func (l *contactListener) EndContact(contact box2d.B2ContactInterface) {
	body1 := contact.GetFixtureA().GetBody()
	body2 := contact.GetFixtureB().GetBody()

	data1 := body1.GetUserData()
	data2 := body2.GetUserData()

	// Contact between cookies
	if cookie1, isACookie := data1.(*Cookie); isACookie {
		if cookie2, isACookie := data2.(*Cookie); isACookie {
			l.contactBetweenCookies(cookie1, cookie2)
			return
		}
	}

	// Contact between cookie and food
	if cookie, isACookie := data1.(*Cookie); isACookie {
		if food, isFood := data2.(*Food); isFood {
			l.contactBetweenCookiesAndFood(cookie, food)
			return
		}
	}

	// Contact between food and cookie
	if food, isFood := data1.(*Food); isFood {
		if cookie, isACookie := data2.(*Cookie); isACookie {
			l.contactBetweenCookiesAndFood(cookie, food)
			return
		}
	}
}

func (l *contactListener) PreSolve(contact box2d.B2ContactInterface, oldManifold box2d.B2Manifold) {
	return
}

func (l *contactListener) PostSolve(contact box2d.B2ContactInterface, impulse *box2d.B2ContactImpulse) {
	return
}

func (l *contactListener) contactBetweenCookies(cookie1 *Cookie, cookie2 *Cookie) {
	l.chColl2Cookies <- &collision2CookiesDTO{cookie1: cookie1, cookie2: cookie2}
}

func (l *contactListener) contactBetweenCookiesAndFood(cookie *Cookie, food *Food) {
	if time.Since(food.createdOn) > 1000*time.Millisecond {
		l.chCollCookieFood <- &collissionCookieFoodDTO{cookie: cookie, food: food}
	}
}
