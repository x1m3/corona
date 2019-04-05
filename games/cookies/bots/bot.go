package bots

import (
	"github.com/pkg/errors"
	"github.com/x1m3/elixir/games/cookies"
	"github.com/x1m3/elixir/games/cookies/messages"
	"log"
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
	sessionID uint64
	viewportResponses chan *messages.ViewportResponse
	ticker *time.Ticker
	finish chan bool
}

func New(game *cookies.Game, bot BotAgent) *Bot {
	return &Bot{game: game, agent: bot, ticker: time.NewTicker(250 * time.Millisecond), finish: make(chan bool, 1)}
}

// Run makes a bot to connect to the game and start playing. It should be
// called on its own goroutine (go bot.Run() )
func (b *Bot) Run() error {
	var resp interface{}
	var err error

	// Creating a session
	b.sessionID, b.viewportResponses = b.game.NewSession()

	// Joining step1
	resp, err = b.game.UserJoin(b.sessionID, b.agent.Join())
	if err != nil {
		return err
	}
	b.agent.JoinResponse(resp.(*messages.UserJoinResponse))

	// StartPlaying
	resp, err = b.game.CreateCookie(b.sessionID, b.agent.CreateCookie())
	if err != nil {
		return err
	}
	b.agent.CreateCookieResponse(resp.(*messages.CreateCookieResponse))

	b.game.UpdateViewPortRequest(b.sessionID, b.agent.Move())

	for {
		select {
		case resp, ok := <-b.viewportResponses:
			if !ok {
				return errors.New("bla bla bla")
			}
			b.agent.UpdateViewWorld(resp)
			b.game.UpdateViewPortRequest(b.sessionID, b.agent.Move())

		case <-b.finish:
			b.destroy()
			return nil
		}
	}
}

func (b *Bot) destroy() {
	b.ticker.Stop()
	b.game.Logout(b.sessionID)
	log.Println("Bot disconnected")
}

func (b *Bot) Destroy() {
	b.finish <- true
}
