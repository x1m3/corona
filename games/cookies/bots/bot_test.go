package bots

import (
	"github.com/stretchr/testify/assert"
	"github.com/x1m3/elixir/games/cookies"
	"github.com/x1m3/elixir/games/cookies/messages"
	"sync"
	"testing"
	"time"
)

type spyAgent struct {
	sync.Mutex
	joinCalls int
	joinResponseCalls int
	CreateCookieCalls int
	CreateCookieResponseCalls int
	MoveCalls int
	UpdateViewWorldCalls int
}

func (a *spyAgent) Join() *messages.UserJoinRequest {
	a.Lock()
	a.joinCalls++
	a.Unlock()
	return messages.NewUserJoinRequest("manolo")
}

func (a *spyAgent) JoinResponse(response *messages.UserJoinResponse) {
	a.Lock()
	a.joinResponseCalls++
	a.Unlock()
}

func (a *spyAgent) CreateCookie() *messages.CreateCookieRequest {
	a.Lock()
	a.CreateCookieCalls++
	a.Unlock()
	return messages.NewCreateCookieRequest()
}

func (a *spyAgent) CreateCookieResponse(response *messages.CreateCookieResponse) {
	a.Lock()
	a.CreateCookieResponseCalls++
	a.Unlock()
}

func (a *spyAgent) Move() *messages.ViewPortRequest {
	a.Lock()
	a.MoveCalls++
	a.Unlock()
	return messages.NewViewPortRequest(0,0,1000,1000, 0, false)
}

func (a *spyAgent) UpdateViewWorld(w *messages.ViewportResponse) {
	a.Lock()
	a.UpdateViewWorldCalls++
	a.Unlock()
}

func TestBot_Run(t *testing.T) {
	game := cookies.New(1000, 1000, 1*time.Millisecond)
	game.Init()

	spy := &spyAgent{}
	bot := New(game, spy)

	go func() {
		assert.NoError(t, bot.Run())
	}()

	// Let's play the game for some time
	time.Sleep(2 * time.Second)

	spy.Lock()
	assert.Equal(t, spy.joinCalls  ,1)
	assert.Equal(t, spy.joinResponseCalls  ,1)
	assert.Equal(t, spy.CreateCookieCalls  ,1)
	assert.Equal(t, spy.CreateCookieResponseCalls,1)

	assert.NotEqual(t, spy.MoveCalls, 0)
	assert.NotEqual(t, spy.UpdateViewWorldCalls, 0)
	spy.Unlock()

	// TODO: Do more testing like ensuring cookie is created and it moves.
}