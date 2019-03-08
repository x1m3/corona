package bots

import (
	"github.com/x1m3/elixir/games/cookies"
	"github.com/x1m3/elixir/games/cookies/messages"
	"time"
)

type BotAgent interface {
	Join() *messages.UserJoinRequest
	JoinResponse(response *messages.UserJoinResponse)
	CreateCookie() *messages.CreateCookieRequest
	CreateCookieResponse(response *messages.CreateCookieResponse)
	Move() *messages.ViewPortRequest
	UpdateViewWorld(w *messages.ViewportResponse)
}

type Bot struct {
	game   *cookies.Game
	agent  BotAgent
	finish chan bool
}

func New(game *cookies.Game, bot BotAgent) *Bot {
	return &Bot{game: game, agent: bot}
}

// Run makes a bot to connect to the game and start playing. It should be
// called on its own goroutine (go bot.Run() )
func (b *Bot) Run() error {
	var resp interface{}
	var err error

	// Creating a session
	sessionID := b.game.NewSession()

	// Joining step1
	resp, err = b.game.UserJoin(sessionID, b.agent.Join())
	if err != nil {
		return err
	}
	b.agent.JoinResponse(resp.(*messages.UserJoinResponse))

	// StartPlaying
	resp, err = b.game.CreateCookie(sessionID, b.agent.CreateCookie())
	if err != nil {
		return err
	}
	b.agent.CreateCookieResponse(resp.(*messages.CreateCookieResponse))

	b.game.UpdateViewPortRequest(sessionID, b.agent.Move())

	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			resp, err := b.game.ViewPortRequest(sessionID)
			if err != nil {
				return err
			}
			b.agent.UpdateViewWorld(resp)
			b.game.UpdateViewPortRequest(sessionID, b.agent.Move())

		case <-b.finish:
			// TODO: Do more things to close properly
			return nil
		}
	}
}

func (b *Bot) Destroy() {
	b.finish <- true
}
