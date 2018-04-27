package pubsub

import (
	"github.com/x1m3/elixir/economy"
	"reflect"
)

type EconomyMovementEvent struct {
	Wins *economy.Money
}

func (e *EconomyMovementEvent) Topic() string {
	return reflect.TypeOf(e).String()
}

func (e *EconomyMovementEvent) Priority() int {
	return HighPriority
}

func (e *EconomyMovementEvent) Error() error {
	return nil
}
