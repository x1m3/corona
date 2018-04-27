package test

import (
	"testing"
	"github.com/x1m3/elixir/random"
	"github.com/x1m3/elixir/games/command"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/x1m3/elixir/games/slots"
)

func TestFruitInit(t *testing.T) {
	r := random.GetSimpleRandomGenerator()
	slot := slots.NewFruitSlot(r)

	response := slot.Init()
	if got, expected := response.Type(), command.SLOT3_INIT_RESPONSE; got != expected {
		t.Errorf("Wrong response time for a Slot3InitRequest. Expecting response of type %v, got %v", expected, got)
		spew.Dump(response)
	}
}

func TestFruitSpin(t *testing.T) {
	r := random.GetSimpleRandomGenerator()
	slot := slots.NewFruitSlot(r)

	slot.Init()

	response := slot.ProcessCommand(&command.Slot3SpinRequest{Bet:100})
	if got, expected := response.Type(), command.SLOT3_SPIN_RESPONSE; got != expected {
		t.Errorf("Wrong response time for a Slot3InitRequest. Expecting response of type %v, got %v", expected, got)
		spew.Dump(response)
	}
}


func TestFruits(t *testing.T) {
	r := random.GetSimpleRandomGenerator()
	slot := slots.NewFruitSlot(r)

	slot.Init()

	plays := 0
	wins := 0

	for i:=0; i<1000000; i++ {
		plays += 100
		response := slot.ProcessCommand(&command.Slot3SpinRequest{Bet:100})
		wins += int(response.(*command.Slot3SpinResponse).Win)
	}

	fmt.Printf("Played %d, Earned %d. Balance %d. Benefit %f %%\n", plays, wins, wins-plays, 100.0*float64(wins)/float64(plays))
}
