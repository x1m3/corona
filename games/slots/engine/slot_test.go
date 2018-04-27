package engine

import (
	"testing"
	"fmt"
	"github.com/x1m3/elixir/random"
	"time"
)

func TestNew3x24(t *testing.T) {

	gen := random.NewSimpleRandomGenerator()
	gen.Init(5 * time.Millisecond)
	w1 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w2 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w3 := [...]int8{1, 2, 3, 0, 4, 5, 6, 0, 7, 3, 9, 0, 1, 2, 3, 0, 4, 5, 6, 0, 7, 8, 9, 0}

	prizes := NewWinTable3Wheels()
	prizes.Add(NewWinRule3Wheels(1, 1, 1, 500))
	prizes.Add(NewWinRule3Wheels(2, 2, 2, 200))
	prizes.Add(NewWinRule3Wheels(3, 3, 3, 100))
	prizes.Add(NewWinRule3Wheels(4, 4, 4, 50))
	prizes.Add(NewWinRule3Wheels(5, 5, 5, 20))
	prizes.Add(NewWinRule3Wheels(6, 6, 6, 10))
	prizes.Add(NewWinRule3Wheels(7, 7, 7, 10))
	prizes.Add(NewWinRule3Wheels(8, 8, 8, 10))
	prizes.Add(NewWinRule3Wheels(9, 9, 9, 10))
	prizes.Add(NewWinRule3Wheels(1, 1, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(2, 2, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(3, 3, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(4, 4, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(5, 5, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(6, 6, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(7, 7, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(8, 8, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(9, 9, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(1, ANY_VALUE, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(ANY_VALUE, 1, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(ANY_VALUE, ANY_VALUE, 1, 2))

	slot := New3x24(gen, w1, w2, w3, prizes)

	plays := 0
	wins := 0

	for i := 0; i < 1000000; i++ {
		plays += 100

		win, p1, p2, p3 := slot.SpinAll(100)
		wins += int(win)

		_, _, _ = p1, p2, p3

	}
	fmt.Printf("Played %d, Earned %d. Balance %d. Benefit %f %%\n", plays, wins, wins-plays, 100.0*float64(wins)/float64(plays))
}

func TestSlot3X24_Init(t *testing.T) {

	gen := random.NewSimpleRandomGenerator()
	gen.Init(5 * time.Millisecond)
	w1 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w2 := [...]int8{1, 2, 3, 0, 1, 6, 7, 0, 9, 3, 0, 6, 0, 6, 7, 8, 9, 0, 3, 4, 5, 0, 7, 8}
	w3 := [...]int8{1, 2, 3, 0, 4, 5, 6, 0, 7, 3, 9, 0, 1, 2, 3, 0, 4, 5, 6, 0, 7, 8, 9, 0}
	prizes := NewWinTable3Wheels()
	prizes.Add(NewWinRule3Wheels(1, 1, 1, 500))
	prizes.Add(NewWinRule3Wheels(2, 2, 2, 200))
	prizes.Add(NewWinRule3Wheels(3, 3, 3, 100))
	prizes.Add(NewWinRule3Wheels(4, 4, 4, 50))
	prizes.Add(NewWinRule3Wheels(5, 5, 5, 20))
	prizes.Add(NewWinRule3Wheels(6, 6, 6, 10))
	prizes.Add(NewWinRule3Wheels(7, 7, 7, 10))
	prizes.Add(NewWinRule3Wheels(8, 8, 8, 10))
	prizes.Add(NewWinRule3Wheels(9, 9, 9, 10))
	prizes.Add(NewWinRule3Wheels(1, 1, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(2, 2, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(3, 3, ANY_VALUE, 5))
	prizes.Add(NewWinRule3Wheels(4, 4, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(5, 5, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(6, 6, ANY_VALUE, 3))
	prizes.Add(NewWinRule3Wheels(7, 7, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(8, 8, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(9, 9, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(1, 0, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(ANY_VALUE, 1, ANY_VALUE, 2))
	prizes.Add(NewWinRule3Wheels(ANY_VALUE, ANY_VALUE, 1, 2))

	slot := New3x24(gen, w1, w2, w3, prizes)
	for i := 0; i < 100000; i++ {
		slot.Init()
		v1 := slot.wheel1[slot.p1]
		v2 := slot.wheel2[slot.p2]
		v3 := slot.wheel3[slot.p3]

		if got, expected := slot.rules.WinMultiplier(v1, v2, v3), 0; got != int64(expected) {
			t.Errorf("Initial position should be a non prize condition, Current value is [%d, %d, %d] with a value of %d", v1, v2, v3, got)
		}
	}
}
