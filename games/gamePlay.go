package games

import (
	"github.com/x1m3/elixir/economy"
	"github.com/x1m3/elixir/games/command"
	"github.com/x1m3/elixir/pubsub"
)

type Game interface {
	Init()
	StartGame() command.Response
	ProcessCommand(command.Request) command.Response
	EventListener() <-chan pubsub.Event
	Stop()
}

type GamePlay struct {
	moneyMoverService *economy.ApplyMovementBetweenAccountsService
	game              Game
	events            chan pubsub.Event
}

func NewGamePlay(s *economy.ApplyMovementBetweenAccountsService, game Game) *GamePlay {
	return &GamePlay{moneyMoverService: s, game: game, events: make(chan pubsub.Event)}
}

func (g *GamePlay) Init() command.Response {

	go g.listenAndPublishEvents()

	return g.game.StartGame()
}

func (g *GamePlay) ProcessCommand(c command.Request) command.Response {
	return g.game.ProcessCommand(c)
}

func (g *GamePlay) EventListener() <-chan pubsub.Event {
	return g.events
}

func (g *GamePlay) Stop() {
	g.game.Stop()
}

func (g *GamePlay) listenAndPublishEvents() {
	for got := range g.game.EventListener() {
		if got.Error() != nil {
			g.game.Stop()
			g.events <- got
			return
		}
		switch event := got.(type) {
		case *pubsub.EconomyMovementEvent:
			g.moneyMoverService.Run(event.Wins)
		}
		g.events <- got
	}
}
