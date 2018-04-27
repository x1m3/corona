package slots

import (
	"github.com/x1m3/elixir/games/slots/engine"
	"github.com/x1m3/elixir/random"
	"github.com/x1m3/elixir/games/command"
)

type FruitsSlot struct {
	engine *engine.Slot3X24
	w1     [24]int8
	w2     [24]int8
	w3     [24]int8
}

func NewFruitSlot(randGenerator random.Generator) *FruitsSlot {

	w1 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w2 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w3 := [...]int8{1, 2, 3, 0, 4, 5, 6, 0, 7, 3, 9, 0, 1, 2, 3, 0, 4, 5, 6, 0, 7, 8, 9, 0}

	prizes := engine.NewWinTable3Wheels()
	prizes.Add(engine.NewWinRule3Wheels(1, 1, 1, 500))
	prizes.Add(engine.NewWinRule3Wheels(2, 2, 2, 200))
	prizes.Add(engine.NewWinRule3Wheels(3, 3, 3, 100))
	prizes.Add(engine.NewWinRule3Wheels(4, 4, 4, 50))
	prizes.Add(engine.NewWinRule3Wheels(5, 5, 5, 20))
	prizes.Add(engine.NewWinRule3Wheels(6, 6, 6, 10))
	prizes.Add(engine.NewWinRule3Wheels(7, 7, 7, 10))
	prizes.Add(engine.NewWinRule3Wheels(8, 8, 8, 10))
	prizes.Add(engine.NewWinRule3Wheels(9, 9, 9, 10))
	prizes.Add(engine.NewWinRule3Wheels(1, 1, engine.ANY_VALUE, 5))
	prizes.Add(engine.NewWinRule3Wheels(2, 2, engine.ANY_VALUE, 5))
	prizes.Add(engine.NewWinRule3Wheels(3, 3, engine.ANY_VALUE, 5))
	prizes.Add(engine.NewWinRule3Wheels(4, 4, engine.ANY_VALUE, 3))
	prizes.Add(engine.NewWinRule3Wheels(5, 5, engine.ANY_VALUE, 3))
	prizes.Add(engine.NewWinRule3Wheels(6, 6, engine.ANY_VALUE, 3))
	prizes.Add(engine.NewWinRule3Wheels(7, 7, engine.ANY_VALUE, 2))
	prizes.Add(engine.NewWinRule3Wheels(8, 8, engine.ANY_VALUE, 2))
	prizes.Add(engine.NewWinRule3Wheels(9, 9, engine.ANY_VALUE, 2))
	prizes.Add(engine.NewWinRule3Wheels(1, engine.ANY_VALUE, engine.ANY_VALUE, 2))
	prizes.Add(engine.NewWinRule3Wheels(engine.ANY_VALUE, 1, engine.ANY_VALUE, 2))
	prizes.Add(engine.NewWinRule3Wheels(engine.ANY_VALUE, engine.ANY_VALUE, 1, 2))

	machine := &FruitsSlot{}
	machine.engine = engine.New3x24(randGenerator, w1, w2, w3, prizes)

	return machine
}

func (s *FruitsSlot) Init() {}

func (s *FruitsSlot) ProcessCommand(c command.Request) command.Response {
	switch c.(type) {

	case *command.Slot3SpinRequest:
		return s.spinAll(c.(*command.Slot3SpinRequest))

	default:
		return nil
	}
}

func (s *FruitsSlot) StartGame() *command.Slot3InitResponse {
	r := &command.Slot3InitResponse{}
	r.Wheel1, r.Wheel2, r.Wheel3, r.P1, r.P2, r.P3 = s.engine.Init()
	return r
}

func (s *FruitsSlot) spinAll(c *command.Slot3SpinRequest) *command.Slot3SpinResponse {
	r := &command.Slot3SpinResponse{}
	r.Win, r.P1, r.P2, r.P3 = s.engine.SpinAll(c.Bet)
	return r
}
