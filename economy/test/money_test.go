package test

import (
	"testing"
	"github.com/x1m3/elixir/economy"
)

func TestMoney_Add(t *testing.T) {

	note1 := economy.NewMoney(100, "EUR")
	note2 := economy.NewMoney(200, "EUR")

	total, err := note1.Add(note2)
	if err!=nil {
		t.Error(err)
	}
	if got, expected := total.Currency(), note1.Currency(); got!=expected {
		t.Errorf("Wrong currency after adding. Got <%v>, expecting <%v>", got, expected)
	}

	if got, expected := total.Amount(), int64(300); got!=expected {
		t.Errorf("Wrong amount after adding. Got <%v>, expecting <%v>", got, expected)
	}

	note3 := economy.NewMoney(100, "USD")
	total, err = note1.Add(note3)
	if got, expected := err, economy.Err_CoinDiffers; got!=expected {
		t.Errorf("Expecting and error of type <%s>, got <%v>", expected, got)
	}
}

func TestMoney_Copy(t *testing.T) {
	note1 := economy.NewMoney(100, "EUR")
	note2 := note1.Copy()

	if note1 == note2 {
		t.Error("note1 and note 2 should be different objects")
	}
}


