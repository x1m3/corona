package test

import (
	"github.com/x1m3/elixir/pubsub"
	"github.com/x1m3/elixir/games/command"
	"github.com/x1m3/elixir/economy"
	"github.com/x1m3/elixir/games"
	"github.com/nu7hatch/gouuid"
	"github.com/x1m3/elixir/economy/repos/memory"
)

type FakeGame struct {
	events chan pubsub.Event
}

func newFakeGame() *FakeGame {
	return &FakeGame{
		events: make(chan pubsub.Event),
	}
}

func (g *FakeGame) Init() command.Response {
	g.events <- &traceEvent{Msg:"INIT"}
	return nil
}

func (g *FakeGame) ProcessCommand(command.Request) command.Response {
	g.events <- &traceEvent{Msg:"PROCESS_COMMAND"}
	return nil
}

func (g *FakeGame) EventListener() <-chan pubsub.Event {
	return g.events
}

func (g *FakeGame) Stop() {
	close(g.events)
}

type traceEvent struct {
	Msg string
}

func (e *traceEvent) Topic() string { return "traceEvent"}
func (e *traceEvent) Error() error { return nil}
func (e *traceEvent) Priority() int {return pubsub.MedPriority}

func helperInitGamePlayTest(bankMoney, userMoney int64) (bankAccount *economy.Account, userAccount *economy.Account, gamePlay *games.GamePlay) {

	id, _ := uuid.NewV4()
	bankAccount = economy.NewAccount(id, economy.NewMoney(bankMoney, "coins"))
	id, _ = uuid.NewV4()
	userAccount = economy.NewAccount(id, economy.NewMoney(userMoney, "coins"))

	gamePlay = games.NewGamePlay(economy.NewApplyMovementBetweenAccountsService(bankAccount, userAccount, memoryRepo.NewAccountRepo()), newFakeGame())

	return bankAccount, userAccount, gamePlay
}
