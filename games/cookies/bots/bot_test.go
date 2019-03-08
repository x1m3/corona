package bots

import (
	"github.com/stretchr/testify/assert"
	"github.com/x1m3/elixir/games/cookies"
	"github.com/x1m3/elixir/games/cookies/messages"
	"testing"
	"time"
)

type spyAgent struct {
	joinCalls int
	joinResponseCalls int
	CreateCookieCalls int
	CreateCookieResponseCalls int
	MoveCalls int
	UpdateViewWorldCalls int
}

func (a *spyAgent) Join() *messages.UserJoinRequest {
	a.joinCalls++
	return messages.NewUserJoinRequest("manolo")
}

func (a *spyAgent) JoinResponse(response *messages.UserJoinResponse) {
	a.joinResponseCalls++
}

func (a *spyAgent) CreateCookie() *messages.CreateCookieRequest {
	a.CreateCookieCalls++
	return messages.NewCreateCookieRequest()
}

func (a *spyAgent) CreateCookieResponse(response *messages.CreateCookieResponse) {
	a.CreateCookieResponseCalls++
}

func (a *spyAgent) Move() *messages.ViewPortRequest {
	a.MoveCalls++
	return messages.NewViewPortRequest(0,0,1000,1000, 0, false)
}

func (a *spyAgent) UpdateViewWorld(w *messages.ViewportResponse) {
	a.UpdateViewWorldCalls++
}

func TestBot_Run(t *testing.T) {
	game := cookies.New(1000, 1000, 0)
	game.Init()

	spy := &spyAgent{}
	bot := New(game, spy)

	go func() {
		assert.NoError(t, bot.Run())
	}()

	// Let's play the game for some time
	time.Sleep(2 * time.Second)

	assert.Equal(t, spy.joinCalls  ,1)
	assert.Equal(t, spy.joinResponseCalls  ,1)
	assert.Equal(t, spy.CreateCookieCalls  ,1)
	assert.Equal(t, spy.CreateCookieResponseCalls,1)

	assert.NotEqual(t, spy.MoveCalls, 0)
	assert.NotEqual(t, spy.UpdateViewWorldCalls, 0)

	// TODO: Do more testing like ensuring cookie is created and it moves.
}