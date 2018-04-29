package games

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/x1m3/elixir/games/command"
)

type Game interface {
	Init()
	StartGame() command.Response
	ProcessCommand(command.Request) command.Response
	EventListener() <-chan pubsub.Event
	Stop()
}
