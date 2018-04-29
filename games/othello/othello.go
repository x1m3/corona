package othello

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/x1m3/elixir/games/command"
)

const EMPTY = 0
const BLACK = 1
const WHITE = 2

const WIDTH = 8
const HEIGHT = 8

type Game struct {
	board board
}

func (g *Game) Init() {
	g.board.Init()
}

func (g *Game) StartGame() command.Response {
	return nil
}

func (g *Game) ProcessCommand(command.Request) command.Response {
	panic("implement me")
}

func (g *Game) EventListener() <-chan pubsub.Event {
	panic("implement me")
}

func (g *Game) Stop() {
	panic("implement me")
}

