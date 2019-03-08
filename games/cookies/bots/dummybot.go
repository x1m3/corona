package bots

import (
	"github.com/x1m3/elixir/games/cookies/messages"
	"log"
	"math"
	"math/rand"
)

type dummyBotAgent struct {
	myInfo *messages.CookieInfo
	viewportWidth float32
	viewportHeight float32
	pendingTurnSteps int
	currentAngle float32
	desiredAngle float32
}

func NewDummyBotAgent(vWidth, vHeight float32) *dummyBotAgent {
	return &dummyBotAgent{viewportWidth: vWidth, viewportHeight:vHeight}
}

func (b *dummyBotAgent) Join() *messages.UserJoinRequest {
	return messages.NewUserJoinRequest("manolo")
}

func (b *dummyBotAgent) JoinResponse(response *messages.UserJoinResponse) {
}

func (b *dummyBotAgent) CreateCookie() *messages.CreateCookieRequest {
	return messages.NewCreateCookieRequest()
}

func (b *dummyBotAgent) CreateCookieResponse(response *messages.CreateCookieResponse) {
	b.myInfo = &response.Data
}

func (b *dummyBotAgent) Move() *messages.ViewPortRequest {
	if b.pendingTurnSteps<0 {

		b.pendingTurnSteps = rand.Intn(100)
		b.desiredAngle = float32(math.Mod(rand.Float64(), 2 * math.Pi)) * 2 * math.Pi
		log.Printf("changing direction to <%f>", b.desiredAngle)
	}
	b.pendingTurnSteps--

	x := b.myInfo.X - b.viewportWidth /2
	y := b.myInfo.Y - b.viewportHeight /2
	xx := b.myInfo.X + b.viewportWidth /2
	yy := b.myInfo.Y + b.viewportHeight /2

	b.currentAngle = float32(math.Mod(float64((b.currentAngle - b.desiredAngle)/2), 2 * math.Pi))

	return messages.NewViewPortRequest(x, y, xx, yy, b.currentAngle, false)
}

func (b *dummyBotAgent) UpdateViewWorld(w *messages.ViewportResponse) {}

